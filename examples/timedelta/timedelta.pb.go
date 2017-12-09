// Code generated by protoc-gen-go. DO NOT EDIT.
// source: timedelta.proto

/*
Package timedelta is a generated protocol buffer package.

It is generated from these files:
	timedelta.proto

It has these top-level messages:
	TimeDeltaRequest
	TimeDeltaResponse
	SleepRequest
	SleepResponse
*/
package timedelta

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
import google_protobuf1 "github.com/golang/protobuf/ptypes/duration"

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

type TimeDeltaRequest struct {
	Now *google_protobuf.Timestamp `protobuf:"bytes,1,opt,name=now" json:"now,omitempty"`
}

func (m *TimeDeltaRequest) Reset()                    { *m = TimeDeltaRequest{} }
func (m *TimeDeltaRequest) String() string            { return proto.CompactTextString(m) }
func (*TimeDeltaRequest) ProtoMessage()               {}
func (*TimeDeltaRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *TimeDeltaRequest) GetNow() *google_protobuf.Timestamp {
	if m != nil {
		return m.Now
	}
	return nil
}

type TimeDeltaResponse struct {
	Now   *google_protobuf.Timestamp `protobuf:"bytes,1,opt,name=now" json:"now,omitempty"`
	Delta *google_protobuf1.Duration `protobuf:"bytes,2,opt,name=delta" json:"delta,omitempty"`
}

func (m *TimeDeltaResponse) Reset()                    { *m = TimeDeltaResponse{} }
func (m *TimeDeltaResponse) String() string            { return proto.CompactTextString(m) }
func (*TimeDeltaResponse) ProtoMessage()               {}
func (*TimeDeltaResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *TimeDeltaResponse) GetNow() *google_protobuf.Timestamp {
	if m != nil {
		return m.Now
	}
	return nil
}

func (m *TimeDeltaResponse) GetDelta() *google_protobuf1.Duration {
	if m != nil {
		return m.Delta
	}
	return nil
}

type SleepRequest struct {
	Duration *google_protobuf1.Duration `protobuf:"bytes,1,opt,name=duration" json:"duration,omitempty"`
}

func (m *SleepRequest) Reset()                    { *m = SleepRequest{} }
func (m *SleepRequest) String() string            { return proto.CompactTextString(m) }
func (*SleepRequest) ProtoMessage()               {}
func (*SleepRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *SleepRequest) GetDuration() *google_protobuf1.Duration {
	if m != nil {
		return m.Duration
	}
	return nil
}

type SleepResponse struct {
}

