// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/supergloo/api/v1/install.proto

package v1 // import "github.com/solo-io/supergloo/pkg/api/v1"

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
// Installs represent a desired installation of a supported mesh.
// Supergloo watches for installs and synchronizes the managed installations
// with the desired configuration in the install object.
//
// Updating the configuration of an install object will cause supergloo to
// modify the corresponding mesh.
type Install struct {
	// Status indicates the validation status of this resource.
	// Status is read-only by clients, and set by supergloo during validation
	Status core.Status `protobuf:"bytes,100,opt,name=status" json:"status" testdiff:"ignore"`
	// Metadata contains the object metadata for this resource
	Metadata core.Metadata `protobuf:"bytes,101,opt,name=metadata" json:"metadata"`
	// disables this install
	// setting this to true will cause supergloo not to
	// install this mesh, or uninstall an active install
	Disabled bool `protobuf:"varint,1,opt,name=disabled,proto3" json:"disabled,omitempty"`
	// The type of object the install handles
	// Currently support types are mesh, and ingress
	//
	// Types that are valid to be assigned to InstallType:
	//	*Install_Mesh
	//	*Install_Ingress
	InstallType isInstall_InstallType `protobuf_oneof:"install_type"`
	// which namespace to install to
	InstallationNamespace string   `protobuf:"bytes,4,opt,name=installation_namespace,json=installationNamespace,proto3" json:"installation_namespace,omitempty"`
	XXX_NoUnkeyedLiteral  struct{} `json:"-"`
	XXX_unrecognized      []byte   `json:"-"`
	XXX_sizecache         int32    `json:"-"`
}

func (m *Install) Reset()         { *m = Install{} }
func (m *Install) String() string { return proto.CompactTextString(m) }
func (*Install) ProtoMessage()    {}
func (*Install) Descriptor() ([]byte, []int) {
	return fileDescriptor_install_a1c3eba2fde794f0, []int{0}
}
func (m *Install) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Install.Unmarshal(m, b)
}
func (m *Install) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Install.Marshal(b, m, deterministic)
}
func (dst *Install) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Install.Merge(dst, src)
}
func (m *Install) XXX_Size() int {
	return xxx_messageInfo_Install.Size(m)
}
func (m *Install) XXX_DiscardUnknown() {
	xxx_messageInfo_Install.DiscardUnknown(m)
}

var xxx_messageInfo_Install proto.InternalMessageInfo

type isInstall_InstallType interface {
	isInstall_InstallType()
	Equal(interface{}) bool
}

type Install_Mesh struct {
	Mesh *MeshInstall `protobuf:"bytes,2,opt,name=mesh,oneof"`
}
type Install_Ingress struct {
	Ingress *MeshIngressInstall `protobuf:"bytes,3,opt,name=ingress,oneof"`
}

func (*Install_Mesh) isInstall_InstallType()    {}
func (*Install_Ingress) isInstall_InstallType() {}

func (m *Install) GetInstallType() isInstall_InstallType {
	if m != nil {
		return m.InstallType
	}
	return nil
}

func (m *Install) GetStatus() core.Status {
	if m != nil {
		return m.Status
	}
	return core.Status{}
}

func (m *Install) GetMetadata() core.Metadata {
	if m != nil {
		return m.Metadata
	}
	return core.Metadata{}
}

func (m *Install) GetDisabled() bool {
	if m != nil {
		return m.Disabled
	}
	return false
}

func (m *Install) GetMesh() *MeshInstall {
	if x, ok := m.GetInstallType().(*Install_Mesh); ok {
		return x.Mesh
	}
	return nil
}

func (m *Install) GetIngress() *MeshIngressInstall {
	if x, ok := m.GetInstallType().(*Install_Ingress); ok {
		return x.Ingress
	}
	return nil
}

func (m *Install) GetInstallationNamespace() string {
	if m != nil {
		return m.InstallationNamespace
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Install) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Install_OneofMarshaler, _Install_OneofUnmarshaler, _Install_OneofSizer, []interface{}{
		(*Install_Mesh)(nil),
		(*Install_Ingress)(nil),
	}
}

func _Install_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Install)
	// install_type
	switch x := m.InstallType.(type) {
	case *Install_Mesh:
		_ = b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Mesh); err != nil {
			return err
		}
	case *Install_Ingress:
		_ = b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Ingress); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Install.InstallType has unexpected type %T", x)
	}
	return nil
}

