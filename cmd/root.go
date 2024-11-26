package cmd

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

var (
	// Version represents the Version of the kkpctl binary, should be set via ldflags -X
	Version string

	// Date represents the Date of when the kkpctl binary was build, should be set via ldflags -X
	Date string

	// Commit represents the Commit-hash from which kkpctl binary was build, should be set via ldflags -X
	Commit string

	// BuiltBy represents who build the binary, should be set via ldflags -X
	BuiltBy string
)

func initTelemetry() (func(), *zap.Logger, *otelzap.Logger) {
	var err error

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

	return undo, zapLog, otelZap
}

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "enable debug logging")
	viper.BindPFlag("server.debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "meetingepd",
	Short: "A CLI for interacting with the meetingroom epd dipslay server.",
	Long:  `This is a CLI for interacting with the meetingroom epd display server`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version, commit, date, builtBy string) {

	// asign build flags for version info
	Version = version
	Date = date
	Commit = commit
	BuiltBy = builtBy

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
