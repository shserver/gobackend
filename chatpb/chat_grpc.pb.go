// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: chatpb/chat.proto

package chatpb

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

// ChatServiceClient is the client API for ChatService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChatServiceClient interface {
	SignUp(ctx context.Context, in *RequestSignUp, opts ...grpc.CallOption) (*ResponseSignUp, error)
	SignIn(ctx context.Context, in *RequestSignIn, opts ...grpc.CallOption) (*ResponseSignIn, error)
	Chat(ctx context.Context, opts ...grpc.CallOption) (ChatService_ChatClient, error)
}

type chatServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewChatServiceClient(cc grpc.ClientConnInterface) ChatServiceClient {
	return &chatServiceClient{cc}
}

func (c *chatServiceClient) SignUp(ctx context.Context, in *RequestSignUp, opts ...grpc.CallOption) (*ResponseSignUp, error) {
	out := new(ResponseSignUp)
	err := c.cc.Invoke(ctx, "/chat.ChatService/SignUp", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatServiceClient) SignIn(ctx context.Context, in *RequestSignIn, opts ...grpc.CallOption) (*ResponseSignIn, error) {
	out := new(ResponseSignIn)
	err := c.cc.Invoke(ctx, "/chat.ChatService/SignIn", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatServiceClient) Chat(ctx context.Context, opts ...grpc.CallOption) (ChatService_ChatClient, error) {
	stream, err := c.cc.NewStream(ctx, &ChatService_ServiceDesc.Streams[0], "/chat.ChatService/Chat", opts...)
	if err != nil {
		return nil, err
	}
	x := &chatServiceChatClient{stream}
	return x, nil
}

type ChatService_ChatClient interface {
	Send(*RequestMessage) error
	Recv() (*ResponseMessage, error)
	grpc.ClientStream
}

type chatServiceChatClient struct {
	grpc.ClientStream
}

func (x *chatServiceChatClient) Send(m *RequestMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *chatServiceChatClient) Recv() (*ResponseMessage, error) {
	m := new(ResponseMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChatServiceServer is the server API for ChatService service.
// All implementations should embed UnimplementedChatServiceServer
// for forward compatibility
type ChatServiceServer interface {
	SignUp(context.Context, *RequestSignUp) (*ResponseSignUp, error)
	SignIn(context.Context, *RequestSignIn) (*ResponseSignIn, error)
	Chat(ChatService_ChatServer) error
}

// UnimplementedChatServiceServer should be embedded to have forward compatible implementations.
type UnimplementedChatServiceServer struct {
}

func (UnimplementedChatServiceServer) SignUp(context.Context, *RequestSignUp) (*ResponseSignUp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignUp not implemented")
}
func (UnimplementedChatServiceServer) SignIn(context.Context, *RequestSignIn) (*ResponseSignIn, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignIn not implemented")
}
func (UnimplementedChatServiceServer) Chat(ChatService_ChatServer) error {
	return status.Errorf(codes.Unimplemented, "method Chat not implemented")
}

// UnsafeChatServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChatServiceServer will
// result in compilation errors.
type UnsafeChatServiceServer interface {
	mustEmbedUnimplementedChatServiceServer()
}

func RegisterChatServiceServer(s grpc.ServiceRegistrar, srv ChatServiceServer) {
	s.RegisterService(&ChatService_ServiceDesc, srv)
}

func _ChatService_SignUp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestSignUp)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServiceServer).SignUp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chat.ChatService/SignUp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServiceServer).SignUp(ctx, req.(*RequestSignUp))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChatService_SignIn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestSignIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServiceServer).SignIn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chat.ChatService/SignIn",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServiceServer).SignIn(ctx, req.(*RequestSignIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChatService_Chat_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ChatServiceServer).Chat(&chatServiceChatServer{stream})
}

type ChatService_ChatServer interface {
	Send(*ResponseMessage) error
	Recv() (*RequestMessage, error)
	grpc.ServerStream
}

type chatServiceChatServer struct {
	grpc.ServerStream
}

func (x *chatServiceChatServer) Send(m *ResponseMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *chatServiceChatServer) Recv() (*RequestMessage, error) {
	m := new(RequestMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChatService_ServiceDesc is the grpc.ServiceDesc for ChatService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ChatService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "chat.ChatService",
	HandlerType: (*ChatServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SignUp",
			Handler:    _ChatService_SignUp_Handler,
		},
		{
			MethodName: "SignIn",
			Handler:    _ChatService_SignIn_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Chat",
			Handler:       _ChatService_Chat_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "chatpb/chat.proto",
}