// Code generated by protoc-gen-go. DO NOT EDIT.
// source: husky.proto

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	husky.proto

It has these top-level messages:
	Header
	BytesMessage
	StringMessage
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// 包头
type Header struct {
	MessageId        *string `protobuf:"bytes,1,req,name=messageId" json:"messageId,omitempty"`
	MessageType      *string `protobuf:"bytes,2,req,name=messageType" json:"messageType,omitempty"`
	CreateTime       *int64  `protobuf:"varint,3,opt,name=createTime" json:"createTime,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Header) Reset()                    { *m = Header{} }
func (m *Header) String() string            { return proto.CompactTextString(m) }
func (*Header) ProtoMessage()               {}
func (*Header) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Header) GetMessageId() string {
	if m != nil && m.MessageId != nil {
		return *m.MessageId
	}
	return ""
}

func (m *Header) GetMessageType() string {
	if m != nil && m.MessageType != nil {
		return *m.MessageType
	}
	return ""
}

func (m *Header) GetCreateTime() int64 {
	if m != nil && m.CreateTime != nil {
		return *m.CreateTime
	}
	return 0
}

// byte类消息
type BytesMessage struct {
	Header           *Header `protobuf:"bytes,1,req,name=header" json:"header,omitempty"`
	Body             []byte  `protobuf:"bytes,2,req,name=body" json:"body,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *BytesMessage) Reset()                    { *m = BytesMessage{} }
func (m *BytesMessage) String() string            { return proto.CompactTextString(m) }
func (*BytesMessage) ProtoMessage()               {}
func (*BytesMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *BytesMessage) GetHeader() *Header {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *BytesMessage) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

// string类型的message
type StringMessage struct {
	Header           *Header `protobuf:"bytes,1,req,name=header" json:"header,omitempty"`
	Body             *string `protobuf:"bytes,2,req,name=body" json:"body,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *StringMessage) Reset()                    { *m = StringMessage{} }
func (m *StringMessage) String() string            { return proto.CompactTextString(m) }
func (*StringMessage) ProtoMessage()               {}
func (*StringMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *StringMessage) GetHeader() *Header {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *StringMessage) GetBody() string {
	if m != nil && m.Body != nil {
		return *m.Body
	}
	return ""
}

func init() {
	proto.RegisterType((*Header)(nil), "pb.Header")
	proto.RegisterType((*BytesMessage)(nil), "pb.BytesMessage")
	proto.RegisterType((*StringMessage)(nil), "pb.StringMessage")
}

func init() { proto.RegisterFile("husky.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 174 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0xce, 0x28, 0x2d, 0xce,
	0xae, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2a, 0x48, 0x52, 0xca, 0xe0, 0x62, 0xf3,
	0x48, 0x4d, 0x4c, 0x49, 0x2d, 0x12, 0x92, 0xe1, 0xe2, 0xcc, 0x4d, 0x2d, 0x2e, 0x4e, 0x4c, 0x4f,
	0xf5, 0x4c, 0x91, 0x60, 0x54, 0x60, 0xd2, 0xe0, 0x0c, 0x42, 0x08, 0x08, 0x29, 0x70, 0x71, 0x43,
	0x39, 0x21, 0x95, 0x05, 0xa9, 0x12, 0x4c, 0x60, 0x79, 0x64, 0x21, 0x21, 0x39, 0x2e, 0xae, 0xe4,
	0xa2, 0xd4, 0xc4, 0x92, 0xd4, 0x90, 0xcc, 0xdc, 0x54, 0x09, 0x66, 0x05, 0x46, 0x0d, 0xe6, 0x20,
	0x24, 0x11, 0x25, 0x37, 0x2e, 0x1e, 0xa7, 0xca, 0x92, 0xd4, 0x62, 0x5f, 0x88, 0x1e, 0x21, 0x25,
	0x2e, 0xb6, 0x0c, 0xb0, 0xcd, 0x60, 0xcb, 0xb8, 0x8d, 0xb8, 0xf4, 0x0a, 0x92, 0xf4, 0x20, 0x6e,
	0x09, 0x82, 0xca, 0x08, 0x09, 0x71, 0xb1, 0x24, 0xe5, 0xa7, 0x54, 0x82, 0xad, 0xe3, 0x09, 0x02,
	0xb3, 0x95, 0xdc, 0xb9, 0x78, 0x83, 0x4b, 0x8a, 0x32, 0xf3, 0xd2, 0xc9, 0x35, 0x88, 0x13, 0x62,
	0x10, 0x20, 0x00, 0x00, 0xff, 0xff, 0xef, 0x94, 0x34, 0xf2, 0x0c, 0x01, 0x00, 0x00,
}
