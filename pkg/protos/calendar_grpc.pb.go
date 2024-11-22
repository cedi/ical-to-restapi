// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.3
// source: calendar.proto

package protos

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	CalenderService_GetCalendar_FullMethodName     = "/meetingroom_display_epd.CalenderService/GetCalendar"
	CalenderService_RefreshCalendar_FullMethodName = "/meetingroom_display_epd.CalenderService/RefreshCalendar"
	CalenderService_GetCustomStatus_FullMethodName = "/meetingroom_display_epd.CalenderService/GetCustomStatus"
	CalenderService_SetCustomStatus_FullMethodName = "/meetingroom_display_epd.CalenderService/SetCustomStatus"
)

// CalenderServiceClient is the client API for CalenderService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CalenderServiceClient interface {
	GetCalendar(ctx context.Context, in *CalendarRequest, opts ...grpc.CallOption) (*CalendarResponse, error)
	RefreshCalendar(ctx context.Context, in *CalendarRequest, opts ...grpc.CallOption) (*RefreshCalendarResponse, error)
	GetCustomStatus(ctx context.Context, in *CustomStatusRequest, opts ...grpc.CallOption) (*CustomStatus, error)
	SetCustomStatus(ctx context.Context, in *CustomStatus, opts ...grpc.CallOption) (*CustomStatus, error)
}

type calenderServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCalenderServiceClient(cc grpc.ClientConnInterface) CalenderServiceClient {
	return &calenderServiceClient{cc}
}

func (c *calenderServiceClient) GetCalendar(ctx context.Context, in *CalendarRequest, opts ...grpc.CallOption) (*CalendarResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CalendarResponse)
	err := c.cc.Invoke(ctx, CalenderService_GetCalendar_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *calenderServiceClient) RefreshCalendar(ctx context.Context, in *CalendarRequest, opts ...grpc.CallOption) (*RefreshCalendarResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RefreshCalendarResponse)
	err := c.cc.Invoke(ctx, CalenderService_RefreshCalendar_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *calenderServiceClient) GetCustomStatus(ctx context.Context, in *CustomStatusRequest, opts ...grpc.CallOption) (*CustomStatus, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CustomStatus)
	err := c.cc.Invoke(ctx, CalenderService_GetCustomStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *calenderServiceClient) SetCustomStatus(ctx context.Context, in *CustomStatus, opts ...grpc.CallOption) (*CustomStatus, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CustomStatus)
	err := c.cc.Invoke(ctx, CalenderService_SetCustomStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CalenderServiceServer is the server API for CalenderService service.
// All implementations must embed UnimplementedCalenderServiceServer
// for forward compatibility.
type CalenderServiceServer interface {
	GetCalendar(context.Context, *CalendarRequest) (*CalendarResponse, error)
	RefreshCalendar(context.Context, *CalendarRequest) (*RefreshCalendarResponse, error)
	GetCustomStatus(context.Context, *CustomStatusRequest) (*CustomStatus, error)
	SetCustomStatus(context.Context, *CustomStatus) (*CustomStatus, error)
	mustEmbedUnimplementedCalenderServiceServer()
}

// UnimplementedCalenderServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedCalenderServiceServer struct{}

func (UnimplementedCalenderServiceServer) GetCalendar(context.Context, *CalendarRequest) (*CalendarResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCalendar not implemented")
}
func (UnimplementedCalenderServiceServer) RefreshCalendar(context.Context, *CalendarRequest) (*RefreshCalendarResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshCalendar not implemented")
}
func (UnimplementedCalenderServiceServer) GetCustomStatus(context.Context, *CustomStatusRequest) (*CustomStatus, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCustomStatus not implemented")
}
func (UnimplementedCalenderServiceServer) SetCustomStatus(context.Context, *CustomStatus) (*CustomStatus, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetCustomStatus not implemented")
}
func (UnimplementedCalenderServiceServer) mustEmbedUnimplementedCalenderServiceServer() {}
func (UnimplementedCalenderServiceServer) testEmbeddedByValue()                         {}

// UnsafeCalenderServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CalenderServiceServer will
// result in compilation errors.
type UnsafeCalenderServiceServer interface {
	mustEmbedUnimplementedCalenderServiceServer()
}

func RegisterCalenderServiceServer(s grpc.ServiceRegistrar, srv CalenderServiceServer) {
	// If the following call pancis, it indicates UnimplementedCalenderServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&CalenderService_ServiceDesc, srv)
}

func _CalenderService_GetCalendar_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CalendarRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CalenderServiceServer).GetCalendar(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CalenderService_GetCalendar_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CalenderServiceServer).GetCalendar(ctx, req.(*CalendarRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CalenderService_RefreshCalendar_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CalendarRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CalenderServiceServer).RefreshCalendar(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CalenderService_RefreshCalendar_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CalenderServiceServer).RefreshCalendar(ctx, req.(*CalendarRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CalenderService_GetCustomStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CustomStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CalenderServiceServer).GetCustomStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CalenderService_GetCustomStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CalenderServiceServer).GetCustomStatus(ctx, req.(*CustomStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CalenderService_SetCustomStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CustomStatus)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CalenderServiceServer).SetCustomStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CalenderService_SetCustomStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CalenderServiceServer).SetCustomStatus(ctx, req.(*CustomStatus))
	}
	return interceptor(ctx, in, info, handler)
}

// CalenderService_ServiceDesc is the grpc.ServiceDesc for CalenderService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CalenderService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "meetingroom_display_epd.CalenderService",
	HandlerType: (*CalenderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetCalendar",
			Handler:    _CalenderService_GetCalendar_Handler,
		},
		{
			MethodName: "RefreshCalendar",
			Handler:    _CalenderService_RefreshCalendar_Handler,
		},
		{
			MethodName: "GetCustomStatus",
			Handler:    _CalenderService_GetCustomStatus_Handler,
		},
		{
			MethodName: "SetCustomStatus",
			Handler:    _CalenderService_SetCustomStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "calendar.proto",
}
