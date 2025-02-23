// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: kubeconfigstore/v1/kubeconfig_store.proto

package kubeconfigstorev1

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
	KubeconfigStoreService_GetID_FullMethodName                 = "/kubeconfigstore.v1.KubeconfigStoreService/GetID"
	KubeconfigStoreService_GetContextPrefix_FullMethodName      = "/kubeconfigstore.v1.KubeconfigStoreService/GetContextPrefix"
	KubeconfigStoreService_VerifyKubeconfigPaths_FullMethodName = "/kubeconfigstore.v1.KubeconfigStoreService/VerifyKubeconfigPaths"
	KubeconfigStoreService_StartSearch_FullMethodName           = "/kubeconfigstore.v1.KubeconfigStoreService/StartSearch"
	KubeconfigStoreService_GetKubeconfigForPath_FullMethodName  = "/kubeconfigstore.v1.KubeconfigStoreService/GetKubeconfigForPath"
)

// KubeconfigStoreServiceClient is the client API for KubeconfigStoreService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KubeconfigStoreServiceClient interface {
	// Retrieves the unique store ID
	GetID(ctx context.Context, in *GetIDRequest, opts ...grpc.CallOption) (*GetIDResponse, error)
	// Retrieves the prefix for the kubeconfig context names
	GetContextPrefix(ctx context.Context, in *GetContextPrefixRequest, opts ...grpc.CallOption) (*GetContextPrefixResponse, error)
	// Verifies the kubeconfig search paths
	VerifyKubeconfigPaths(ctx context.Context, in *VerifyKubeconfigPathsRequest, opts ...grpc.CallOption) (*VerifyKubeconfigPathsResponse, error)
	// Starts the search and streams results
	StartSearch(ctx context.Context, in *StartSearchRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StartSearchResponse], error)
	// Retrieves kubeconfig bytes for a given path and optional tags
	GetKubeconfigForPath(ctx context.Context, in *GetKubeconfigForPathRequest, opts ...grpc.CallOption) (*GetKubeconfigForPathResponse, error)
}

type kubeconfigStoreServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewKubeconfigStoreServiceClient(cc grpc.ClientConnInterface) KubeconfigStoreServiceClient {
	return &kubeconfigStoreServiceClient{cc}
}

