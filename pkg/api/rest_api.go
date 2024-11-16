package api

import (
	"fmt"
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/mcuadros/go-gin-prometheus"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/cedi/meeting_epd/pkg/client"
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
	router := gin.New(func(e *gin.Engine) {})

	// Setup otelgin to expose Open Telemetry
	router.Use(otelgin.Middleware("conf_room_display"))

	// Setup ginzap to log everything correctly to zap
	router.Use(ginzap.GinzapWithConfig(zapLog, &ginzap.Config{
		UTC:        true,
		TimeFormat: time.RFC3339,
		Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
			fields := []zapcore.Field{}
			// log request ID
			if requestID := c.Writer.Header().Get("X-Request-Id"); requestID != "" {
				fields = append(fields, zap.String("request_id", requestID))
			}

			// log trace and span ID
			if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
				fields = append(fields, zap.String("trace_id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()))
				fields = append(fields, zap.String("span_id", trace.SpanFromContext(c.Request.Context()).SpanContext().SpanID().String()))
			}
			return fields
		}),
	}))

	// Set-up Prometheus to expose prometheus metrics
	p := ginprometheus.NewPrometheus("conf_room_display")
	p.Use(router)

	router.GET("/calendar", e.GetCalendar)

	// configure the HTTP Server
	e.srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.httpPort")),
		Handler: router,
	}

	return e
}

func (e *RestApi) ListenAndServe() error {
	return e.srv.ListenAndServe()
}

func (e *RestApi) GetCalendar(ct *gin.Context) {
	switch ct.ContentType() {
	case "application/protobuf":
		ct.ProtoBuf(http.StatusOK, e.client.GetEvents(ct.Request.Context()))
	default:
		ct.JSON(http.StatusOK, e.client.GetEvents(ct.Request.Context()))
	}
}

func (e *RestApi) Addr() string {
	return e.srv.Addr
}
