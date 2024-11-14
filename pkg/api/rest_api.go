package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	ginprometheus "github.com/mcuadros/go-gin-prometheus"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/cedi/icaltest/pkg/client"
)

type RestApi struct {
	client *client.ICalClient
	zapLog *otelzap.Logger
	srv    *http.Server
}

func NewRestApiServer(zapLog *otelzap.Logger, client *client.ICalClient) *RestApi {
	e := &RestApi{
		zapLog: zapLog,
		client: client,
	}

	// Setup Gin router
	router := gin.New()
	router.Use(
		otelgin.Middleware("conf_room_display"),
	)

	// Set-up Prometheus
	p := ginprometheus.NewPrometheus("conf_room_display")
	p.Use(router)

	router.GET("/calendar", e.GetCalendar)

	// Set up the HTTP Server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.httpPort")),
		Handler: router,
	}

	e.srv = srv

	return e
}

func (e *RestApi) ListenAndServe() error {
	return e.srv.ListenAndServe()
}

func (e *RestApi) GetCalendar(ct *gin.Context) {
	ct.JSON(http.StatusOK, e.client.GetEvents(ct.Request.Context()))
}
