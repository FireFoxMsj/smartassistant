// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/datatunnel/proto/datatunnel_control.proto

package proto // import "github.com/zhiting-tech/smartassistant/pkg/datatunnel/proto"

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

type ControlStreamData struct {
	Version              int32    `protobuf:"varint,1,opt,name=Version,proto3" json:"Version,omitempty"`
	Action               string   `protobuf:"bytes,2,opt,name=Action,proto3" json:"Action,omitempty"`
	ActionValue          string   `protobuf:"bytes,3,opt,name=ActionValue,proto3" json:"ActionValue,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ControlStreamData) Reset()         { *m = ControlStreamData{} }
func (m *ControlStreamData) String() string { return proto.CompactTextString(m) }
func (*ControlStreamData) ProtoMessage()    {}
func (*ControlStreamData) Descriptor() ([]byte, []int) {
	return fileDescriptor_datatunnel_control_a77fc5aeeadcf792, []int{0}
}
func (m *ControlStreamData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ControlStreamData.Unmarshal(m, b)
}
func (m *ControlStreamData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ControlStreamData.Marshal(b, m, deterministic)
}
func (dst *ControlStreamData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ControlStreamData.Merge(dst, src)
}
func (m *ControlStreamData) XXX_Size() int {
	return xxx_messageInfo_ControlStreamData.Size(m)
}
func (m *ControlStreamData) XXX_DiscardUnknown() {
	xxx_messageInfo_ControlStreamData.DiscardUnknown(m)
}

var xxx_messageInfo_ControlStreamData proto.InternalMessageInfo

func (m *ControlStreamData) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *ControlStreamData) GetAction() string {
	if m != nil {
		return m.Action
	}
	return ""
}

func (m *ControlStreamData) GetActionValue() string {
	if m != nil {
		return m.ActionValue
	}
	return ""
}

func init() {
	proto.RegisterType((*ControlStreamData)(nil), "proto.ControlStreamData")
}

func init() {
	proto.RegisterFile("pkg/datatunnel/proto/datatunnel_control.proto", fileDescriptor_datatunnel_control_a77fc5aeeadcf792)
}

var fileDescriptor_datatunnel_control_a77fc5aeeadcf792 = []byte{
	// 215 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x90, 0x31, 0x4b, 0xc5, 0x30,
	0x14, 0x85, 0x89, 0xf2, 0x2a, 0x46, 0x1c, 0x0c, 0x22, 0xc1, 0xa9, 0x38, 0x75, 0x69, 0x23, 0x3a,
	0x8a, 0x83, 0x5a, 0x70, 0xaf, 0xd0, 0xc1, 0xa5, 0xdc, 0xc6, 0x90, 0x06, 0xd3, 0xa4, 0x24, 0xb7,
	0x8b, 0xbf, 0x5e, 0x4c, 0xab, 0x54, 0xf4, 0x4d, 0x27, 0xe7, 0x7c, 0x43, 0x3e, 0x2e, 0x2d, 0xa7,
	0x77, 0x2d, 0xde, 0x00, 0x01, 0x67, 0xe7, 0x94, 0x15, 0x53, 0xf0, 0xe8, 0x37, 0x43, 0x27, 0xbd,
	0xc3, 0xe0, 0x6d, 0x95, 0x00, 0xdb, 0xa5, 0xb8, 0xd2, 0xf4, 0xec, 0x69, 0xd9, 0x5f, 0x30, 0x28,
	0x18, 0x6b, 0x40, 0x60, 0x9c, 0x1e, 0xb5, 0x2a, 0x44, 0xe3, 0x1d, 0x27, 0x39, 0x29, 0x76, 0xcd,
	0x77, 0x65, 0x17, 0x34, 0x7b, 0x90, 0xf8, 0x05, 0x0e, 0x72, 0x52, 0x1c, 0x37, 0x6b, 0x63, 0x39,
	0x3d, 0x59, 0x5e, 0x2d, 0xd8, 0x59, 0xf1, 0xc3, 0x04, 0xb7, 0xd3, 0x4d, 0x47, 0xcf, 0xeb, 0x1f,
	0x97, 0xf5, 0x4b, 0xab, 0x02, 0x7b, 0xa6, 0xa7, 0xbf, 0x04, 0x18, 0x5f, 0x04, 0xab, 0x3f, 0x5a,
	0x97, 0x7b, 0x49, 0x41, 0xae, 0xc9, 0xe3, 0xfd, 0xeb, 0x9d, 0x36, 0x38, 0xcc, 0x7d, 0x25, 0xfd,
	0x28, 0x3e, 0x06, 0x83, 0xc6, 0xe9, 0x12, 0x95, 0x1c, 0x44, 0x1c, 0x21, 0x20, 0xc4, 0x68, 0x22,
	0x82, 0x43, 0xf1, 0xdf, 0xa1, 0xfa, 0x2c, 0xc5, 0xed, 0x67, 0x00, 0x00, 0x00, 0xff, 0xff, 0xbe,
	0xbe, 0xcf, 0xc7, 0x47, 0x01, 0x00, 0x00,
}
