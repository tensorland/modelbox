// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.1
// source: admin.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ModelBoxAdminClient is the client API for ModelBoxAdmin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ModelBoxAdminClient interface {
	// Register an agent capable of running plugins
	RegisterAgent(ctx context.Context, in *RegisterAgentRequest, opts ...grpc.CallOption) (*RegisterAgentResponse, error)
	// Workers heartbeat with the server about their presence
	// and work progress periodically
	Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
	// Download the list of work that can be exectuted by a action runner
	GetRunnableActionInstances(ctx context.Context, in *GetRunnableActionInstancesRequest, opts ...grpc.CallOption) (*GetRunnableActionInstancesResponse, error)
	// Update action status
	UpdateActionStatus(ctx context.Context, in *UpdateActionStatusRequest, opts ...grpc.CallOption) (*UpdateActionStatusResponse, error)
}

type modelBoxAdminClient struct {
	cc grpc.ClientConnInterface
}

func NewModelBoxAdminClient(cc grpc.ClientConnInterface) ModelBoxAdminClient {
	return &modelBoxAdminClient{cc}
}

func (c *modelBoxAdminClient) RegisterAgent(ctx context.Context, in *RegisterAgentRequest, opts ...grpc.CallOption) (*RegisterAgentResponse, error) {
	out := new(RegisterAgentResponse)
	err := c.cc.Invoke(ctx, "/modelbox.ModelBoxAdmin/RegisterAgent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelBoxAdminClient) Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	out := new(HeartbeatResponse)
	err := c.cc.Invoke(ctx, "/modelbox.ModelBoxAdmin/Heartbeat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelBoxAdminClient) GetRunnableActionInstances(ctx context.Context, in *GetRunnableActionInstancesRequest, opts ...grpc.CallOption) (*GetRunnableActionInstancesResponse, error) {
	out := new(GetRunnableActionInstancesResponse)
	err := c.cc.Invoke(ctx, "/modelbox.ModelBoxAdmin/GetRunnableActionInstances", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelBoxAdminClient) UpdateActionStatus(ctx context.Context, in *UpdateActionStatusRequest, opts ...grpc.CallOption) (*UpdateActionStatusResponse, error) {
	out := new(UpdateActionStatusResponse)
	err := c.cc.Invoke(ctx, "/modelbox.ModelBoxAdmin/UpdateActionStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ModelBoxAdminServer is the server API for ModelBoxAdmin service.
// All implementations must embed UnimplementedModelBoxAdminServer
// for forward compatibility
type ModelBoxAdminServer interface {
	// Register an agent capable of running plugins
	RegisterAgent(context.Context, *RegisterAgentRequest) (*RegisterAgentResponse, error)
	// Workers heartbeat with the server about their presence
	// and work progress periodically
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	// Download the list of work that can be exectuted by a action runner
	GetRunnableActionInstances(context.Context, *GetRunnableActionInstancesRequest) (*GetRunnableActionInstancesResponse, error)
	// Update action status
	UpdateActionStatus(context.Context, *UpdateActionStatusRequest) (*UpdateActionStatusResponse, error)
	mustEmbedUnimplementedModelBoxAdminServer()
}

// UnimplementedModelBoxAdminServer must be embedded to have forward compatible implementations.
type UnimplementedModelBoxAdminServer struct {
}

func (UnimplementedModelBoxAdminServer) RegisterAgent(context.Context, *RegisterAgentRequest) (*RegisterAgentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterAgent not implemented")
}
func (UnimplementedModelBoxAdminServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Heartbeat not implemented")
}
func (UnimplementedModelBoxAdminServer) GetRunnableActionInstances(context.Context, *GetRunnableActionInstancesRequest) (*GetRunnableActionInstancesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRunnableActionInstances not implemented")
}
func (UnimplementedModelBoxAdminServer) UpdateActionStatus(context.Context, *UpdateActionStatusRequest) (*UpdateActionStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateActionStatus not implemented")
}
func (UnimplementedModelBoxAdminServer) mustEmbedUnimplementedModelBoxAdminServer() {}

// UnsafeModelBoxAdminServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ModelBoxAdminServer will
// result in compilation errors.
type UnsafeModelBoxAdminServer interface {
	mustEmbedUnimplementedModelBoxAdminServer()
}

func RegisterModelBoxAdminServer(s grpc.ServiceRegistrar, srv ModelBoxAdminServer) {
	s.RegisterService(&ModelBoxAdmin_ServiceDesc, srv)
}

func _ModelBoxAdmin_RegisterAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterAgentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelBoxAdminServer).RegisterAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/modelbox.ModelBoxAdmin/RegisterAgent",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelBoxAdminServer).RegisterAgent(ctx, req.(*RegisterAgentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelBoxAdmin_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelBoxAdminServer).Heartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/modelbox.ModelBoxAdmin/Heartbeat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelBoxAdminServer).Heartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelBoxAdmin_GetRunnableActionInstances_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRunnableActionInstancesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelBoxAdminServer).GetRunnableActionInstances(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/modelbox.ModelBoxAdmin/GetRunnableActionInstances",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelBoxAdminServer).GetRunnableActionInstances(ctx, req.(*GetRunnableActionInstancesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelBoxAdmin_UpdateActionStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateActionStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelBoxAdminServer).UpdateActionStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/modelbox.ModelBoxAdmin/UpdateActionStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelBoxAdminServer).UpdateActionStatus(ctx, req.(*UpdateActionStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ModelBoxAdmin_ServiceDesc is the grpc.ServiceDesc for ModelBoxAdmin service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ModelBoxAdmin_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "modelbox.ModelBoxAdmin",
	HandlerType: (*ModelBoxAdminServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterAgent",
			Handler:    _ModelBoxAdmin_RegisterAgent_Handler,
		},
		{
			MethodName: "Heartbeat",
			Handler:    _ModelBoxAdmin_Heartbeat_Handler,
		},
		{
			MethodName: "GetRunnableActionInstances",
			Handler:    _ModelBoxAdmin_GetRunnableActionInstances_Handler,
		},
		{
			MethodName: "UpdateActionStatus",
			Handler:    _ModelBoxAdmin_UpdateActionStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "admin.proto",
}
