// Code generated by protoc-gen-go. DO NOT EDIT.
// source: entry.proto

/*
Package protobuf is a generated protocol buffer package.

It is generated from these files:
	entry.proto

It has these top-level messages:
	Entry
*/
package protobuf

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

type Entry struct {
	Fields    map[string]string `protobuf:"bytes,1,rep,name=Fields" json:"Fields,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Level     string            `protobuf:"bytes,2,opt,name=Level" json:"Level,omitempty"`
	Timestamp []byte            `protobuf:"bytes,3,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	Message   string            `protobuf:"bytes,4,opt,name=Message" json:"Message,omitempty"`
}

func (m *Entry) Reset()                    { *m = Entry{} }
func (m *Entry) String() string            { return proto.CompactTextString(m) }
func (*Entry) ProtoMessage()               {}
func (*Entry) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Entry) GetFields() map[string]string {
	if m != nil {
		return m.Fields
	}
	return nil
}

func (m *Entry) GetLevel() string {
	if m != nil {
		return m.Level
	}
	return ""
}

func (m *Entry) GetTimestamp() []byte {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *Entry) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*Entry)(nil), "protobuf.Entry")
}

func init() { proto.RegisterFile("entry.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 179 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4e, 0xcd, 0x2b, 0x29,
	0xaa, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x00, 0x53, 0x49, 0xa5, 0x69, 0x4a, 0x47,
	0x19, 0xb9, 0x58, 0x5d, 0x41, 0x32, 0x42, 0xc6, 0x5c, 0x6c, 0x6e, 0x99, 0xa9, 0x39, 0x29, 0xc5,
	0x12, 0x8c, 0x0a, 0xcc, 0x1a, 0xdc, 0x46, 0xd2, 0x7a, 0x30, 0x45, 0x7a, 0x60, 0x05, 0x7a, 0x10,
	0x59, 0x30, 0x3b, 0x08, 0xaa, 0x54, 0x48, 0x84, 0x8b, 0xd5, 0x27, 0xb5, 0x2c, 0x35, 0x47, 0x82,
	0x49, 0x81, 0x51, 0x83, 0x33, 0x08, 0xc2, 0x11, 0x92, 0xe1, 0xe2, 0x0c, 0xc9, 0xcc, 0x4d, 0x2d,
	0x2e, 0x49, 0xcc, 0x2d, 0x90, 0x60, 0x56, 0x60, 0xd4, 0xe0, 0x09, 0x42, 0x08, 0x08, 0x49, 0x70,
	0xb1, 0xfb, 0xa6, 0x16, 0x17, 0x27, 0xa6, 0xa7, 0x4a, 0xb0, 0x80, 0x75, 0xc1, 0xb8, 0x52, 0x96,
	0x5c, 0xdc, 0x48, 0x96, 0x08, 0x09, 0x70, 0x31, 0x67, 0xa7, 0x56, 0x4a, 0x30, 0x82, 0x15, 0x81,
	0x98, 0x20, 0xeb, 0xca, 0x12, 0x73, 0x4a, 0x53, 0x61, 0xd6, 0x81, 0x39, 0x56, 0x4c, 0x16, 0x8c,
	0x49, 0x6c, 0x60, 0xc7, 0x1a, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x99, 0xa7, 0xc7, 0x6e, 0xe7,
	0x00, 0x00, 0x00,
}
