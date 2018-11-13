// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: plugins/kubernetes/kubernetes.proto

package kubernetes // import "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/kubernetes"

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

// Upstream Spec for Kubernetes Upstreams
// Kubernetes Upstreams represent a set of one or more addressable pods for a Kubernetes Service
// the Gloo Kubernetes Upstream maps to a single service port. Because Kubernetes Services support mulitple ports,
// Gloo requires that a different upstream be created for each port
// Kubernetes Upstreams are typically generated automatically by Gloo from the Kubernetes API
type UpstreamSpec struct {
	// The name of the Kubernetes Service
	ServiceName string `protobuf:"bytes,1,opt,name=service_name,json=serviceName,proto3" json:"service_name,omitempty"`
	// The namespace where the Service lives
	ServiceNamespace string `protobuf:"bytes,2,opt,name=service_namespace,json=serviceNamespace,proto3" json:"service_namespace,omitempty"`
	// The port where the Service is listening.
	ServicePort uint32 `protobuf:"varint,3,opt,name=service_port,json=servicePort,proto3" json:"service_port,omitempty"`
	// Allows finer-grained filtering of pods for the Upstream. Gloo will select pods based on their labels if
	// any are provided here.
	// (see [Kubernetes labels and selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
	Selector map[string]string `protobuf:"bytes,4,rep,name=selector" json:"selector,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	//     An optional Service Spec describing the service listening at this address
	ServiceSpec          *plugins.ServiceSpec `protobuf:"bytes,5,opt,name=service_spec,json=serviceSpec" json:"service_spec,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *UpstreamSpec) Reset()         { *m = UpstreamSpec{} }
func (m *UpstreamSpec) String() string { return proto.CompactTextString(m) }
func (*UpstreamSpec) ProtoMessage()    {}
func (*UpstreamSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_kubernetes_bb8dd296f2ebcea3, []int{0}
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

func (m *UpstreamSpec) GetServiceName() string {
	if m != nil {
		return m.ServiceName
	}
	return ""
}

func (m *UpstreamSpec) GetServiceNamespace() string {
	if m != nil {
		return m.ServiceNamespace
	}
	return ""
}

func (m *UpstreamSpec) GetServicePort() uint32 {
	if m != nil {
		return m.ServicePort
	}
	return 0
}

func (m *UpstreamSpec) GetSelector() map[string]string {
	if m != nil {
		return m.Selector
	}
	return nil
}

func (m *UpstreamSpec) GetServiceSpec() *plugins.ServiceSpec {
	if m != nil {
		return m.ServiceSpec
	}
	return nil
}

func init() {
	proto.RegisterType((*UpstreamSpec)(nil), "kubernetes.plugins.gloo.solo.io.UpstreamSpec")
	proto.RegisterMapType((map[string]string)(nil), "kubernetes.plugins.gloo.solo.io.UpstreamSpec.SelectorEntry")
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
	if this.ServiceName != that1.ServiceName {
		return false
	}
	if this.ServiceNamespace != that1.ServiceNamespace {
		return false
	}
	if this.ServicePort != that1.ServicePort {
		return false
	}
	if len(this.Selector) != len(that1.Selector) {
		return false
	}
	for i := range this.Selector {
		if this.Selector[i] != that1.Selector[i] {
			return false
		}
	}
	if !this.ServiceSpec.Equal(that1.ServiceSpec) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}

func init() {
	proto.RegisterFile("plugins/kubernetes/kubernetes.proto", fileDescriptor_kubernetes_bb8dd296f2ebcea3)
}

var fileDescriptor_kubernetes_bb8dd296f2ebcea3 = []byte{
	// 344 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0xc1, 0x4b, 0xc3, 0x30,
	0x14, 0xc6, 0xe9, 0xe6, 0x44, 0xb3, 0x0d, 0x66, 0xd9, 0xa1, 0xec, 0xa0, 0x9d, 0x5e, 0x0a, 0x62,
	0x8a, 0xf3, 0x22, 0xee, 0x26, 0x0a, 0x9e, 0xc6, 0xd8, 0x10, 0xc1, 0x8b, 0x64, 0xe1, 0x51, 0xeb,
	0xda, 0xbe, 0x90, 0xa4, 0xc3, 0xfd, 0x47, 0xfe, 0x53, 0x5e, 0xfc, 0x4b, 0x24, 0xcd, 0xd6, 0x45,
	0x18, 0x78, 0xea, 0x97, 0xc7, 0x2f, 0x5f, 0x5e, 0xbf, 0x8f, 0x5c, 0x88, 0xac, 0x4c, 0xd2, 0x42,
	0xc5, 0xcb, 0x72, 0x01, 0xb2, 0x00, 0x0d, 0xae, 0xa4, 0x42, 0xa2, 0x46, 0xff, 0xcc, 0x9d, 0x58,
	0x9e, 0x26, 0x19, 0x22, 0x55, 0x98, 0x21, 0x4d, 0x71, 0xd0, 0x4f, 0x30, 0xc1, 0x8a, 0x8d, 0x8d,
	0xb2, 0xd7, 0x06, 0xd3, 0x24, 0xd5, 0xef, 0xe5, 0x82, 0x72, 0xcc, 0x63, 0x43, 0x5e, 0xa5, 0x68,
	0xbf, 0x42, 0xe2, 0x07, 0x70, 0xad, 0xe2, 0x5a, 0x18, 0xb7, 0x98, 0x89, 0x34, 0x5e, 0x5d, 0xc7,
	0xdb, 0x8d, 0x14, 0xc8, 0x55, 0xca, 0xe1, 0x4d, 0x09, 0xe0, 0xd6, 0xf1, 0xfc, 0xbb, 0x41, 0x3a,
	0xcf, 0x42, 0x69, 0x09, 0x2c, 0x9f, 0x0b, 0xe0, 0xfe, 0x90, 0x74, 0xb6, 0x58, 0xc1, 0x72, 0x08,
	0xbc, 0xd0, 0x8b, 0x8e, 0x67, 0xed, 0xcd, 0x6c, 0xc2, 0x72, 0xf0, 0x2f, 0xc9, 0x89, 0x8b, 0x28,
	0xc1, 0x38, 0x04, 0x8d, 0x8a, 0xeb, 0x39, 0x5c, 0x35, 0x77, 0xfd, 0x04, 0x4a, 0x1d, 0x34, 0x43,
	0x2f, 0xea, 0xd6, 0x7e, 0x53, 0x94, 0xda, 0x7f, 0x21, 0x47, 0x0a, 0x32, 0xe0, 0x1a, 0x65, 0x70,
	0x10, 0x36, 0xa3, 0xf6, 0x68, 0x4c, 0xff, 0xc9, 0x87, 0xba, 0x3b, 0xd3, 0xf9, 0xe6, 0xf6, 0x63,
	0xa1, 0xe5, 0x7a, 0x56, 0x9b, 0xf9, 0x0f, 0xbb, 0xb7, 0xcd, 0x2f, 0x07, 0xad, 0xd0, 0x8b, 0xda,
	0xa3, 0xe1, 0x7e, 0xc7, 0xb9, 0x25, 0x8d, 0x61, 0xbd, 0x9e, 0x39, 0x0c, 0xc6, 0xa4, 0xfb, 0xe7,
	0x01, 0xbf, 0x47, 0x9a, 0x4b, 0x58, 0x6f, 0x92, 0x31, 0xd2, 0xef, 0x93, 0xd6, 0x8a, 0x65, 0xe5,
	0x36, 0x05, 0x7b, 0xb8, 0x6b, 0xdc, 0x7a, 0xf7, 0x93, 0xaf, 0x9f, 0x53, 0xef, 0xf5, 0x69, 0x5f,
	0x6f, 0xa5, 0x00, 0x59, 0xb5, 0x24, 0x96, 0x49, 0xd5, 0x14, 0x7c, 0x6a, 0x90, 0x05, 0xcb, 0x6c,
	0x77, 0x4e, 0x6f, 0xbb, 0x30, 0x16, 0x87, 0x55, 0x6d, 0x37, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff,
	0xda, 0x4a, 0x99, 0xa2, 0x66, 0x02, 0x00, 0x00,
}
