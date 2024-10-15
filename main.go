package main

import (
	"fmt"
	"net/http"

	"github.com/cedi/icaltest/pkg/calendar"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/mcuadros/go-gin-prometheus"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

func main() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "")
	viper.SetDefault("server.debug", false)

	viper.SetConfigName("display")                          // name of config file (without extension)
	viper.SetConfigType("yaml")                             // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("$HOME/.config/conference-display") // call multiple times to add many search paths
	viper.AddConfigPath(".")                                // optionally look for config in the working directory

	viper.SetEnvPrefix("DISPLAY")
	viper.AutomaticEnv()

	err := viper.ReadInConfig() // Find and read the config file

	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

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

	// Setup Gin router
	router := gin.New()
	router.Use(
		otelgin.Middleware("conf_room_display"),
	)

	// Set-up Prometheus
	p := ginprometheus.NewPrometheus("conf_room_display")
	p.Use(router)

	// Set up the HTTP Server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port")),
		Handler: router,
	}

	var eventList calendar.EventList

	// Register the routes
	router.GET("/events", eventList.GetEvents)

	viper.OnConfigChange(func(e fsnotify.Event) {
		otelzap.L().Sugar().Infow("config file change detected. Reloading.", "filename", e.Name)
		eventList.CacheInvalidate()
	})
	viper.WatchConfig()

	// Serve
	if err := srv.ListenAndServe(); err != nil {
		panic(err.Error())
	}
}