func (c *kubeconfigStoreServiceClient) GetID(ctx context.Context, in *GetIDRequest, opts ...grpc.CallOption) (*GetIDResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetIDResponse)
	err := c.cc.Invoke(ctx, KubeconfigStoreService_GetID_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kubeconfigStoreServiceClient) GetContextPrefix(ctx context.Context, in *GetContextPrefixRequest, opts ...grpc.CallOption) (*GetContextPrefixResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetContextPrefixResponse)
	err := c.cc.Invoke(ctx, KubeconfigStoreService_GetContextPrefix_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kubeconfigStoreServiceClient) VerifyKubeconfigPaths(ctx context.Context, in *VerifyKubeconfigPathsRequest, opts ...grpc.CallOption) (*VerifyKubeconfigPathsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(VerifyKubeconfigPathsResponse)
	err := c.cc.Invoke(ctx, KubeconfigStoreService_VerifyKubeconfigPaths_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kubeconfigStoreServiceClient) StartSearch(ctx context.Context, in *StartSearchRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StartSearchResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &KubeconfigStoreService_ServiceDesc.Streams[0], KubeconfigStoreService_StartSearch_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StartSearchRequest, StartSearchResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type KubeconfigStoreService_StartSearchClient = grpc.ServerStreamingClient[StartSearchResponse]

func (c *kubeconfigStoreServiceClient) GetKubeconfigForPath(ctx context.Context, in *GetKubeconfigForPathRequest, opts ...grpc.CallOption) (*GetKubeconfigForPathResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetKubeconfigForPathResponse)
	err := c.cc.Invoke(ctx, KubeconfigStoreService_GetKubeconfigForPath_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KubeconfigStoreServiceServer is the server API for KubeconfigStoreService service.
// All implementations should embed UnimplementedKubeconfigStoreServiceServer
// for forward compatibility.
type KubeconfigStoreServiceServer interface {
	// Retrieves the unique store ID
	GetID(context.Context, *GetIDRequest) (*GetIDResponse, error)
	// Retrieves the prefix for the kubeconfig context names
	GetContextPrefix(context.Context, *GetContextPrefixRequest) (*GetContextPrefixResponse, error)
	// Verifies the kubeconfig search paths
	VerifyKubeconfigPaths(context.Context, *VerifyKubeconfigPathsRequest) (*VerifyKubeconfigPathsResponse, error)
	// Starts the search and streams results
	StartSearch(*StartSearchRequest, grpc.ServerStreamingServer[StartSearchResponse]) error
	// Retrieves kubeconfig bytes for a given path and optional tags
	GetKubeconfigForPath(context.Context, *GetKubeconfigForPathRequest) (*GetKubeconfigForPathResponse, error)
}

// UnimplementedKubeconfigStoreServiceServer should be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedKubeconfigStoreServiceServer struct{}

func (UnimplementedKubeconfigStoreServiceServer) GetID(context.Context, *GetIDRequest) (*GetIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetID not implemented")
}
func (UnimplementedKubeconfigStoreServiceServer) GetContextPrefix(context.Context, *GetContextPrefixRequest) (*GetContextPrefixResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetContextPrefix not implemented")
}
func (UnimplementedKubeconfigStoreServiceServer) VerifyKubeconfigPaths(context.Context, *VerifyKubeconfigPathsRequest) (*VerifyKubeconfigPathsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyKubeconfigPaths not implemented")
}
func (UnimplementedKubeconfigStoreServiceServer) StartSearch(*StartSearchRequest, grpc.ServerStreamingServer[StartSearchResponse]) error {
	return status.Errorf(codes.Unimplemented, "method StartSearch not implemented")
}
func (UnimplementedKubeconfigStoreServiceServer) GetKubeconfigForPath(context.Context, *GetKubeconfigForPathRequest) (*GetKubeconfigForPathResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetKubeconfigForPath not implemented")
}
func (UnimplementedKubeconfigStoreServiceServer) testEmbeddedByValue() {}

// UnsafeKubeconfigStoreServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KubeconfigStoreServiceServer will
// result in compilation errors.
type UnsafeKubeconfigStoreServiceServer interface {
	mustEmbedUnimplementedKubeconfigStoreServiceServer()
}

func RegisterKubeconfigStoreServiceServer(s grpc.ServiceRegistrar, srv KubeconfigStoreServiceServer) {
	// If the following call pancis, it indicates UnimplementedKubeconfigStoreServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&KubeconfigStoreService_ServiceDesc, srv)
}

func _KubeconfigStoreService_GetID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KubeconfigStoreServiceServer).GetID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KubeconfigStoreService_GetID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KubeconfigStoreServiceServer).GetID(ctx, req.(*GetIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KubeconfigStoreService_GetContextPrefix_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetContextPrefixRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KubeconfigStoreServiceServer).GetContextPrefix(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KubeconfigStoreService_GetContextPrefix_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KubeconfigStoreServiceServer).GetContextPrefix(ctx, req.(*GetContextPrefixRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KubeconfigStoreService_VerifyKubeconfigPaths_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyKubeconfigPathsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KubeconfigStoreServiceServer).VerifyKubeconfigPaths(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KubeconfigStoreService_VerifyKubeconfigPaths_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KubeconfigStoreServiceServer).VerifyKubeconfigPaths(ctx, req.(*VerifyKubeconfigPathsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KubeconfigStoreService_StartSearch_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StartSearchRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(KubeconfigStoreServiceServer).StartSearch(m, &grpc.GenericServerStream[StartSearchRequest, StartSearchResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type KubeconfigStoreService_StartSearchServer = grpc.ServerStreamingServer[StartSearchResponse]

func _KubeconfigStoreService_GetKubeconfigForPath_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetKubeconfigForPathRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KubeconfigStoreServiceServer).GetKubeconfigForPath(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KubeconfigStoreService_GetKubeconfigForPath_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KubeconfigStoreServiceServer).GetKubeconfigForPath(ctx, req.(*GetKubeconfigForPathRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// KubeconfigStoreService_ServiceDesc is the grpc.ServiceDesc for KubeconfigStoreService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var KubeconfigStoreService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kubeconfigstore.v1.KubeconfigStoreService",
	HandlerType: (*KubeconfigStoreServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetID",
			Handler:    _KubeconfigStoreService_GetID_Handler,
		},
		{
			MethodName: "GetContextPrefix",
			Handler:    _KubeconfigStoreService_GetContextPrefix_Handler,
		},
		{
			MethodName: "VerifyKubeconfigPaths",
			Handler:    _KubeconfigStoreService_VerifyKubeconfigPaths_Handler,
		},
		{
			MethodName: "GetKubeconfigForPath",
			Handler:    _KubeconfigStoreService_GetKubeconfigForPath_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StartSearch",
			Handler:       _KubeconfigStoreService_StartSearch_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "kubeconfigstore/v1/kubeconfig_store.proto",
}