func _Install_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Install)
	switch tag {
	case 2: // install_type.mesh
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(MeshInstall)
		err := b.DecodeMessage(msg)
		m.InstallType = &Install_Mesh{msg}
		return true, err
	case 3: // install_type.ingress
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(MeshIngressInstall)
		err := b.DecodeMessage(msg)
		m.InstallType = &Install_Ingress{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Install_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Install)
	// install_type
	switch x := m.InstallType.(type) {
	case *Install_Mesh:
		s := proto.Size(x.Mesh)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Install_Ingress:
		s := proto.Size(x.Ingress)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

//
// Generic container for mesh installs handled by supergloo
//
// Holds all configuration shared between different mesh types
type MeshInstall struct {
	// The type of mesh to install
	// currently only istio is supported
	//
	// Types that are valid to be assigned to MeshInstallType:
	//	*MeshInstall_Istio
	//	*MeshInstall_Linkerd
	MeshInstallType      isMeshInstall_MeshInstallType `protobuf_oneof:"mesh_install_type"`
	XXX_NoUnkeyedLiteral struct{}                      `json:"-"`
	XXX_unrecognized     []byte                        `json:"-"`
	XXX_sizecache        int32                         `json:"-"`
}

func (m *MeshInstall) Reset()         { *m = MeshInstall{} }
func (m *MeshInstall) String() string { return proto.CompactTextString(m) }
func (*MeshInstall) ProtoMessage()    {}
func (*MeshInstall) Descriptor() ([]byte, []int) {
	return fileDescriptor_install_a1c3eba2fde794f0, []int{1}
}
func (m *MeshInstall) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MeshInstall.Unmarshal(m, b)
}
func (m *MeshInstall) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MeshInstall.Marshal(b, m, deterministic)
}
func (dst *MeshInstall) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MeshInstall.Merge(dst, src)
}
func (m *MeshInstall) XXX_Size() int {
	return xxx_messageInfo_MeshInstall.Size(m)
}
func (m *MeshInstall) XXX_DiscardUnknown() {
	xxx_messageInfo_MeshInstall.DiscardUnknown(m)
}

var xxx_messageInfo_MeshInstall proto.InternalMessageInfo

type isMeshInstall_MeshInstallType interface {
	isMeshInstall_MeshInstallType()
	Equal(interface{}) bool
}

type MeshInstall_Istio struct {
	Istio *IstioInstall `protobuf:"bytes,2,opt,name=istio,oneof"`
}
type MeshInstall_Linkerd struct {
	Linkerd *LinkerdInstall `protobuf:"bytes,3,opt,name=linkerd,oneof"`
}

func (*MeshInstall_Istio) isMeshInstall_MeshInstallType()   {}
func (*MeshInstall_Linkerd) isMeshInstall_MeshInstallType() {}

func (m *MeshInstall) GetMeshInstallType() isMeshInstall_MeshInstallType {
	if m != nil {
		return m.MeshInstallType
	}
	return nil
}

func (m *MeshInstall) GetIstio() *IstioInstall {
	if x, ok := m.GetMeshInstallType().(*MeshInstall_Istio); ok {
		return x.Istio
	}
	return nil
}

func (m *MeshInstall) GetLinkerd() *LinkerdInstall {
	if x, ok := m.GetMeshInstallType().(*MeshInstall_Linkerd); ok {
		return x.Linkerd
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*MeshInstall) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _MeshInstall_OneofMarshaler, _MeshInstall_OneofUnmarshaler, _MeshInstall_OneofSizer, []interface{}{
		(*MeshInstall_Istio)(nil),
		(*MeshInstall_Linkerd)(nil),
	}
}

func _MeshInstall_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*MeshInstall)
	// mesh_install_type
	switch x := m.MeshInstallType.(type) {
	case *MeshInstall_Istio:
		_ = b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Istio); err != nil {
			return err
		}
	case *MeshInstall_Linkerd:
		_ = b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Linkerd); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("MeshInstall.MeshInstallType has unexpected type %T", x)
	}
	return nil
}