func (m *SleepResponse) Reset()                    { *m = SleepResponse{} }
func (m *SleepResponse) String() string            { return proto.CompactTextString(m) }
func (*SleepResponse) ProtoMessage()               {}
func (*SleepResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func init() {
	proto.RegisterType((*TimeDeltaRequest)(nil), "timedelta.TimeDeltaRequest")
	proto.RegisterType((*TimeDeltaResponse)(nil), "timedelta.TimeDeltaResponse")
	proto.RegisterType((*SleepRequest)(nil), "timedelta.SleepRequest")
	proto.RegisterType((*SleepResponse)(nil), "timedelta.SleepResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Timedelta service

type TimedeltaClient interface {
	TimeDelta(ctx context.Context, in *TimeDeltaRequest, opts ...grpc.CallOption) (*TimeDeltaResponse, error)
	Sleep(ctx context.Context, in *SleepRequest, opts ...grpc.CallOption) (*SleepResponse, error)
}

type timedeltaClient struct {
	cc *grpc.ClientConn
}

func NewTimedeltaClient(cc *grpc.ClientConn) TimedeltaClient {
	return &timedeltaClient{cc}
}

func (c *timedeltaClient) TimeDelta(ctx context.Context, in *TimeDeltaRequest, opts ...grpc.CallOption) (*TimeDeltaResponse, error) {
	out := new(TimeDeltaResponse)
	err := grpc.Invoke(ctx, "/timedelta.timedelta/TimeDelta", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *timedeltaClient) Sleep(ctx context.Context, in *SleepRequest, opts ...grpc.CallOption) (*SleepResponse, error) {
	out := new(SleepResponse)
	err := grpc.Invoke(ctx, "/timedelta.timedelta/Sleep", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Timedelta service

type TimedeltaServer interface {
	TimeDelta(context.Context, *TimeDeltaRequest) (*TimeDeltaResponse, error)
	Sleep(context.Context, *SleepRequest) (*SleepResponse, error)
}

func RegisterTimedeltaServer(s *grpc.Server, srv TimedeltaServer) {
	s.RegisterService(&_Timedelta_serviceDesc, srv)
}

func _Timedelta_TimeDelta_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TimeDeltaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TimedeltaServer).TimeDelta(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/timedelta.timedelta/TimeDelta",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TimedeltaServer).TimeDelta(ctx, req.(*TimeDeltaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Timedelta_Sleep_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SleepRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TimedeltaServer).Sleep(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/timedelta.timedelta/Sleep",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TimedeltaServer).Sleep(ctx, req.(*SleepRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Timedelta_serviceDesc = grpc.ServiceDesc{
	ServiceName: "timedelta.timedelta",
	HandlerType: (*TimedeltaServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "TimeDelta",
			Handler:    _Timedelta_TimeDelta_Handler,
		},
		{
			MethodName: "Sleep",
			Handler:    _Timedelta_Sleep_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "timedelta.proto",
}

func init() { proto.RegisterFile("timedelta.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 246 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x2f, 0xc9, 0xcc, 0x4d,
	0x4d, 0x49, 0xcd, 0x29, 0x49, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x84, 0x0b, 0x48,
	0xc9, 0xa7, 0xe7, 0xe7, 0xa7, 0xe7, 0xa4, 0xea, 0x83, 0x25, 0x92, 0x4a, 0xd3, 0xf4, 0x41, 0x52,
	0xc5, 0x25, 0x89, 0xb9, 0x05, 0x10, 0xb5, 0x52, 0x72, 0xe8, 0x0a, 0x52, 0x4a, 0x8b, 0x12, 0x4b,
	0x32, 0xf3, 0xf3, 0x20, 0xf2, 0x4a, 0x0e, 0x5c, 0x02, 0x21, 0x40, 0x2d, 0x2e, 0x20, 0xd3, 0x82,
	0x52, 0x0b, 0x4b, 0x81, 0x9a, 0x85, 0x74, 0xb8, 0x98, 0xf3, 0xf2, 0xcb, 0x25, 0x18, 0x15, 0x18,
	0x35, 0xb8, 0x8d, 0xa4, 0xf4, 0x20, 0x26, 0xe8, 0xc1, 0x4c, 0xd0, 0x0b, 0x81, 0x59, 0x11, 0x04,
	0x52, 0xa6, 0x54, 0xc4, 0x25, 0x88, 0x64, 0x42, 0x71, 0x41, 0x7e, 0x5e, 0x71, 0x2a, 0x69, 0x46,
	0x08, 0xe9, 0x73, 0xb1, 0x82, 0xbd, 0x23, 0xc1, 0x04, 0x56, 0x2f, 0x89, 0xa1, 0xde, 0x05, 0xea,
	0xe8, 0x20, 0x88, 0x3a, 0x25, 0x57, 0x2e, 0x9e, 0xe0, 0x9c, 0xd4, 0xd4, 0x02, 0x98, 0x8b, 0x4d,
	0xb9, 0x38, 0x60, 0xfe, 0x82, 0xda, 0x89, 0xc7, 0x0c, 0xb8, 0x52, 0x25, 0x7e, 0x2e, 0x5e, 0xa8,
	0x31, 0x10, 0x67, 0x1b, 0x4d, 0x66, 0xe4, 0x42, 0x04, 0xae, 0x90, 0x07, 0x17, 0x27, 0xdc, 0x67,
	0x42, 0xd2, 0x7a, 0x88, 0x68, 0x40, 0x0f, 0x31, 0x29, 0x19, 0xec, 0x92, 0x10, 0x53, 0x95, 0x18,
	0x84, 0x6c, 0xb8, 0x58, 0xc1, 0x16, 0x09, 0x89, 0x23, 0x29, 0x44, 0xf6, 0x81, 0x94, 0x04, 0xa6,
	0x04, 0x4c, 0x77, 0x12, 0x1b, 0xd8, 0x0f, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x4c, 0xbf,
	0xf7, 0x21, 0x09, 0x02, 0x00, 0x00,
}
