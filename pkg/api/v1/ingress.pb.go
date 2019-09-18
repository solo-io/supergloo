// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/supergloo/api/v1/ingress.proto

package v1

import (
	bytes "bytes"
	fmt "fmt"
	math "math"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

//
//MeshIngress represents a managed ingress (edge router) which can proxy connections
//for services in Mesh managed by SuperGloo. SuperGloo will perform additional configuration,
//if necessary, to enable proxying services which are using mTLS for communication.
type MeshIngress struct {
	// Status indicates the validation status of this resource.
	// Status is read-only by clients, and set by supergloo during validation
	Status core.Status `protobuf:"bytes,100,opt,name=status,proto3" json:"status" testdiff:"ignore"`
	// Metadata contains the object metadata for this resource
	Metadata core.Metadata `protobuf:"bytes,101,opt,name=metadata,proto3" json:"metadata"`
	// type of Mesh ingress represented by this resource
	//
	// Types that are valid to be assigned to MeshIngressType:
	//	*MeshIngress_Gloo
	MeshIngressType isMeshIngress_MeshIngressType `protobuf_oneof:"mesh_ingress_type"`
	// where the ingress has been installed
	InstallationNamespace string `protobuf:"bytes,2,opt,name=installation_namespace,json=installationNamespace,proto3" json:"installation_namespace,omitempty"`
	// reference to the Mesh(s) that this ingress is acting upon
	// enable the ingress to route to services within these mTLS-enabled meshes
	Meshes               []*core.ResourceRef `protobuf:"bytes,3,rep,name=meshes,proto3" json:"meshes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *MeshIngress) Reset()         { *m = MeshIngress{} }
func (m *MeshIngress) String() string { return proto.CompactTextString(m) }
func (*MeshIngress) ProtoMessage()    {}
func (*MeshIngress) Descriptor() ([]byte, []int) {
	return fileDescriptor_5d220723bdc29937, []int{0}
}
func (m *MeshIngress) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MeshIngress.Unmarshal(m, b)
}
func (m *MeshIngress) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MeshIngress.Marshal(b, m, deterministic)
}
func (m *MeshIngress) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MeshIngress.Merge(m, src)
}
func (m *MeshIngress) XXX_Size() int {
	return xxx_messageInfo_MeshIngress.Size(m)
}
func (m *MeshIngress) XXX_DiscardUnknown() {
	xxx_messageInfo_MeshIngress.DiscardUnknown(m)
}

var xxx_messageInfo_MeshIngress proto.InternalMessageInfo

type isMeshIngress_MeshIngressType interface {
	isMeshIngress_MeshIngressType()
	Equal(interface{}) bool
}

type MeshIngress_Gloo struct {
	Gloo *GlooMeshIngress `protobuf:"bytes,1,opt,name=gloo,proto3,oneof" json:"gloo,omitempty"`
}

func (*MeshIngress_Gloo) isMeshIngress_MeshIngressType() {}

func (m *MeshIngress) GetMeshIngressType() isMeshIngress_MeshIngressType {
	if m != nil {
		return m.MeshIngressType
	}
	return nil
}

func (m *MeshIngress) GetStatus() core.Status {
	if m != nil {
		return m.Status
	}
	return core.Status{}
}

func (m *MeshIngress) GetMetadata() core.Metadata {
	if m != nil {
		return m.Metadata
	}
	return core.Metadata{}
}

func (m *MeshIngress) GetGloo() *GlooMeshIngress {
	if x, ok := m.GetMeshIngressType().(*MeshIngress_Gloo); ok {
		return x.Gloo
	}
	return nil
}

func (m *MeshIngress) GetInstallationNamespace() string {
	if m != nil {
		return m.InstallationNamespace
	}
	return ""
}

func (m *MeshIngress) GetMeshes() []*core.ResourceRef {
	if m != nil {
		return m.Meshes
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*MeshIngress) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*MeshIngress_Gloo)(nil),
	}
}

// Mesh ingress object for gloo
type GlooMeshIngress struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GlooMeshIngress) Reset()         { *m = GlooMeshIngress{} }
func (m *GlooMeshIngress) String() string { return proto.CompactTextString(m) }
func (*GlooMeshIngress) ProtoMessage()    {}
func (*GlooMeshIngress) Descriptor() ([]byte, []int) {
	return fileDescriptor_5d220723bdc29937, []int{1}
}
func (m *GlooMeshIngress) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GlooMeshIngress.Unmarshal(m, b)
}
func (m *GlooMeshIngress) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GlooMeshIngress.Marshal(b, m, deterministic)
}
func (m *GlooMeshIngress) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GlooMeshIngress.Merge(m, src)
}
func (m *GlooMeshIngress) XXX_Size() int {
	return xxx_messageInfo_GlooMeshIngress.Size(m)
}
func (m *GlooMeshIngress) XXX_DiscardUnknown() {
	xxx_messageInfo_GlooMeshIngress.DiscardUnknown(m)
}

var xxx_messageInfo_GlooMeshIngress proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MeshIngress)(nil), "supergloo.solo.io.MeshIngress")
	proto.RegisterType((*GlooMeshIngress)(nil), "supergloo.solo.io.GlooMeshIngress")
}

func init() {
	proto.RegisterFile("github.com/solo-io/supergloo/api/v1/ingress.proto", fileDescriptor_5d220723bdc29937)
}

var fileDescriptor_5d220723bdc29937 = []byte{
	// 387 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0xc1, 0xce, 0xd2, 0x40,
	0x10, 0xc7, 0x29, 0x34, 0x44, 0x97, 0x18, 0xd2, 0x8a, 0xa4, 0x60, 0x22, 0xa4, 0x17, 0x39, 0xc8,
	0x36, 0x60, 0x4c, 0x08, 0xc7, 0x5e, 0xd0, 0x03, 0x1e, 0xea, 0xcd, 0x0b, 0x59, 0xca, 0xb4, 0x6c,
	0x68, 0x3b, 0x4d, 0x77, 0x6b, 0xe2, 0x95, 0xa7, 0xf1, 0x51, 0x7c, 0x0a, 0x0e, 0x1e, 0xbc, 0xe3,
	0x13, 0x98, 0xb6, 0x5b, 0x02, 0xfa, 0xe5, 0x0b, 0xdf, 0xa9, 0xed, 0xfc, 0xe7, 0xf7, 0xcf, 0xbf,
	0x33, 0x43, 0x66, 0x21, 0x97, 0xfb, 0x7c, 0x4b, 0x7d, 0x8c, 0x1d, 0x81, 0x11, 0x4e, 0x39, 0x3a,
	0x22, 0x4f, 0x21, 0x0b, 0x23, 0x44, 0x87, 0xa5, 0xdc, 0xf9, 0x36, 0x73, 0x78, 0x12, 0x66, 0x20,
	0x04, 0x4d, 0x33, 0x94, 0x68, 0x1a, 0x17, 0x9d, 0x16, 0x04, 0xe5, 0x38, 0xec, 0x85, 0x18, 0x62,
	0xa9, 0x3a, 0xc5, 0x5b, 0xd5, 0x38, 0x7c, 0xd0, 0xbb, 0x78, 0x1e, 0xb8, 0xac, 0xad, 0x63, 0x90,
	0x6c, 0xc7, 0x24, 0x53, 0x88, 0x73, 0x07, 0x22, 0x24, 0x93, 0xb9, 0x0a, 0x33, 0x7c, 0x77, 0x07,
	0x90, 0x41, 0xf0, 0x84, 0x44, 0xf5, 0x77, 0x85, 0xd8, 0xbf, 0x9b, 0xa4, 0xb3, 0x06, 0xb1, 0xff,
	0x54, 0xcd, 0xc0, 0x5c, 0x91, 0x76, 0x15, 0xc0, 0xda, 0x8d, 0xb5, 0x49, 0x67, 0xde, 0xa3, 0x3e,
	0x66, 0x50, 0x4f, 0x82, 0x7e, 0x29, 0x35, 0x77, 0xf0, 0xf3, 0x34, 0x6a, 0xfc, 0x39, 0x8d, 0x0c,
	0x09, 0x42, 0xee, 0x78, 0x10, 0x2c, 0x6d, 0x1e, 0x26, 0x98, 0x81, 0xed, 0x29, 0xdc, 0x5c, 0x90,
	0x67, 0xf5, 0xcf, 0x5b, 0x50, 0x5a, 0xf5, 0x6f, 0xad, 0xd6, 0x4a, 0x75, 0xf5, 0xc2, 0xcc, 0xbb,
	0x74, 0x9b, 0x0b, 0xa2, 0x17, 0xd3, 0xb7, 0xb4, 0x92, 0xb2, 0xe9, 0x7f, 0xfb, 0xa0, 0xab, 0x08,
	0xf1, 0x2a, 0xf4, 0xc7, 0x86, 0x57, 0x12, 0xe6, 0x07, 0xd2, 0xe7, 0x89, 0x90, 0x2c, 0x8a, 0x98,
	0xe4, 0x98, 0x6c, 0x12, 0x16, 0x83, 0x48, 0x99, 0x0f, 0x56, 0x73, 0xac, 0x4d, 0x9e, 0x7b, 0xaf,
	0xae, 0xd5, 0xcf, 0xb5, 0x68, 0xce, 0x48, 0x3b, 0x06, 0xb1, 0x07, 0x61, 0xb5, 0xc6, 0xad, 0x49,
	0x67, 0x3e, 0xb8, 0x0d, 0xea, 0x81, 0xc0, 0x3c, 0xf3, 0xc1, 0x83, 0xc0, 0x53, 0x8d, 0xcb, 0xd7,
	0xc7, 0xb3, 0xae, 0x93, 0x66, 0xcc, 0x8f, 0x67, 0xbd, 0x6b, 0xbe, 0x28, 0xaa, 0xea, 0x8a, 0x40,
	0xb8, 0x2f, 0x89, 0x51, 0x14, 0x36, 0xaa, 0xb2, 0x91, 0xdf, 0x53, 0xb0, 0x0d, 0xd2, 0xfd, 0x27,
	0xb6, 0x3b, 0xfd, 0xf1, 0xeb, 0x8d, 0xf6, 0xf5, 0xed, 0xa3, 0x27, 0x9a, 0x1e, 0x42, 0xb5, 0xb9,
	0x6d, 0xbb, 0xdc, 0xd8, 0xfb, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0xcc, 0xfd, 0x25, 0x86, 0xd4,
	0x02, 0x00, 0x00,
}

func (this *MeshIngress) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MeshIngress)
	if !ok {
		that2, ok := that.(MeshIngress)
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
	if !this.Status.Equal(&that1.Status) {
		return false
	}
	if !this.Metadata.Equal(&that1.Metadata) {
		return false
	}
	if that1.MeshIngressType == nil {
		if this.MeshIngressType != nil {
			return false
		}
	} else if this.MeshIngressType == nil {
		return false
	} else if !this.MeshIngressType.Equal(that1.MeshIngressType) {
		return false
	}
	if this.InstallationNamespace != that1.InstallationNamespace {
		return false
	}
	if len(this.Meshes) != len(that1.Meshes) {
		return false
	}
	for i := range this.Meshes {
		if !this.Meshes[i].Equal(that1.Meshes[i]) {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *MeshIngress_Gloo) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MeshIngress_Gloo)
	if !ok {
		that2, ok := that.(MeshIngress_Gloo)
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
	if !this.Gloo.Equal(that1.Gloo) {
		return false
	}
	return true
}
func (this *GlooMeshIngress) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*GlooMeshIngress)
	if !ok {
		that2, ok := that.(GlooMeshIngress)
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
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