func _MeshInstall_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*MeshInstall)
	switch tag {
	case 2: // mesh_install_type.istio
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(IstioInstall)
		err := b.DecodeMessage(msg)
		m.MeshInstallType = &MeshInstall_Istio{msg}
		return true, err
	case 3: // mesh_install_type.linkerd
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(LinkerdInstall)
		err := b.DecodeMessage(msg)
		m.MeshInstallType = &MeshInstall_Linkerd{msg}
		return true, err
	default:
		return false, nil
	}
}

func _MeshInstall_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*MeshInstall)
	// mesh_install_type
	switch x := m.MeshInstallType.(type) {
	case *MeshInstall_Istio:
		s := proto.Size(x.Istio)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *MeshInstall_Linkerd:
		s := proto.Size(x.Linkerd)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Installation options for Istio
type IstioInstall struct {
	// which version of the istio helm chart to install
	Version string `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	// enable auto injection of pods
	EnableAutoInject bool `protobuf:"varint,3,opt,name=enable_auto_inject,json=enableAutoInject,proto3" json:"enable_auto_inject,omitempty"`
	// enable mutual tls between pods
	EnableMtls bool `protobuf:"varint,4,opt,name=enable_mtls,json=enableMtls,proto3" json:"enable_mtls,omitempty"`
	// optional. set to use a custom root ca
	// to issue certificates for mtls
	// ignored if mtls is disabled
	CustomRootCert *core.ResourceRef `protobuf:"bytes,9,opt,name=custom_root_cert,json=customRootCert" json:"custom_root_cert,omitempty"`
	// install grafana with istio
	InstallGrafana bool `protobuf:"varint,6,opt,name=install_grafana,json=installGrafana,proto3" json:"install_grafana,omitempty"`
	// install prometheus with istio
	InstallPrometheus bool `protobuf:"varint,7,opt,name=install_prometheus,json=installPrometheus,proto3" json:"install_prometheus,omitempty"`
	// install jaeger with istio
	InstallJaeger bool `protobuf:"varint,8,opt,name=install_jaeger,json=installJaeger,proto3" json:"install_jaeger,omitempty"`
	// enable ingress gateway
	EnableIngress bool `protobuf:"varint,10,opt,name=enable_ingress,json=enableIngress,proto3" json:"enable_ingress,omitempty"`
	// enable egress gateway
	EnableEgress         bool     `protobuf:"varint,11,opt,name=enable_egress,json=enableEgress,proto3" json:"enable_egress,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IstioInstall) Reset()         { *m = IstioInstall{} }
func (m *IstioInstall) String() string { return proto.CompactTextString(m) }
func (*IstioInstall) ProtoMessage()    {}
func (*IstioInstall) Descriptor() ([]byte, []int) {
	return fileDescriptor_install_a1c3eba2fde794f0, []int{2}
}
func (m *IstioInstall) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IstioInstall.Unmarshal(m, b)
}
func (m *IstioInstall) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IstioInstall.Marshal(b, m, deterministic)
}
func (dst *IstioInstall) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IstioInstall.Merge(dst, src)
}
func (m *IstioInstall) XXX_Size() int {
	return xxx_messageInfo_IstioInstall.Size(m)
}
func (m *IstioInstall) XXX_DiscardUnknown() {
	xxx_messageInfo_IstioInstall.DiscardUnknown(m)
}

var xxx_messageInfo_IstioInstall proto.InternalMessageInfo

func (m *IstioInstall) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *IstioInstall) GetEnableAutoInject() bool {
	if m != nil {
		return m.EnableAutoInject
	}
	return false
}

func (m *IstioInstall) GetEnableMtls() bool {
	if m != nil {
		return m.EnableMtls
	}
	return false
}

func (m *IstioInstall) GetCustomRootCert() *core.ResourceRef {
	if m != nil {
		return m.CustomRootCert
	}
	return nil
}

func (m *IstioInstall) GetInstallGrafana() bool {
	if m != nil {
		return m.InstallGrafana
	}
	return false
}

func (m *IstioInstall) GetInstallPrometheus() bool {
	if m != nil {
		return m.InstallPrometheus
	}
	return false
}

func (m *IstioInstall) GetInstallJaeger() bool {
	if m != nil {
		return m.InstallJaeger
	}
	return false
}

func (m *IstioInstall) GetEnableIngress() bool {
	if m != nil {
		return m.EnableIngress
	}
	return false
}

