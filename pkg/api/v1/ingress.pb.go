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
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

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
	// reference to the Mesh that this ingress is acting upon
	Mesh                 *core.ResourceRef `protobuf:"bytes,2,opt,name=mesh,proto3" json:"mesh,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
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
	Gloo *GlooMeshIngress `protobuf:"bytes,1,opt,name=gloo,proto3,oneof"`
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

func (m *MeshIngress) GetMesh() *core.ResourceRef {
	if m != nil {
		return m.Mesh
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*MeshIngress) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _MeshIngress_OneofMarshaler, _MeshIngress_OneofUnmarshaler, _MeshIngress_OneofSizer, []interface{}{
		(*MeshIngress_Gloo)(nil),
	}
}

func _MeshIngress_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*MeshIngress)
	// mesh_ingress_type
	switch x := m.MeshIngressType.(type) {
	case *MeshIngress_Gloo:
		_ = b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Gloo); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("MeshIngress.MeshIngressType has unexpected type %T", x)
	}
	return nil
}

func _MeshIngress_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*MeshIngress)
	switch tag {
	case 1: // mesh_ingress_type.gloo
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(GlooMeshIngress)
		err := b.DecodeMessage(msg)
		m.MeshIngressType = &MeshIngress_Gloo{msg}
		return true, err
	default:
		return false, nil
	}
}

func _MeshIngress_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*MeshIngress)
	// mesh_ingress_type
	switch x := m.MeshIngressType.(type) {
	case *MeshIngress_Gloo:
		s := proto.Size(x.Gloo)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Mesh ingress object for gloo
type GlooMeshIngress struct {
	// where Gloo has been installed
	InstallationNamespace string   `protobuf:"bytes,1,opt,name=installation_namespace,json=installationNamespace,proto3" json:"installation_namespace,omitempty"`
	XXX_NoUnkeyedLiteral  struct{} `json:"-"`
	XXX_unrecognized      []byte   `json:"-"`
	XXX_sizecache         int32    `json:"-"`
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

func (m *GlooMeshIngress) GetInstallationNamespace() string {
	if m != nil {
		return m.InstallationNamespace
	}
	return ""
}

func init() {
	proto.RegisterType((*MeshIngress)(nil), "supergloo.solo.io.MeshIngress")
	proto.RegisterType((*GlooMeshIngress)(nil), "supergloo.solo.io.GlooMeshIngress")
}

func init() {
	proto.RegisterFile("github.com/solo-io/supergloo/api/v1/ingress.proto", fileDescriptor_5d220723bdc29937)
}

var fileDescriptor_5d220723bdc29937 = []byte{
	// 382 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0x41, 0x6b, 0xe2, 0x40,
	0x14, 0xc7, 0x55, 0x82, 0xec, 0x8e, 0x2c, 0x62, 0xd6, 0x95, 0xe8, 0xc2, 0xba, 0xe4, 0xb2, 0x7b,
	0xa8, 0x13, 0x6c, 0x29, 0x88, 0xc7, 0x5c, 0xb4, 0x07, 0x7b, 0x48, 0x6f, 0xbd, 0xc8, 0x18, 0x5f,
	0xe2, 0x60, 0x92, 0x17, 0x32, 0x93, 0x42, 0xaf, 0x7e, 0x9a, 0x42, 0xbf, 0x48, 0x3f, 0x85, 0x87,
	0x7e, 0x03, 0xfb, 0x09, 0x4a, 0x26, 0x89, 0x68, 0x5b, 0x8a, 0x3d, 0x25, 0x79, 0xff, 0xf7, 0xfb,
	0xcf, 0x3f, 0xef, 0x0d, 0x19, 0xfa, 0x5c, 0xae, 0xd2, 0x05, 0x75, 0x31, 0xb4, 0x04, 0x06, 0x38,
	0xe0, 0x68, 0x89, 0x34, 0x86, 0xc4, 0x0f, 0x10, 0x2d, 0x16, 0x73, 0xeb, 0x6e, 0x68, 0xf1, 0xc8,
	0x4f, 0x40, 0x08, 0x1a, 0x27, 0x28, 0x51, 0x6f, 0xed, 0x75, 0x9a, 0x11, 0x94, 0x63, 0xaf, 0xed,
	0xa3, 0x8f, 0x4a, 0xb5, 0xb2, 0xb7, 0xbc, 0xb1, 0xf7, 0xa1, 0x77, 0xf6, 0x5c, 0x73, 0x59, 0x5a,
	0x87, 0x20, 0xd9, 0x92, 0x49, 0x56, 0x20, 0xd6, 0x09, 0x88, 0x90, 0x4c, 0xa6, 0x45, 0x98, 0xde,
	0xd9, 0x09, 0x40, 0x02, 0xde, 0x17, 0x12, 0x95, 0xdf, 0x39, 0x62, 0x3e, 0xd6, 0x48, 0x63, 0x06,
	0x62, 0x75, 0x95, 0xcf, 0x40, 0x9f, 0x90, 0x7a, 0x1e, 0xc0, 0x58, 0xfe, 0xad, 0xfe, 0x6f, 0x9c,
	0xb7, 0xa9, 0x8b, 0x09, 0x94, 0x93, 0xa0, 0x37, 0x4a, 0xb3, 0xbb, 0x4f, 0xdb, 0x7e, 0xe5, 0x65,
	0xdb, 0x6f, 0x49, 0x10, 0x72, 0xc9, 0x3d, 0x6f, 0x6c, 0x72, 0x3f, 0xc2, 0x04, 0x4c, 0xa7, 0xc0,
	0xf5, 0x11, 0xf9, 0x56, 0xfe, 0xbc, 0x01, 0xca, 0xaa, 0x73, 0x6c, 0x35, 0x2b, 0x54, 0x5b, 0xcb,
	0xcc, 0x9c, 0x7d, 0xb7, 0x3e, 0x22, 0x5a, 0x36, 0x7d, 0xa3, 0xaa, 0x28, 0x93, 0xbe, 0xdb, 0x07,
	0x9d, 0x04, 0x88, 0x07, 0xa1, 0xa7, 0x15, 0x47, 0x11, 0xfa, 0x80, 0x68, 0x21, 0x88, 0x95, 0x51,
	0x53, 0x64, 0xf7, 0xf8, 0x3c, 0x07, 0x04, 0xa6, 0x89, 0x0b, 0x0e, 0x78, 0x8e, 0x6a, 0x1b, 0xff,
	0xde, 0xec, 0x34, 0x8d, 0xd4, 0x42, 0xbe, 0xd9, 0x69, 0x4d, 0xfd, 0x47, 0x56, 0x2b, 0xae, 0x02,
	0x08, 0xfb, 0x27, 0x69, 0x65, 0x85, 0x79, 0x51, 0x99, 0xcb, 0xfb, 0x18, 0xcc, 0x29, 0x69, 0xbe,
	0x39, 0x5b, 0xbf, 0x24, 0x1d, 0x1e, 0x09, 0xc9, 0x82, 0x80, 0x49, 0x8e, 0xd1, 0x3c, 0x62, 0x21,
	0x88, 0x98, 0xb9, 0xa0, 0xf2, 0x7f, 0x77, 0x7e, 0x1d, 0xaa, 0xd7, 0xa5, 0x68, 0x0f, 0x1e, 0x9e,
	0xff, 0x54, 0x6f, 0xff, 0x7d, 0x7a, 0x3d, 0xe3, 0xb5, 0x5f, 0x6c, 0x6d, 0x51, 0x57, 0xdb, 0xba,
	0x78, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x8d, 0x2a, 0xf5, 0x26, 0xd0, 0x02, 0x00, 0x00,
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
	if !this.Mesh.Equal(that1.Mesh) {
		return false
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
	if this.InstallationNamespace != that1.InstallationNamespace {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
