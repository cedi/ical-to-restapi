package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"

	"github.com/cedi/meeting_epd/pkg/api"
	"github.com/cedi/meeting_epd/pkg/client"
)

var (
	savedHostname string
	savedGrpcPort int
	savedRestPort int
)

func main() {
	viper.SetDefault("server.httpPort", 8080)
	viper.SetDefault("server.grpcPort", 50051)
	viper.SetDefault("server.host", "")
	viper.SetDefault("server.debug", false)
	viper.SetDefault("server.refresh", "5m")
	viper.SetDefault("rules", []client.Rule{client.Rule{Name: "Catch All", Key: "*", Contains: []string{"*"}, Skip: false}})

	viper.SetConfigName("display")                          // name of config file (without extension)
	viper.SetConfigType("yaml")                             // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("$HOME/.config/conference-display") // call multiple times to add many search paths
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

	// Initialize Logging
	var zapLog *zap.Logger
	if viper.GetBool("server.debug") {
		zapLog, err = zap.NewDevelopment()
		gin.SetMode(gin.DebugMode)
	} else {
		zapLog, err = zap.NewProduction()
		gin.SetMode(gin.ReleaseMode)
	}

	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %w", err))
	}

	otelZap := otelzap.New(zapLog,
		otelzap.WithCaller(true),
		otelzap.WithErrorStatusLevel(zap.ErrorLevel),
		otelzap.WithStackTrace(false),
	)

	undo := otelzap.ReplaceGlobals(otelZap)
	defer zapLog.Sync()
	defer undo()

	iCalClient := client.NewICalClient(otelZap)

	refresh, err := time.ParseDuration(viper.GetString("server.refresh"))
	if err != nil {
		refresh = 5 * time.Minute
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

	restApiServer := api.NewRestApiServer(otelZap, iCalClient)
	gRpcApiServer := api.NewGrpcApiServer(otelZap, iCalClient)

	viper.OnConfigChange(func(e fsnotify.Event) {
		otelzap.L().Sugar().Infow("Config file change detected. Reloading.", "filename", e.Name)
		iCalClient.FetchEvents(context.Background())

		if viper.GetBool("server.debug") {
			zapLog, err = zap.NewDevelopment()
			gin.SetMode(gin.DebugMode)
		} else {
			zapLog, err = zap.NewProduction()
			gin.SetMode(gin.ReleaseMode)
		}

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

	viper.WatchConfig()

	// Serve Rest-API
	go func() {
		if err := restApiServer.ListenAndServe(); err != nil {
			panic(err.Error())
		}
	}()

	// Serve gRPC-API
	go func() {
		if err := gRpcApiServer.Serve(); err != nil {
			panic(err.Error())
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// close timer
	close(quitRefreshTicker)
}