func (m *IstioInstall) GetEnableEgress() bool {
	if m != nil {
		return m.EnableEgress
	}
	return false
}

// Installation options for Linkerd
type LinkerdInstall struct {
	// which version of the Linkerd helm chart to install
	Version string `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	// enable auto injection of pods
	EnableAutoInject bool `protobuf:"varint,3,opt,name=enable_auto_inject,json=enableAutoInject,proto3" json:"enable_auto_inject,omitempty"`
	// enable mutual tls between pods
	EnableMtls           bool     `protobuf:"varint,4,opt,name=enable_mtls,json=enableMtls,proto3" json:"enable_mtls,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LinkerdInstall) Reset()         { *m = LinkerdInstall{} }
func (m *LinkerdInstall) String() string { return proto.CompactTextString(m) }
func (*LinkerdInstall) ProtoMessage()    {}
func (*LinkerdInstall) Descriptor() ([]byte, []int) {
	return fileDescriptor_install_a1c3eba2fde794f0, []int{3}
}
func (m *LinkerdInstall) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LinkerdInstall.Unmarshal(m, b)
}
func (m *LinkerdInstall) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LinkerdInstall.Marshal(b, m, deterministic)
}
func (dst *LinkerdInstall) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LinkerdInstall.Merge(dst, src)
}
func (m *LinkerdInstall) XXX_Size() int {
	return xxx_messageInfo_LinkerdInstall.Size(m)
}
func (m *LinkerdInstall) XXX_DiscardUnknown() {
	xxx_messageInfo_LinkerdInstall.DiscardUnknown(m)
}

var xxx_messageInfo_LinkerdInstall proto.InternalMessageInfo

func (m *LinkerdInstall) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *LinkerdInstall) GetEnableAutoInject() bool {
	if m != nil {
		return m.EnableAutoInject
	}
	return false
}

func (m *LinkerdInstall) GetEnableMtls() bool {
	if m != nil {
		return m.EnableMtls
	}
	return false
}

