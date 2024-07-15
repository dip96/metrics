// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             v3.12.4
// source: protos/metric/v2/metric_service.proto

package v2

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	MetricService_AddMetricV2_FullMethodName = "/metrics.v2.MetricService/AddMetricV2"
	MetricService_GetMetricV2_FullMethodName = "/metrics.v2.MetricService/GetMetricV2"
)

// MetricServiceClient is the client API for MetricService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetricServiceClient interface {
	AddMetricV2(ctx context.Context, in *AddMetricV2Request, opts ...grpc.CallOption) (*AddMetricV2Response, error)
	GetMetricV2(ctx context.Context, in *AddMetricV2Request, opts ...grpc.CallOption) (*AddMetricV2Response, error)
}

type metricServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricServiceClient(cc grpc.ClientConnInterface) MetricServiceClient {
	return &metricServiceClient{cc}
}

func (c *metricServiceClient) AddMetricV2(ctx context.Context, in *AddMetricV2Request, opts ...grpc.CallOption) (*AddMetricV2Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AddMetricV2Response)
	err := c.cc.Invoke(ctx, MetricService_AddMetricV2_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricServiceClient) GetMetricV2(ctx context.Context, in *AddMetricV2Request, opts ...grpc.CallOption) (*AddMetricV2Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AddMetricV2Response)
	err := c.cc.Invoke(ctx, MetricService_GetMetricV2_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetricServiceServer is the server API for MetricService service.
// All implementations must embed UnimplementedMetricServiceServer
// for forward compatibility
type MetricServiceServer interface {
	AddMetricV2(context.Context, *AddMetricV2Request) (*AddMetricV2Response, error)
	GetMetricV2(context.Context, *AddMetricV2Request) (*AddMetricV2Response, error)
	mustEmbedUnimplementedMetricServiceServer()
}

// UnimplementedMetricServiceServer must be embedded to have forward compatible implementations.
type UnimplementedMetricServiceServer struct {
}

func (UnimplementedMetricServiceServer) AddMetricV2(context.Context, *AddMetricV2Request) (*AddMetricV2Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddMetricV2 not implemented")
}
func (UnimplementedMetricServiceServer) GetMetricV2(context.Context, *AddMetricV2Request) (*AddMetricV2Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMetricV2 not implemented")
}
func (UnimplementedMetricServiceServer) mustEmbedUnimplementedMetricServiceServer() {}

// UnsafeMetricServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetricServiceServer will
// result in compilation errors.
type UnsafeMetricServiceServer interface {
	mustEmbedUnimplementedMetricServiceServer()
}

func RegisterMetricServiceServer(s grpc.ServiceRegistrar, srv MetricServiceServer) {
	s.RegisterService(&MetricService_ServiceDesc, srv)
}

func _MetricService_AddMetricV2_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddMetricV2Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServiceServer).AddMetricV2(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricService_AddMetricV2_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServiceServer).AddMetricV2(ctx, req.(*AddMetricV2Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricService_GetMetricV2_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddMetricV2Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServiceServer).GetMetricV2(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricService_GetMetricV2_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServiceServer).GetMetricV2(ctx, req.(*AddMetricV2Request))
	}
	return interceptor(ctx, in, info, handler)
}

// MetricService_ServiceDesc is the grpc.ServiceDesc for MetricService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MetricService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "metrics.v2.MetricService",
	HandlerType: (*MetricServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddMetricV2",
			Handler:    _MetricService_AddMetricV2_Handler,
		},
		{
			MethodName: "GetMetricV2",
			Handler:    _MetricService_GetMetricV2_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protos/metric/v2/metric_service.proto",
}