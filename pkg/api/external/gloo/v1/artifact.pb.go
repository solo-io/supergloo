// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: artifact.proto

package v1 // import "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"
import core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

import bytes "bytes"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

//
// @solo-kit:resource.short_name=art
// @solo-kit:resource.plural_name=artifacts
// @solo-kit:resource.resource_groups=api.gloo.solo.io
//
// Gloo Artifacts are used by Gloo to store small bits of binary or file data.
//
// Certain plugins such as the gRPC plugin read and write artifacts to one of Gloo's configured
// storage layer.
//
// Artifacts can be backed by files on disk, Kubernetes ConfigMaps, and Consul Key/Value pairs.
//
// Supported artifact backends can be selected in Gloo's boostrap options.
type Artifact struct {
	// Raw data data being stored
	Data string `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	// Metadata contains the object metadata for this resource
	Metadata             core.Metadata `protobuf:"bytes,7,opt,name=metadata" json:"metadata"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Artifact) Reset()         { *m = Artifact{} }
func (m *Artifact) String() string { return proto.CompactTextString(m) }
func (*Artifact) ProtoMessage()    {}
func (*Artifact) Descriptor() ([]byte, []int) {
	return fileDescriptor_artifact_b8402968e57a42f5, []int{0}
}
func (m *Artifact) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Artifact.Unmarshal(m, b)
}
func (m *Artifact) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Artifact.Marshal(b, m, deterministic)
}
func (dst *Artifact) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Artifact.Merge(dst, src)
}
func (m *Artifact) XXX_Size() int {
	return xxx_messageInfo_Artifact.Size(m)
}
func (m *Artifact) XXX_DiscardUnknown() {
	xxx_messageInfo_Artifact.DiscardUnknown(m)
}

var xxx_messageInfo_Artifact proto.InternalMessageInfo

func (m *Artifact) GetData() string {
	if m != nil {
		return m.Data
	}
	return ""
}

func (m *Artifact) GetMetadata() core.Metadata {
	if m != nil {
		return m.Metadata
	}
	return core.Metadata{}
}

func init() {
	proto.RegisterType((*Artifact)(nil), "gloo.solo.io.Artifact")
}
func (this *Artifact) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Artifact)
	if !ok {
		that2, ok := that.(Artifact)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Data != that1.Data {
		return false
	}
	if !this.Metadata.Equal(&that1.Metadata) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}

func init() { proto.RegisterFile("artifact.proto", fileDescriptor_artifact_b8402968e57a42f5) }

var fileDescriptor_artifact_b8402968e57a42f5 = []byte{
	// 205 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4b, 0x2c, 0x2a, 0xc9,
	0x4c, 0x4b, 0x4c, 0x2e, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x49, 0xcf, 0xc9, 0xcf,
	0xd7, 0x2b, 0xce, 0xcf, 0xc9, 0xd7, 0xcb, 0xcc, 0x97, 0x12, 0x49, 0xcf, 0x4f, 0xcf, 0x07, 0x4b,
	0xe8, 0x83, 0x58, 0x10, 0x35, 0x52, 0x86, 0xe9, 0x99, 0x25, 0x19, 0xa5, 0x49, 0x7a, 0xc9, 0xf9,
	0xb9, 0xfa, 0x20, 0x95, 0xba, 0x99, 0xf9, 0x10, 0x3a, 0x3b, 0xb3, 0x44, 0x3f, 0xb1, 0x20, 0x53,
	0xbf, 0xcc, 0x50, 0x3f, 0x37, 0xb5, 0x24, 0x31, 0x25, 0xb1, 0x24, 0x11, 0xa2, 0x45, 0x29, 0x82,
	0x8b, 0xc3, 0x11, 0x6a, 0x91, 0x90, 0x10, 0x17, 0x0b, 0x48, 0x46, 0x82, 0x51, 0x81, 0x51, 0x83,
	0x33, 0x08, 0xcc, 0x16, 0xb2, 0xe0, 0xe2, 0x80, 0xe9, 0x90, 0x60, 0x57, 0x60, 0xd4, 0xe0, 0x36,
	0x12, 0xd3, 0x4b, 0xce, 0x2f, 0x4a, 0x85, 0xb9, 0x44, 0xcf, 0x17, 0x2a, 0xeb, 0xc4, 0x72, 0xe2,
	0x9e, 0x3c, 0x43, 0x10, 0x5c, 0xb5, 0x93, 0xf5, 0x8a, 0x47, 0x72, 0x8c, 0x51, 0xa6, 0xd8, 0x9c,
	0x54, 0x5a, 0x90, 0x5a, 0x04, 0xf2, 0x8e, 0x7e, 0x41, 0x76, 0x3a, 0xd8, 0x5d, 0xa9, 0x15, 0x25,
	0xa9, 0x45, 0x79, 0x89, 0x39, 0xfa, 0x60, 0xd1, 0x32, 0xc3, 0x24, 0x36, 0xb0, 0xeb, 0x8c, 0x01,
	0x01, 0x00, 0x00, 0xff, 0xff, 0x98, 0xd6, 0x89, 0xb5, 0x06, 0x01, 0x00, 0x00,
}