//
// Generic container for ingress installs handled by supergloo
//
// Holds all configuration shared between different ingress types
type MeshIngressInstall struct {
	// The type of mesh to install
	// currently only gloo is supported
	//
	// Types that are valid to be assigned to IngressInstallType:
	//	*MeshIngressInstall_Gloo
	IngressInstallType isMeshIngressInstall_IngressInstallType `protobuf_oneof:"ingress_install_type"`
	// reference to the Ingress crd that was created from this install
	// read-only, set by the server after successful installation.
	InstalledIngress     *core.ResourceRef `protobuf:"bytes,3,opt,name=installed_ingress,json=installedIngress" json:"installed_ingress,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *MeshIngressInstall) Reset()         { *m = MeshIngressInstall{} }
func (m *MeshIngressInstall) String() string { return proto.CompactTextString(m) }
func (*MeshIngressInstall) ProtoMessage()    {}
func (*MeshIngressInstall) Descriptor() ([]byte, []int) {
	return fileDescriptor_install_a1c3eba2fde794f0, []int{4}
}
func (m *MeshIngressInstall) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MeshIngressInstall.Unmarshal(m, b)
}
func (m *MeshIngressInstall) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MeshIngressInstall.Marshal(b, m, deterministic)
}
func (dst *MeshIngressInstall) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MeshIngressInstall.Merge(dst, src)
}
func (m *MeshIngressInstall) XXX_Size() int {
	return xxx_messageInfo_MeshIngressInstall.Size(m)
}
func (m *MeshIngressInstall) XXX_DiscardUnknown() {
	xxx_messageInfo_MeshIngressInstall.DiscardUnknown(m)
}

var xxx_messageInfo_MeshIngressInstall proto.InternalMessageInfo

type isMeshIngressInstall_IngressInstallType interface {
	isMeshIngressInstall_IngressInstallType()
	Equal(interface{}) bool
}

type MeshIngressInstall_Gloo struct {
	Gloo *GlooInstall `protobuf:"bytes,1,opt,name=gloo,oneof"`
}

func (*MeshIngressInstall_Gloo) isMeshIngressInstall_IngressInstallType() {}

func (m *MeshIngressInstall) GetIngressInstallType() isMeshIngressInstall_IngressInstallType {
	if m != nil {
		return m.IngressInstallType
	}
	return nil
}

func (m *MeshIngressInstall) GetGloo() *GlooInstall {
	if x, ok := m.GetIngressInstallType().(*MeshIngressInstall_Gloo); ok {
		return x.Gloo
	}
	return nil
}

func (m *MeshIngressInstall) GetInstalledIngress() *core.ResourceRef {
	if m != nil {
		return m.InstalledIngress
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*MeshIngressInstall) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _MeshIngressInstall_OneofMarshaler, _MeshIngressInstall_OneofUnmarshaler, _MeshIngressInstall_OneofSizer, []interface{}{
		(*MeshIngressInstall_Gloo)(nil),
	}
}

func _MeshIngressInstall_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*MeshIngressInstall)
	// ingress_install_type
	switch x := m.IngressInstallType.(type) {
	case *MeshIngressInstall_Gloo:
		_ = b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Gloo); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("MeshIngressInstall.IngressInstallType has unexpected type %T", x)
	}
	return nil
}

func _MeshIngressInstall_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*MeshIngressInstall)
	switch tag {
	case 1: // ingress_install_type.gloo
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(GlooInstall)
		err := b.DecodeMessage(msg)
		m.IngressInstallType = &MeshIngressInstall_Gloo{msg}
		return true, err
	default:
		return false, nil
	}
}

func _MeshIngressInstall_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*MeshIngressInstall)
	// ingress_install_type
	switch x := m.IngressInstallType.(type) {
	case *MeshIngressInstall_Gloo:
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

// Installation options for Gloo Ingress
type GlooInstall struct {
	// which version of the gloo helm chart to install
	// ignored if using custom helm chart
	Version string `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	// reference to the Mesh(s) that this ingress is acting upon
	Meshes               []*core.ResourceRef `protobuf:"bytes,3,rep,name=meshes" json:"meshes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *GlooInstall) Reset()         { *m = GlooInstall{} }
func (m *GlooInstall) String() string { return proto.CompactTextString(m) }
func (*GlooInstall) ProtoMessage()    {}
func (*GlooInstall) Descriptor() ([]byte, []int) {
	return fileDescriptor_install_a1c3eba2fde794f0, []int{5}
}
func (m *GlooInstall) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GlooInstall.Unmarshal(m, b)
}
func (m *GlooInstall) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GlooInstall.Marshal(b, m, deterministic)
}
func (dst *GlooInstall) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GlooInstall.Merge(dst, src)
}
func (m *GlooInstall) XXX_Size() int {
	return xxx_messageInfo_GlooInstall.Size(m)
}
func (m *GlooInstall) XXX_DiscardUnknown() {
	xxx_messageInfo_GlooInstall.DiscardUnknown(m)
}

var xxx_messageInfo_GlooInstall proto.InternalMessageInfo

func (m *GlooInstall) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *GlooInstall) GetMeshes() []*core.ResourceRef {
	if m != nil {
		return m.Meshes
	}
	return nil
}

func init() {
	proto.RegisterType((*Install)(nil), "supergloo.solo.io.Install")
	proto.RegisterType((*MeshInstall)(nil), "supergloo.solo.io.MeshInstall")
	proto.RegisterType((*IstioInstall)(nil), "supergloo.solo.io.IstioInstall")
	proto.RegisterType((*LinkerdInstall)(nil), "supergloo.solo.io.LinkerdInstall")
	proto.RegisterType((*MeshIngressInstall)(nil), "supergloo.solo.io.MeshIngressInstall")
	proto.RegisterType((*GlooInstall)(nil), "supergloo.solo.io.GlooInstall")
}
func (this *Install) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Install)
	if !ok {
		that2, ok := that.(Install)
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
	if this.Disabled != that1.Disabled {
		return false
	}
	if that1.InstallType == nil {
		if this.InstallType != nil {
			return false
		}
	} else if this.InstallType == nil {
		return false
	} else if !this.InstallType.Equal(that1.InstallType) {
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
func (this *Install_Mesh) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Install_Mesh)
	if !ok {
		that2, ok := that.(Install_Mesh)
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
	if !this.Mesh.Equal(that1.Mesh) {
		return false
	}
	return true
}
func (this *Install_Ingress) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Install_Ingress)
	if !ok {
		that2, ok := that.(Install_Ingress)
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
	if !this.Ingress.Equal(that1.Ingress) {
		return false
	}
	return true
}
func (this *MeshInstall) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MeshInstall)
	if !ok {
		that2, ok := that.(MeshInstall)
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
	if that1.MeshInstallType == nil {
		if this.MeshInstallType != nil {
			return false
		}
	} else if this.MeshInstallType == nil {
		return false
	} else if !this.MeshInstallType.Equal(that1.MeshInstallType) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *MeshInstall_Istio) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MeshInstall_Istio)
	if !ok {
		that2, ok := that.(MeshInstall_Istio)
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
	if !this.Istio.Equal(that1.Istio) {
		return false
	}
	return true
}
func (this *MeshInstall_Linkerd) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MeshInstall_Linkerd)
	if !ok {
		that2, ok := that.(MeshInstall_Linkerd)
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
	if !this.Linkerd.Equal(that1.Linkerd) {
		return false
	}
	return true
}
func (this *IstioInstall) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*IstioInstall)
	if !ok {
		that2, ok := that.(IstioInstall)
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
	if this.Version != that1.Version {
		return false
	}
	if this.EnableAutoInject != that1.EnableAutoInject {
		return false
	}
	if this.EnableMtls != that1.EnableMtls {
		return false
	}
	if !this.CustomRootCert.Equal(that1.CustomRootCert) {
		return false
	}
	if this.InstallGrafana != that1.InstallGrafana {
		return false
	}
	if this.InstallPrometheus != that1.InstallPrometheus {
		return false
	}
	if this.InstallJaeger != that1.InstallJaeger {
		return false
	}
	if this.EnableIngress != that1.EnableIngress {
		return false
	}
	if this.EnableEgress != that1.EnableEgress {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *LinkerdInstall) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*LinkerdInstall)
	if !ok {
		that2, ok := that.(LinkerdInstall)
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
	if this.Version != that1.Version {
		return false
	}
	if this.EnableAutoInject != that1.EnableAutoInject {
		return false
	}
	if this.EnableMtls != that1.EnableMtls {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *MeshIngressInstall) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MeshIngressInstall)
	if !ok {
		that2, ok := that.(MeshIngressInstall)
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
	if that1.IngressInstallType == nil {
		if this.IngressInstallType != nil {
			return false
		}
	} else if this.IngressInstallType == nil {
		return false
	} else if !this.IngressInstallType.Equal(that1.IngressInstallType) {
		return false
	}
	if !this.InstalledIngress.Equal(that1.InstalledIngress) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *MeshIngressInstall_Gloo) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MeshIngressInstall_Gloo)
	if !ok {
		that2, ok := that.(MeshIngressInstall_Gloo)
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
func (this *GlooInstall) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*GlooInstall)
	if !ok {
		that2, ok := that.(GlooInstall)
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
	if this.Version != that1.Version {
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

func init() {
	proto.RegisterFile("github.com/solo-io/supergloo/api/v1/install.proto", fileDescriptor_install_a1c3eba2fde794f0)
}

var fileDescriptor_install_a1c3eba2fde794f0 = []byte{
	// 702 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x55, 0x4d, 0x4e, 0xdb, 0x4e,
	0x14, 0xc7, 0x24, 0xff, 0x24, 0xbc, 0xf0, 0x4f, 0xc9, 0x14, 0x90, 0x61, 0x41, 0x68, 0x2a, 0x04,
	0x0b, 0x70, 0x44, 0x3f, 0xd4, 0x0a, 0xa9, 0x0b, 0x82, 0x5a, 0x4a, 0x55, 0xaa, 0xca, 0xdd, 0xb1,
	0xb1, 0x86, 0xe4, 0xc5, 0x19, 0x70, 0x3c, 0xd6, 0xcc, 0x18, 0x89, 0x2d, 0x9b, 0xde, 0xa1, 0x27,
	0xa8, 0x7a, 0x92, 0x9e, 0x82, 0x45, 0x6f, 0x40, 0x2f, 0xd0, 0xca, 0x33, 0xe3, 0x90, 0x88, 0x34,
	0xa2, 0x9b, 0xae, 0x62, 0xff, 0x3e, 0xde, 0xc7, 0xcc, 0x7b, 0x0e, 0xec, 0x86, 0x4c, 0xf5, 0xd3,
	0x53, 0xaf, 0xc3, 0x07, 0x2d, 0xc9, 0x23, 0xbe, 0xc3, 0x78, 0x4b, 0xa6, 0x09, 0x8a, 0x30, 0xe2,
	0xbc, 0x45, 0x13, 0xd6, 0xba, 0xd8, 0x6d, 0xb1, 0x58, 0x2a, 0x1a, 0x45, 0x5e, 0x22, 0xb8, 0xe2,
	0xa4, 0x3e, 0xe4, 0xbd, 0xcc, 0xe1, 0x31, 0xbe, 0xba, 0x18, 0xf2, 0x90, 0x6b, 0xb6, 0x95, 0x3d,
	0x19, 0xe1, 0xea, 0xc4, 0xd8, 0xd9, 0xef, 0x39, 0x53, 0x79, 0xe8, 0x01, 0x2a, 0xda, 0xa5, 0x8a,
	0x5a, 0x4b, 0xeb, 0x1e, 0x16, 0xa9, 0xa8, 0x4a, 0xa5, 0x35, 0x6c, 0xdf, 0xc3, 0x20, 0xb0, 0xf7,
	0x17, 0x15, 0xe5, 0xef, 0xc6, 0xd2, 0xfc, 0x35, 0x0b, 0xe5, 0x23, 0xd3, 0x3f, 0x39, 0x84, 0x92,
	0x49, 0xee, 0x76, 0xd7, 0x9d, 0xad, 0xea, 0x93, 0x45, 0xaf, 0xc3, 0x05, 0xe6, 0xa7, 0xe0, 0x7d,
	0xd2, 0x5c, 0x7b, 0xe5, 0xfb, 0x75, 0x63, 0xe6, 0xe7, 0x75, 0xa3, 0xae, 0x50, 0xaa, 0x2e, 0xeb,
	0xf5, 0xf6, 0x9a, 0x2c, 0x8c, 0xb9, 0xc0, 0xa6, 0x6f, 0xed, 0xe4, 0x25, 0x54, 0xf2, 0xc6, 0x5d,
	0xd4, 0xa1, 0x96, 0xc7, 0x43, 0x1d, 0x5b, 0xb6, 0x5d, 0xcc, 0x82, 0xf9, 0x43, 0x35, 0x59, 0x85,
	0x4a, 0x97, 0x49, 0x7a, 0x1a, 0x61, 0xd7, 0x75, 0xd6, 0x9d, 0xad, 0x8a, 0x3f, 0x7c, 0x27, 0xcf,
	0xa0, 0x38, 0x40, 0xd9, 0x77, 0x67, 0x75, 0xc4, 0x35, 0xef, 0xce, 0x3d, 0x79, 0xc7, 0x28, 0xfb,
	0xb6, 0x99, 0xb7, 0x33, 0xbe, 0x56, 0x93, 0x7d, 0x28, 0xb3, 0x38, 0x14, 0x28, 0xa5, 0x5b, 0xd0,
	0xc6, 0x8d, 0x3f, 0x1a, 0xb5, 0xea, 0xd6, 0x9f, 0xfb, 0xc8, 0x73, 0x58, 0xb6, 0x23, 0x42, 0x15,
	0xe3, 0x71, 0x10, 0xd3, 0x01, 0xca, 0x84, 0x76, 0xd0, 0x2d, 0xae, 0x3b, 0x5b, 0x73, 0xfe, 0xd2,
	0x28, 0xfb, 0x21, 0x27, 0xf7, 0x96, 0xae, 0x6e, 0x8a, 0x05, 0x70, 0xd8, 0xd5, 0x4d, 0x11, 0x48,
	0xc5, 0x6a, 0x64, 0xbb, 0x06, 0xf3, 0xf6, 0x39, 0x50, 0x97, 0x09, 0x36, 0xbf, 0x38, 0x50, 0x1d,
	0x29, 0x9c, 0xbc, 0x80, 0xff, 0x98, 0x54, 0x8c, 0xdb, 0x3e, 0x1b, 0x13, 0xca, 0x3d, 0xca, 0xf8,
	0xdb, 0x42, 0x8d, 0x9e, 0xbc, 0x82, 0x72, 0xc4, 0xe2, 0x73, 0x14, 0x5d, 0xdb, 0xe9, 0xa3, 0x09,
	0xd6, 0xf7, 0x46, 0x31, 0xd2, 0xa5, 0xf5, 0xb4, 0x1f, 0x42, 0x3d, 0x3b, 0xb0, 0x60, 0xac, 0xb8,
	0xcf, 0x05, 0x98, 0x1f, 0xcd, 0x46, 0x5c, 0x28, 0x5f, 0xa0, 0x90, 0x8c, 0xc7, 0xba, 0xbe, 0x39,
	0x3f, 0x7f, 0x25, 0xdb, 0x40, 0x30, 0xce, 0x6e, 0x2a, 0xa0, 0xa9, 0xe2, 0x01, 0x8b, 0xcf, 0xb0,
	0xa3, 0x74, 0x25, 0x15, 0x7f, 0xc1, 0x30, 0xfb, 0xa9, 0xe2, 0x47, 0x1a, 0x27, 0x0d, 0xa8, 0x5a,
	0xf5, 0x40, 0x45, 0x52, 0x1f, 0x64, 0xc5, 0x07, 0x03, 0x1d, 0xab, 0x48, 0x92, 0x03, 0x58, 0xe8,
	0xa4, 0x52, 0xf1, 0x41, 0x20, 0x38, 0x57, 0x41, 0x07, 0x85, 0x72, 0xe7, 0x74, 0x5b, 0x2b, 0xe3,
	0xb3, 0xe4, 0xa3, 0xe4, 0xa9, 0xe8, 0xa0, 0x8f, 0x3d, 0xbf, 0x66, 0x2c, 0x3e, 0xe7, 0xea, 0x00,
	0x85, 0x22, 0x9b, 0xf0, 0x20, 0x6f, 0x27, 0x14, 0xb4, 0x47, 0x63, 0xea, 0x96, 0x74, 0xa6, 0x9a,
	0x85, 0x0f, 0x0d, 0x4a, 0x76, 0x80, 0xe4, 0xc2, 0x44, 0xf0, 0x01, 0xaa, 0x3e, 0xa6, 0xd2, 0x2d,
	0x6b, 0x6d, 0xdd, 0x32, 0x1f, 0x87, 0x04, 0xd9, 0x80, 0x3c, 0x40, 0x70, 0x46, 0x31, 0x44, 0xe1,
	0x56, 0xb4, 0xf4, 0x7f, 0x8b, 0xbe, 0xd3, 0x60, 0x26, 0xb3, 0x4d, 0xe6, 0x23, 0x08, 0x46, 0x66,
	0x50, 0x3b, 0x71, 0xe4, 0x31, 0x58, 0x20, 0x40, 0xa3, 0xaa, 0x6a, 0xd5, 0xbc, 0x01, 0x5f, 0x6b,
	0xac, 0x79, 0x09, 0xb5, 0xf1, 0xbb, 0xfb, 0x67, 0x57, 0xd1, 0xfc, 0xe6, 0x00, 0xb9, 0xbb, 0x21,
	0xd9, 0x3e, 0x66, 0xa3, 0xa5, 0xf7, 0x74, 0xf2, 0x3e, 0x1e, 0x46, 0x7c, 0x64, 0x4c, 0xb5, 0x9a,
	0xbc, 0x81, 0xfc, 0x3c, 0xb1, 0x1b, 0x8c, 0x6f, 0xe6, 0x94, 0x8b, 0x5d, 0x18, 0x7a, 0x6c, 0x11,
	0xed, 0x65, 0x58, 0xb4, 0xee, 0xf1, 0x89, 0x3d, 0x81, 0xea, 0x48, 0xda, 0x29, 0x87, 0xb4, 0x0b,
	0xa5, 0x6c, 0xde, 0x31, 0xcb, 0x5e, 0x98, 0x9e, 0xdd, 0x0a, 0xdb, 0x3b, 0x5f, 0x7f, 0xac, 0x39,
	0x27, 0x9b, 0x53, 0xff, 0x53, 0x92, 0xf3, 0xd0, 0x7e, 0x6a, 0x4f, 0x4b, 0xfa, 0x13, 0xfb, 0xf4,
	0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0x5e, 0xbe, 0x3e, 0x75, 0x85, 0x06, 0x00, 0x00,
}
