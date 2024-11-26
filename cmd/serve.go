package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cedi/meeting_epd/pkg/api"
	"github.com/cedi/meeting_epd/pkg/client"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

var (
	savedHostname          string
	savedGrpcPort          int
	savedRestPort          int
	defaultCalendarRefresh time.Duration = 30 * time.Minute
)

func initCalendarRefresh(zapLog *otelzap.Logger, iCalClient *client.ICalClient) chan struct{} {
	refreshConfig := viper.GetString("server.refresh")
	refresh, err := time.ParseDuration(refreshConfig)
	if err != nil {
		zapLog.Sugar().Errorf("Failed to parse '%s' as time.Duration: %v. Failing back to default refresh duration (%s)",
			refreshConfig, err.Error(),
			defaultCalendarRefresh,
		)
		refresh = defaultCalendarRefresh
	}

	refreshTicker := time.NewTicker(refresh)
	quitRefreshTicker := make(chan struct{})
	go func() {
		// initial load
		iCalClient.FetchEvents(context.Background())

		for {
			select {
			case <-refreshTicker.C:
				iCalClient.FetchEvents(context.Background())
			case <-quitRefreshTicker:
				refreshTicker.Stop()
				return
			}
		}
	}()

	return quitRefreshTicker
}

func viperConfigChange(undo func(), zapLog *zap.Logger, otelZap *otelzap.Logger, iCalClient *client.ICalClient, quitRefreshTicker *chan struct{}) {
	viper.OnConfigChange(func(e fsnotify.Event) {
		otelzap.L().Sugar().Infow("Config file change detected. Reloading.", "filename", e.Name)
		iCalClient.FetchEvents(context.Background())

		// refresh logger
		zapLog.Sync()
		undo()
		undo, zapLog, otelZap = initTelemetry()

		// Refresh calendar watch timer
		close(*quitRefreshTicker)
		*quitRefreshTicker = initCalendarRefresh(otelZap, iCalClient)

		if savedHostname != viper.GetString("server.host") ||
			savedGrpcPort != viper.GetInt("server.grpcPort") ||
			savedRestPort != viper.GetInt("server.httpPort") {
			zapLog.Sugar().Errorw("Unable to change host or port at runtime!",
				"new_host", viper.GetString("server.host"),
				"old_host", savedHostname,
				"new_grpcPort", viper.GetInt("server.grpcPort"),
				"old_grpcPort", savedGrpcPort,
				"new_restPort", viper.GetInt("server.httpPort"),
				"old_grpcPort", savedRestPort,
			)
		}
	})
}

var serveCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Shows version information",
	Example: "meetingepd version",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("server.httpPort", 8099)
		viper.SetDefault("server.grpcPort", 50051)
		viper.SetDefault("server.host", "")
		viper.SetDefault("server.debug", false)
		viper.SetDefault("server.refresh", "5m")
		viper.SetDefault("rules", []client.Rule{{Name: "Catch All", Key: "*", Contains: []string{"*"}, Skip: false}})

		viper.SetConfigName("options")                          // name of config file (without extension)
		viper.AddConfigPath("$HOME/.config/conference-display") // call multiple times to add many search paths
		viper.AddConfigPath("/data")                            // optionally look for config in the working directory
		viper.AddConfigPath(".")                                // optionally look for config in the working directory

		viper.SetEnvPrefix("DISPLAY")
		viper.AutomaticEnv()

		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {             // Handle errors reading the config file
			panic(fmt.Errorf("fatal error config file: %w", err))
		}

		savedHostname = viper.GetString("server.host")
		savedGrpcPort = viper.GetInt("server.grpcPort")
		savedRestPort = viper.GetInt("server.httpPort")

		undo, zapLog, otelZap := initTelemetry()
		defer zapLog.Sync()
		defer undo()

		iCalClient := client.NewICalClient(otelZap)

		quitRefreshTicker := initCalendarRefresh(otelZap, iCalClient)
		viperConfigChange(undo, zapLog, otelZap, iCalClient, &quitRefreshTicker)
		viper.WatchConfig()

		// Serve Rest-API
		go func() {
			restApiServer := api.NewRestApiServer(otelZap, iCalClient)
			if err := restApiServer.ListenAndServe(); err != nil {
				panic(err.Error())
			}
		}()

		// Serve gRPC-API
		go func() {
			gRpcApiServer := api.NewGrpcApiServer(otelZap, iCalClient)
			if err := gRpcApiServer.Serve(); err != nil {
				panic(err.Error())
			}
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		// close timer
		close(quitRefreshTicker)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
