// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: plugins/static/static.proto

package static // import "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/static"

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"
import plugins "github.com/solo-io/solo-projects/projects/gloo/pkg/api/v1/plugins"

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

// Static upstreams are used to route request to services listening at fixed IP/Addresses.
// Static upstreams can be used to proxy any kind of service, and therefore contain a ServiceSpec
// for additional service-specific configuration.
// Unlike upstreams created by service discovery, Static Upstreams must be created manually by users
type UpstreamSpec struct {
	// A list of addresses and ports
	// at least one must be specified
	Hosts []*Host `protobuf:"bytes,1,rep,name=hosts" json:"hosts,omitempty"`
	// Attempt to use outbound TLS
	// Gloo will automatically set this to true for port 443
	UseTls bool `protobuf:"varint,3,opt,name=use_tls,json=useTls,proto3" json:"use_tls,omitempty"`
	// An optional Service Spec describing the service listening at this address
	ServiceSpec          *plugins.ServiceSpec `protobuf:"bytes,5,opt,name=service_spec,json=serviceSpec" json:"service_spec,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *UpstreamSpec) Reset()         { *m = UpstreamSpec{} }
func (m *UpstreamSpec) String() string { return proto.CompactTextString(m) }
func (*UpstreamSpec) ProtoMessage()    {}
func (*UpstreamSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_static_c4264f0b256167be, []int{0}
}
func (m *UpstreamSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpstreamSpec.Unmarshal(m, b)
}
func (m *UpstreamSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpstreamSpec.Marshal(b, m, deterministic)
}
func (dst *UpstreamSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpstreamSpec.Merge(dst, src)
}
func (m *UpstreamSpec) XXX_Size() int {
	return xxx_messageInfo_UpstreamSpec.Size(m)
}
func (m *UpstreamSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_UpstreamSpec.DiscardUnknown(m)
}

var xxx_messageInfo_UpstreamSpec proto.InternalMessageInfo

func (m *UpstreamSpec) GetHosts() []*Host {
	if m != nil {
		return m.Hosts
	}
	return nil
}

func (m *UpstreamSpec) GetUseTls() bool {
	if m != nil {
		return m.UseTls
	}
	return false
}

func (m *UpstreamSpec) GetServiceSpec() *plugins.ServiceSpec {
	if m != nil {
		return m.ServiceSpec
	}
	return nil
}

// Represents a single instance of an upstream
type Host struct {
	// Address (hostname or IP)
	Addr string `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	// Port the instance is listening on
	Port                 uint32   `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Host) Reset()         { *m = Host{} }
func (m *Host) String() string { return proto.CompactTextString(m) }
func (*Host) ProtoMessage()    {}
func (*Host) Descriptor() ([]byte, []int) {
	return fileDescriptor_static_c4264f0b256167be, []int{1}
}
func (m *Host) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Host.Unmarshal(m, b)
}
func (m *Host) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Host.Marshal(b, m, deterministic)
}
func (dst *Host) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Host.Merge(dst, src)
}
func (m *Host) XXX_Size() int {
	return xxx_messageInfo_Host.Size(m)
}
func (m *Host) XXX_DiscardUnknown() {
	xxx_messageInfo_Host.DiscardUnknown(m)
}

var xxx_messageInfo_Host proto.InternalMessageInfo

func (m *Host) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *Host) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func init() {
	proto.RegisterType((*UpstreamSpec)(nil), "static.plugins.gloo.solo.io.UpstreamSpec")
	proto.RegisterType((*Host)(nil), "static.plugins.gloo.solo.io.Host")
}
func (this *UpstreamSpec) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*UpstreamSpec)
	if !ok {
		that2, ok := that.(UpstreamSpec)
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
	if len(this.Hosts) != len(that1.Hosts) {
		return false
	}
	for i := range this.Hosts {
		if !this.Hosts[i].Equal(that1.Hosts[i]) {
			return false
		}
	}
	if this.UseTls != that1.UseTls {
		return false
	}
	if !this.ServiceSpec.Equal(that1.ServiceSpec) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *Host) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Host)
	if !ok {
		that2, ok := that.(Host)
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
	if this.Addr != that1.Addr {
		return false
	}
	if this.Port != that1.Port {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}

func init() { proto.RegisterFile("plugins/static/static.proto", fileDescriptor_static_c4264f0b256167be) }

var fileDescriptor_static_c4264f0b256167be = []byte{
	// 295 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x90, 0x4f, 0x4f, 0x02, 0x31,
	0x10, 0xc5, 0x53, 0xf9, 0xa3, 0x16, 0xbc, 0x34, 0x26, 0x6e, 0x20, 0x31, 0x2b, 0xa7, 0xbd, 0xd8,
	0x46, 0x3d, 0x78, 0x37, 0x1c, 0x8c, 0x27, 0xb3, 0xe8, 0xc5, 0x0b, 0x59, 0x96, 0xa6, 0x54, 0x0b,
	0xd3, 0x74, 0xba, 0xc4, 0x4f, 0x64, 0xfc, 0x5c, 0x7e, 0x12, 0xd3, 0x16, 0x91, 0x03, 0xe1, 0x34,
	0x6f, 0x36, 0x6f, 0xdf, 0xbc, 0xfe, 0xe8, 0xd0, 0x9a, 0x46, 0xe9, 0x15, 0x0a, 0xf4, 0x95, 0xd7,
	0xf5, 0x66, 0x70, 0xeb, 0xc0, 0x03, 0x1b, 0xfe, 0x6d, 0xc9, 0xc3, 0x95, 0x01, 0xe0, 0x08, 0x06,
	0xb8, 0x86, 0xc1, 0xb9, 0x02, 0x05, 0xd1, 0x27, 0x82, 0x4a, 0xbf, 0x0c, 0x9e, 0x95, 0xf6, 0x8b,
	0x66, 0xc6, 0x6b, 0x58, 0x8a, 0xe0, 0xbc, 0xd6, 0x90, 0xa6, 0x75, 0xf0, 0x2e, 0x6b, 0x8f, 0x62,
	0x2b, 0x42, 0x9a, 0xa8, 0xac, 0x16, 0xeb, 0x1b, 0xb1, 0x6d, 0x21, 0xdd, 0x5a, 0xd7, 0x72, 0x8a,
	0x56, 0x6e, 0x4a, 0x8c, 0xbe, 0x08, 0xed, 0xbf, 0x5a, 0xf4, 0x4e, 0x56, 0xcb, 0x89, 0x95, 0x35,
	0xbb, 0xa7, 0x9d, 0x05, 0xa0, 0xc7, 0x8c, 0xe4, 0xad, 0xa2, 0x77, 0x7b, 0xc5, 0x0f, 0xb4, 0xe4,
	0x8f, 0x80, 0xbe, 0x4c, 0x7e, 0x76, 0x41, 0x8f, 0x1b, 0x94, 0x53, 0x6f, 0x30, 0x6b, 0xe5, 0xa4,
	0x38, 0x29, 0xbb, 0x0d, 0xca, 0x17, 0x83, 0x6c, 0x4c, 0xfb, 0xbb, 0x87, 0xb3, 0x4e, 0x4e, 0x62,
	0xf0, 0xde, 0xc4, 0x49, 0x72, 0x86, 0x2a, 0x65, 0x0f, 0xff, 0x97, 0x11, 0xa7, 0xed, 0x70, 0x8d,
	0x31, 0xda, 0xae, 0xe6, 0x73, 0x97, 0x91, 0x9c, 0x14, 0xa7, 0x65, 0xd4, 0xe1, 0x9b, 0x05, 0xe7,
	0xb3, 0xa3, 0x9c, 0x14, 0x67, 0x65, 0xd4, 0x0f, 0x4f, 0xdf, 0x3f, 0x97, 0xe4, 0x6d, 0xbc, 0x0f,
	0x58, 0x63, 0xa5, 0x8b, 0x78, 0xec, 0x87, 0x8a, 0x88, 0xe4, 0xa7, 0x97, 0x6e, 0x55, 0x99, 0x04,
	0x6d, 0x17, 0x58, 0x7c, 0xfb, 0xac, 0x1b, 0x59, 0xdd, 0xfd, 0x06, 0x00, 0x00, 0xff, 0xff, 0xb0,
	0xb3, 0x12, 0x6c, 0xcf, 0x01, 0x00, 0x00,
}
