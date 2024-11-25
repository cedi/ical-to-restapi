package api

import (
	"context"
	"fmt"
	"net"

	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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
		zapLog.Fatal(fmt.Sprintf("gRPC API: failed to listen: %v", err))
	}

	return e
}

func NewGrpcApiClient(zapLog *otelzap.Logger, addr string) (*grpc.ClientConn, pb.CalenderServiceClient) {

	// Set up a connection to the server.
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zapLog.Fatal(fmt.Sprintf("gRPC API: failed to connect: %v", err))
	}

	c := pb.NewCalenderServiceClient(conn)

	return conn, c
}

func (e *GrpcApi) GetCalendar(ctx context.Context, _ *pb.CalendarRequest) (*pb.CalendarResponse, error) {
	return e.client.GetEvents(ctx), nil
}

func (e *GrpcApi) RefreshCalendar(ctx context.Context, _ *pb.CalendarRequest) (*pb.RefreshCalendarResponse, error) {
	e.client.FetchEvents(ctx)
	return nil, nil
}

func (e *GrpcApi) GetCustomStatus(ctx context.Context, _ *pb.CustomStatusRequest) (*pb.CustomStatus, error) {
	return e.client.GetCustomStatus(ctx), nil
}

func (e *GrpcApi) SetCustomStatus(ctx context.Context, status *pb.CustomStatus) (*pb.CustomStatus, error) {
	e.client.SetCustomStatus(ctx, status)
	return e.client.GetCustomStatus(ctx), nil
}

func (e *GrpcApi) Serve() error {
	otelzap.L().Sugar().Infof("gRPC Server listening at %s", e.lis.Addr())
	return e.srv.Serve(e.lis)
}

func (e *GrpcApi) Addr() string {
	return e.lis.Addr().String()
}
