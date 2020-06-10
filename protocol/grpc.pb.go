// Code generated by protoc-gen-go. DO NOT EDIT.
// source: grpc.proto

package protocol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ScoClient is the client API for Sco service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ScoClient interface {
	HandlerCall(ctx context.Context, in *HandlerReq, opts ...grpc.CallOption) (*HandlerRes, error)
	RemoteCall(ctx context.Context, in *RemoteReq, opts ...grpc.CallOption) (*RemoteRes, error)
}

type scoClient struct {
	cc *grpc.ClientConn
}

func NewScoClient(cc *grpc.ClientConn) ScoClient {
	return &scoClient{cc}
}

func (c *scoClient) HandlerCall(ctx context.Context, in *HandlerReq, opts ...grpc.CallOption) (*HandlerRes, error) {
	out := new(HandlerRes)
	err := c.cc.Invoke(ctx, "/protocol.Sco/HandlerCall", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *scoClient) RemoteCall(ctx context.Context, in *RemoteReq, opts ...grpc.CallOption) (*RemoteRes, error) {
	out := new(RemoteRes)
	err := c.cc.Invoke(ctx, "/protocol.Sco/RemoteCall", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ScoServer is the server API for Sco service.
type ScoServer interface {
	HandlerCall(context.Context, *HandlerReq) (*HandlerRes, error)
	RemoteCall(context.Context, *RemoteReq) (*RemoteRes, error)
}

func RegisterScoServer(s *grpc.Server, srv ScoServer) {
	s.RegisterService(&_Sco_serviceDesc, srv)
}

func _Sco_HandlerCall_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandlerReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ScoServer).HandlerCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.Sco/HandlerCall",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ScoServer).HandlerCall(ctx, req.(*HandlerReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Sco_RemoteCall_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoteReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ScoServer).RemoteCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.Sco/RemoteCall",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ScoServer).RemoteCall(ctx, req.(*RemoteReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Sco_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protocol.Sco",
	HandlerType: (*ScoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HandlerCall",
			Handler:    _Sco_HandlerCall_Handler,
		},
		{
			MethodName: "RemoteCall",
			Handler:    _Sco_RemoteCall_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc.proto",
}

func init() { proto.RegisterFile("grpc.proto", fileDescriptor_grpc_e787ae16f38597f1) }

var fileDescriptor_grpc_e787ae16f38597f1 = []byte{
	// 124 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4a, 0x2f, 0x2a, 0x48,
	0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x00, 0x53, 0xc9, 0xf9, 0x39, 0x52, 0xbc, 0x19,
	0x89, 0x79, 0x29, 0x39, 0xa9, 0x45, 0x10, 0x09, 0x29, 0x9e, 0xa2, 0xd4, 0xdc, 0xfc, 0x92, 0x54,
	0x08, 0xcf, 0xa8, 0x86, 0x8b, 0x39, 0x38, 0x39, 0x5f, 0xc8, 0x9a, 0x8b, 0xdb, 0x03, 0xa2, 0xca,
	0x39, 0x31, 0x27, 0x47, 0x48, 0x44, 0x0f, 0xa6, 0x5b, 0x0f, 0x2a, 0x1c, 0x94, 0x5a, 0x28, 0x85,
	0x4d, 0xb4, 0x58, 0x89, 0x41, 0xc8, 0x82, 0x8b, 0x2b, 0x08, 0x6c, 0x26, 0x58, 0xaf, 0x30, 0x42,
	0x15, 0x44, 0x14, 0xa4, 0x15, 0x8b, 0x60, 0xb1, 0x12, 0x43, 0x12, 0x1b, 0x58, 0xd4, 0x18, 0x10,
	0x00, 0x00, 0xff, 0xff, 0x3b, 0x86, 0x17, 0x23, 0xb9, 0x00, 0x00, 0x00,
}
