package api

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"google.golang.org/grpc"

	"github.com/cedi/meeting_epd/pkg/client"
	pb "github.com/cedi/meeting_epd/pkg/protos"
)

type GrpcApi struct {
	pb.UnimplementedCalenderServiceServer
	client *client.ICalClient
	zapLog *otelzap.Logger

	srv *grpc.Server
	lis net.Listener
}

func NewGrpcApiServer(zapLog *otelzap.Logger, client *client.ICalClient) *GrpcApi {
	e := &GrpcApi{
		zapLog: zapLog,
		client: client,
		srv:    grpc.NewServer(),
	}

	pb.RegisterCalenderServiceServer(e.srv, e)

	addr := fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.grpcPort"))

	var err error
	e.lis, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("gRPC PI: failed to listen: %v", err)
	}

	return e
}

func (e *GrpcApi) GetCalendar(ctx context.Context, _ *pb.CalendarRequest) (*pb.CalendarResponse, error) {
	return e.client.GetEvents(ctx), nil
}

func (e *GrpcApi) Serve() error {
	otelzap.L().Sugar().Infof("gRPC Server listening at %s", e.lis.Addr())
	return e.srv.Serve(e.lis)
}

func (e *GrpcApi) Addr() string {
	return e.lis.Addr().String()
}
