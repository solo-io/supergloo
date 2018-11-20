// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: plugins.proto

package v1 // import "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"
import aws "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/aws"
import azure "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/azure"
import consul "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/consul"
import faultinjection "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/faultinjection"
import grpc1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/grpc"
import kubernetes "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/kubernetes"
import ratelimit "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/ratelimit"
import rest "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/rest"
import sqoop "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/sqoop"
import static "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/static"
import transformation "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/transformation"

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

// Plugin-specific configuration that lives on listeners
// Each ListenerPlugin object contains configuration for a specific plugin
// Note to developers: new Listener Plugins must be added to this struct
// to be usable by Gloo.
type ListenerPlugins struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListenerPlugins) Reset()         { *m = ListenerPlugins{} }
func (m *ListenerPlugins) String() string { return proto.CompactTextString(m) }
func (*ListenerPlugins) ProtoMessage()    {}
func (*ListenerPlugins) Descriptor() ([]byte, []int) {
	return fileDescriptor_plugins_f4f4de2ad43b6712, []int{0}
}
func (m *ListenerPlugins) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListenerPlugins.Unmarshal(m, b)
}
func (m *ListenerPlugins) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListenerPlugins.Marshal(b, m, deterministic)
}
func (dst *ListenerPlugins) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListenerPlugins.Merge(dst, src)
}
func (m *ListenerPlugins) XXX_Size() int {
	return xxx_messageInfo_ListenerPlugins.Size(m)
}
func (m *ListenerPlugins) XXX_DiscardUnknown() {
	xxx_messageInfo_ListenerPlugins.DiscardUnknown(m)
}

var xxx_messageInfo_ListenerPlugins proto.InternalMessageInfo

// Plugin-specific configuration that lives on virtual hosts
// Each VirtualHostPlugin object contains configuration for a specific plugin
// Note to developers: new Virtual Host Plugins must be added to this struct
// to be usable by Gloo.
type VirtualHostPlugins struct {
	RateLimits           *ratelimit.IngressRateLimit `protobuf:"bytes,1,opt,name=rate_limits,json=rateLimits" json:"rate_limits,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                    `json:"-"`
	XXX_unrecognized     []byte                      `json:"-"`
	XXX_sizecache        int32                       `json:"-"`
}

func (m *VirtualHostPlugins) Reset()         { *m = VirtualHostPlugins{} }
func (m *VirtualHostPlugins) String() string { return proto.CompactTextString(m) }
func (*VirtualHostPlugins) ProtoMessage()    {}
func (*VirtualHostPlugins) Descriptor() ([]byte, []int) {
	return fileDescriptor_plugins_f4f4de2ad43b6712, []int{1}
}
func (m *VirtualHostPlugins) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VirtualHostPlugins.Unmarshal(m, b)
}
func (m *VirtualHostPlugins) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VirtualHostPlugins.Marshal(b, m, deterministic)
}
func (dst *VirtualHostPlugins) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VirtualHostPlugins.Merge(dst, src)
}
func (m *VirtualHostPlugins) XXX_Size() int {
	return xxx_messageInfo_VirtualHostPlugins.Size(m)
}
func (m *VirtualHostPlugins) XXX_DiscardUnknown() {
	xxx_messageInfo_VirtualHostPlugins.DiscardUnknown(m)
}

var xxx_messageInfo_VirtualHostPlugins proto.InternalMessageInfo

func (m *VirtualHostPlugins) GetRateLimits() *ratelimit.IngressRateLimit {
	if m != nil {
		return m.RateLimits
	}
	return nil
}

// Plugin-specific configuration that lives on routes
// Each RoutePlugin object contains configuration for a specific plugin
// Note to developers: new Route Plugins must be added to this struct
// to be usable by Gloo.
type RoutePlugins struct {
	Transformations      *transformation.RouteTransformations `protobuf:"bytes,1,opt,name=transformations" json:"transformations,omitempty"`
	Faults               *faultinjection.RouteFaults          `protobuf:"bytes,2,opt,name=faults" json:"faults,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                             `json:"-"`
	XXX_unrecognized     []byte                               `json:"-"`
	XXX_sizecache        int32                                `json:"-"`
}

func (m *RoutePlugins) Reset()         { *m = RoutePlugins{} }
func (m *RoutePlugins) String() string { return proto.CompactTextString(m) }
func (*RoutePlugins) ProtoMessage()    {}
func (*RoutePlugins) Descriptor() ([]byte, []int) {
	return fileDescriptor_plugins_f4f4de2ad43b6712, []int{2}
}
func (m *RoutePlugins) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RoutePlugins.Unmarshal(m, b)
}
func (m *RoutePlugins) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RoutePlugins.Marshal(b, m, deterministic)
}
func (dst *RoutePlugins) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RoutePlugins.Merge(dst, src)
}
func (m *RoutePlugins) XXX_Size() int {
	return xxx_messageInfo_RoutePlugins.Size(m)
}
func (m *RoutePlugins) XXX_DiscardUnknown() {
	xxx_messageInfo_RoutePlugins.DiscardUnknown(m)
}

var xxx_messageInfo_RoutePlugins proto.InternalMessageInfo

func (m *RoutePlugins) GetTransformations() *transformation.RouteTransformations {
	if m != nil {
		return m.Transformations
	}
	return nil
}

func (m *RoutePlugins) GetFaults() *faultinjection.RouteFaults {
	if m != nil {
		return m.Faults
	}
	return nil
}

// Configuration for Destinations that are tied to the UpstreamSpec or ServiceSpec on that destination
type DestinationSpec struct {
	// Note to developers: new DestinationSpecs must be added to this oneof field
	// to be usable by Gloo.
	//
	// Types that are valid to be assigned to DestinationType:
	//	*DestinationSpec_Aws
	//	*DestinationSpec_Azure
	//	*DestinationSpec_Rest
	//	*DestinationSpec_Grpc
	//	*DestinationSpec_Sqoop
	DestinationType      isDestinationSpec_DestinationType `protobuf_oneof:"destination_type"`
	XXX_NoUnkeyedLiteral struct{}                          `json:"-"`
	XXX_unrecognized     []byte                            `json:"-"`
	XXX_sizecache        int32                             `json:"-"`
}

func (m *DestinationSpec) Reset()         { *m = DestinationSpec{} }
func (m *DestinationSpec) String() string { return proto.CompactTextString(m) }
func (*DestinationSpec) ProtoMessage()    {}
func (*DestinationSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_plugins_f4f4de2ad43b6712, []int{3}
}
func (m *DestinationSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DestinationSpec.Unmarshal(m, b)
}
func (m *DestinationSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DestinationSpec.Marshal(b, m, deterministic)
}
func (dst *DestinationSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DestinationSpec.Merge(dst, src)
}
func (m *DestinationSpec) XXX_Size() int {
	return xxx_messageInfo_DestinationSpec.Size(m)
}
func (m *DestinationSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_DestinationSpec.DiscardUnknown(m)
}

var xxx_messageInfo_DestinationSpec proto.InternalMessageInfo

type isDestinationSpec_DestinationType interface {
	isDestinationSpec_DestinationType()
	Equal(interface{}) bool
}

type DestinationSpec_Aws struct {
	Aws *aws.DestinationSpec `protobuf:"bytes,1,opt,name=aws,oneof"`
}
type DestinationSpec_Azure struct {
	Azure *azure.DestinationSpec `protobuf:"bytes,2,opt,name=azure,oneof"`
}
type DestinationSpec_Rest struct {
	Rest *rest.DestinationSpec `protobuf:"bytes,3,opt,name=rest,oneof"`
}
type DestinationSpec_Grpc struct {
	Grpc *grpc1.DestinationSpec `protobuf:"bytes,4,opt,name=grpc,oneof"`
}
type DestinationSpec_Sqoop struct {
	Sqoop *sqoop.DestinationSpec `protobuf:"bytes,5,opt,name=sqoop,oneof"`
}

func (*DestinationSpec_Aws) isDestinationSpec_DestinationType()   {}
func (*DestinationSpec_Azure) isDestinationSpec_DestinationType() {}
func (*DestinationSpec_Rest) isDestinationSpec_DestinationType()  {}
func (*DestinationSpec_Grpc) isDestinationSpec_DestinationType()  {}
func (*DestinationSpec_Sqoop) isDestinationSpec_DestinationType() {}

func (m *DestinationSpec) GetDestinationType() isDestinationSpec_DestinationType {
	if m != nil {
		return m.DestinationType
	}
	return nil
}

func (m *DestinationSpec) GetAws() *aws.DestinationSpec {
	if x, ok := m.GetDestinationType().(*DestinationSpec_Aws); ok {
		return x.Aws
	}
	return nil
}

func (m *DestinationSpec) GetAzure() *azure.DestinationSpec {
	if x, ok := m.GetDestinationType().(*DestinationSpec_Azure); ok {
		return x.Azure
	}
	return nil
}

func (m *DestinationSpec) GetRest() *rest.DestinationSpec {
	if x, ok := m.GetDestinationType().(*DestinationSpec_Rest); ok {
		return x.Rest
	}
	return nil
}

func (m *DestinationSpec) GetGrpc() *grpc1.DestinationSpec {
	if x, ok := m.GetDestinationType().(*DestinationSpec_Grpc); ok {
		return x.Grpc
	}
	return nil
}

func (m *DestinationSpec) GetSqoop() *sqoop.DestinationSpec {
	if x, ok := m.GetDestinationType().(*DestinationSpec_Sqoop); ok {
		return x.Sqoop
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*DestinationSpec) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _DestinationSpec_OneofMarshaler, _DestinationSpec_OneofUnmarshaler, _DestinationSpec_OneofSizer, []interface{}{
		(*DestinationSpec_Aws)(nil),
		(*DestinationSpec_Azure)(nil),
		(*DestinationSpec_Rest)(nil),
		(*DestinationSpec_Grpc)(nil),
		(*DestinationSpec_Sqoop)(nil),
	}
}

func _DestinationSpec_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*DestinationSpec)
	// destination_type
	switch x := m.DestinationType.(type) {
	case *DestinationSpec_Aws:
		_ = b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Aws); err != nil {
			return err
		}
	case *DestinationSpec_Azure:
		_ = b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Azure); err != nil {
			return err
		}
	case *DestinationSpec_Rest:
		_ = b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Rest); err != nil {
			return err
		}
	case *DestinationSpec_Grpc:
		_ = b.EncodeVarint(4<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Grpc); err != nil {
			return err
		}
	case *DestinationSpec_Sqoop:
		_ = b.EncodeVarint(5<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Sqoop); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("DestinationSpec.DestinationType has unexpected type %T", x)
	}
	return nil
}

func _DestinationSpec_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*DestinationSpec)
	switch tag {
	case 1: // destination_type.aws
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(aws.DestinationSpec)
		err := b.DecodeMessage(msg)
		m.DestinationType = &DestinationSpec_Aws{msg}
		return true, err
	case 2: // destination_type.azure
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(azure.DestinationSpec)
		err := b.DecodeMessage(msg)
		m.DestinationType = &DestinationSpec_Azure{msg}
		return true, err
	case 3: // destination_type.rest
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(rest.DestinationSpec)
		err := b.DecodeMessage(msg)
		m.DestinationType = &DestinationSpec_Rest{msg}
		return true, err
	case 4: // destination_type.grpc
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(grpc1.DestinationSpec)
		err := b.DecodeMessage(msg)
		m.DestinationType = &DestinationSpec_Grpc{msg}
		return true, err
	case 5: // destination_type.sqoop
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(sqoop.DestinationSpec)
		err := b.DecodeMessage(msg)
		m.DestinationType = &DestinationSpec_Sqoop{msg}
		return true, err
	default:
		return false, nil
	}
}

func _DestinationSpec_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*DestinationSpec)
	// destination_type
	switch x := m.DestinationType.(type) {
	case *DestinationSpec_Aws:
		s := proto.Size(x.Aws)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *DestinationSpec_Azure:
		s := proto.Size(x.Azure)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *DestinationSpec_Rest:
		s := proto.Size(x.Rest)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *DestinationSpec_Grpc:
		s := proto.Size(x.Grpc)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *DestinationSpec_Sqoop:
		s := proto.Size(x.Sqoop)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Each upstream in Gloo has a type. Supported types include `static`, `kubernetes`, `aws`, `consul`, and more.
// Each upstream type is handled by a corresponding Gloo plugin.
type UpstreamSpec struct {
	// Note to developers: new Upstream Plugins must be added to this oneof field
	// to be usable by Gloo.
	//
	// Types that are valid to be assigned to UpstreamType:
	//	*UpstreamSpec_Kube
	//	*UpstreamSpec_Static
	//	*UpstreamSpec_Aws
	//	*UpstreamSpec_Azure
	//	*UpstreamSpec_Consul
	UpstreamType         isUpstreamSpec_UpstreamType `protobuf_oneof:"upstream_type"`
	XXX_NoUnkeyedLiteral struct{}                    `json:"-"`
	XXX_unrecognized     []byte                      `json:"-"`
	XXX_sizecache        int32                       `json:"-"`
}

func (m *UpstreamSpec) Reset()         { *m = UpstreamSpec{} }
func (m *UpstreamSpec) String() string { return proto.CompactTextString(m) }
func (*UpstreamSpec) ProtoMessage()    {}
func (*UpstreamSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_plugins_f4f4de2ad43b6712, []int{4}
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

type isUpstreamSpec_UpstreamType interface {
	isUpstreamSpec_UpstreamType()
	Equal(interface{}) bool
}

type UpstreamSpec_Kube struct {
	Kube *kubernetes.UpstreamSpec `protobuf:"bytes,1,opt,name=kube,oneof"`
}
type UpstreamSpec_Static struct {
	Static *static.UpstreamSpec `protobuf:"bytes,4,opt,name=static,oneof"`
}
type UpstreamSpec_Aws struct {
	Aws *aws.UpstreamSpec `protobuf:"bytes,2,opt,name=aws,oneof"`
}
type UpstreamSpec_Azure struct {
	Azure *azure.UpstreamSpec `protobuf:"bytes,3,opt,name=azure,oneof"`
}
type UpstreamSpec_Consul struct {
	Consul *consul.UpstreamSpec `protobuf:"bytes,5,opt,name=consul,oneof"`
}

func (*UpstreamSpec_Kube) isUpstreamSpec_UpstreamType()   {}
func (*UpstreamSpec_Static) isUpstreamSpec_UpstreamType() {}
func (*UpstreamSpec_Aws) isUpstreamSpec_UpstreamType()    {}
func (*UpstreamSpec_Azure) isUpstreamSpec_UpstreamType()  {}
func (*UpstreamSpec_Consul) isUpstreamSpec_UpstreamType() {}

func (m *UpstreamSpec) GetUpstreamType() isUpstreamSpec_UpstreamType {
	if m != nil {
		return m.UpstreamType
	}
	return nil
}

func (m *UpstreamSpec) GetKube() *kubernetes.UpstreamSpec {
	if x, ok := m.GetUpstreamType().(*UpstreamSpec_Kube); ok {
		return x.Kube
	}
	return nil
}

func (m *UpstreamSpec) GetStatic() *static.UpstreamSpec {
	if x, ok := m.GetUpstreamType().(*UpstreamSpec_Static); ok {
		return x.Static
	}
	return nil
}

func (m *UpstreamSpec) GetAws() *aws.UpstreamSpec {
	if x, ok := m.GetUpstreamType().(*UpstreamSpec_Aws); ok {
		return x.Aws
	}
	return nil
}

func (m *UpstreamSpec) GetAzure() *azure.UpstreamSpec {
	if x, ok := m.GetUpstreamType().(*UpstreamSpec_Azure); ok {
		return x.Azure
	}
	return nil
}

func (m *UpstreamSpec) GetConsul() *consul.UpstreamSpec {
	if x, ok := m.GetUpstreamType().(*UpstreamSpec_Consul); ok {
		return x.Consul
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*UpstreamSpec) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _UpstreamSpec_OneofMarshaler, _UpstreamSpec_OneofUnmarshaler, _UpstreamSpec_OneofSizer, []interface{}{
		(*UpstreamSpec_Kube)(nil),
		(*UpstreamSpec_Static)(nil),
		(*UpstreamSpec_Aws)(nil),
		(*UpstreamSpec_Azure)(nil),
		(*UpstreamSpec_Consul)(nil),
	}
}

func _UpstreamSpec_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*UpstreamSpec)
	// upstream_type
	switch x := m.UpstreamType.(type) {
	case *UpstreamSpec_Kube:
		_ = b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Kube); err != nil {
			return err
		}
	case *UpstreamSpec_Static:
		_ = b.EncodeVarint(4<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Static); err != nil {
			return err
		}
	case *UpstreamSpec_Aws:
		_ = b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Aws); err != nil {
			return err
		}
	case *UpstreamSpec_Azure:
		_ = b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Azure); err != nil {
			return err
		}
	case *UpstreamSpec_Consul:
		_ = b.EncodeVarint(5<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Consul); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("UpstreamSpec.UpstreamType has unexpected type %T", x)
	}
	return nil
}

func _UpstreamSpec_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*UpstreamSpec)
	switch tag {
	case 1: // upstream_type.kube
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(kubernetes.UpstreamSpec)
		err := b.DecodeMessage(msg)
		m.UpstreamType = &UpstreamSpec_Kube{msg}
		return true, err
	case 4: // upstream_type.static
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(static.UpstreamSpec)
		err := b.DecodeMessage(msg)
		m.UpstreamType = &UpstreamSpec_Static{msg}
		return true, err
	case 2: // upstream_type.aws
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(aws.UpstreamSpec)
		err := b.DecodeMessage(msg)
		m.UpstreamType = &UpstreamSpec_Aws{msg}
		return true, err
	case 3: // upstream_type.azure
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(azure.UpstreamSpec)
		err := b.DecodeMessage(msg)
		m.UpstreamType = &UpstreamSpec_Azure{msg}
		return true, err
	case 5: // upstream_type.consul
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(consul.UpstreamSpec)
		err := b.DecodeMessage(msg)
		m.UpstreamType = &UpstreamSpec_Consul{msg}
		return true, err
	default:
		return false, nil
	}
}

func _UpstreamSpec_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*UpstreamSpec)
	// upstream_type
	switch x := m.UpstreamType.(type) {
	case *UpstreamSpec_Kube:
		s := proto.Size(x.Kube)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *UpstreamSpec_Static:
		s := proto.Size(x.Static)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *UpstreamSpec_Aws:
		s := proto.Size(x.Aws)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *UpstreamSpec_Azure:
		s := proto.Size(x.Azure)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *UpstreamSpec_Consul:
		s := proto.Size(x.Consul)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

func init() {
	proto.RegisterType((*ListenerPlugins)(nil), "gloo.solo.io.ListenerPlugins")
	proto.RegisterType((*VirtualHostPlugins)(nil), "gloo.solo.io.VirtualHostPlugins")
	proto.RegisterType((*RoutePlugins)(nil), "gloo.solo.io.RoutePlugins")
	proto.RegisterType((*DestinationSpec)(nil), "gloo.solo.io.DestinationSpec")
	proto.RegisterType((*UpstreamSpec)(nil), "gloo.solo.io.UpstreamSpec")
}
func (this *ListenerPlugins) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ListenerPlugins)
	if !ok {
		that2, ok := that.(ListenerPlugins)
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
func (this *VirtualHostPlugins) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*VirtualHostPlugins)
	if !ok {
		that2, ok := that.(VirtualHostPlugins)
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
	if !this.RateLimits.Equal(that1.RateLimits) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *RoutePlugins) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RoutePlugins)
	if !ok {
		that2, ok := that.(RoutePlugins)
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
	if !this.Transformations.Equal(that1.Transformations) {
		return false
	}
	if !this.Faults.Equal(that1.Faults) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *DestinationSpec) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DestinationSpec)
	if !ok {
		that2, ok := that.(DestinationSpec)
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
	if that1.DestinationType == nil {
		if this.DestinationType != nil {
			return false
		}
	} else if this.DestinationType == nil {
		return false
	} else if !this.DestinationType.Equal(that1.DestinationType) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *DestinationSpec_Aws) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DestinationSpec_Aws)
	if !ok {
		that2, ok := that.(DestinationSpec_Aws)
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
	if !this.Aws.Equal(that1.Aws) {
		return false
	}
	return true
}
func (this *DestinationSpec_Azure) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DestinationSpec_Azure)
	if !ok {
		that2, ok := that.(DestinationSpec_Azure)
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
	if !this.Azure.Equal(that1.Azure) {
		return false
	}
	return true
}
func (this *DestinationSpec_Rest) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DestinationSpec_Rest)
	if !ok {
		that2, ok := that.(DestinationSpec_Rest)
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
	if !this.Rest.Equal(that1.Rest) {
		return false
	}
	return true
}
func (this *DestinationSpec_Grpc) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DestinationSpec_Grpc)
	if !ok {
		that2, ok := that.(DestinationSpec_Grpc)
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
	if !this.Grpc.Equal(that1.Grpc) {
		return false
	}
	return true
}
func (this *DestinationSpec_Sqoop) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DestinationSpec_Sqoop)
	if !ok {
		that2, ok := that.(DestinationSpec_Sqoop)
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
	if !this.Sqoop.Equal(that1.Sqoop) {
		return false
	}
	return true
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
	if that1.UpstreamType == nil {
		if this.UpstreamType != nil {
			return false
		}
	} else if this.UpstreamType == nil {
		return false
	} else if !this.UpstreamType.Equal(that1.UpstreamType) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *UpstreamSpec_Kube) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*UpstreamSpec_Kube)
	if !ok {
		that2, ok := that.(UpstreamSpec_Kube)
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
	if !this.Kube.Equal(that1.Kube) {
		return false
	}
	return true
}
func (this *UpstreamSpec_Static) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*UpstreamSpec_Static)
	if !ok {
		that2, ok := that.(UpstreamSpec_Static)
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
	if !this.Static.Equal(that1.Static) {
		return false
	}
	return true
}
func (this *UpstreamSpec_Aws) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*UpstreamSpec_Aws)
	if !ok {
		that2, ok := that.(UpstreamSpec_Aws)
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
	if !this.Aws.Equal(that1.Aws) {
		return false
	}
	return true
}
func (this *UpstreamSpec_Azure) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*UpstreamSpec_Azure)
	if !ok {
		that2, ok := that.(UpstreamSpec_Azure)
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
	if !this.Azure.Equal(that1.Azure) {
		return false
	}
	return true
}
func (this *UpstreamSpec_Consul) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*UpstreamSpec_Consul)
	if !ok {
		that2, ok := that.(UpstreamSpec_Consul)
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
	if !this.Consul.Equal(that1.Consul) {
		return false
	}
	return true
}

func init() { proto.RegisterFile("plugins.proto", fileDescriptor_plugins_f4f4de2ad43b6712) }

var fileDescriptor_plugins_f4f4de2ad43b6712 = []byte{
	// 600 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x94, 0xcf, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0x49, 0x9a, 0xe6, 0xb0, 0x4d, 0x15, 0xb0, 0x38, 0x44, 0x3d, 0x20, 0x94, 0x03, 0x34,
	0xa0, 0xae, 0xf9, 0x23, 0x0e, 0x80, 0x10, 0x55, 0x82, 0xa0, 0x15, 0x45, 0x02, 0xb7, 0xe5, 0x80,
	0x84, 0xaa, 0x8d, 0xd9, 0x9a, 0xa5, 0x8e, 0xd7, 0xec, 0x8e, 0x29, 0xf0, 0x44, 0x3c, 0x00, 0x4f,
	0xc4, 0x99, 0x13, 0x4f, 0x80, 0x76, 0x76, 0xac, 0xa6, 0x56, 0x5a, 0x59, 0xf1, 0x21, 0xeb, 0xb5,
	0x3d, 0xdf, 0x6f, 0x3d, 0x33, 0x99, 0x8f, 0xad, 0xe7, 0x69, 0x91, 0xa8, 0xcc, 0xf2, 0xdc, 0x68,
	0xd0, 0x41, 0x2f, 0x49, 0xb5, 0xe6, 0x56, 0xa7, 0x9a, 0x2b, 0xbd, 0x71, 0x3d, 0xd1, 0x89, 0xc6,
	0x17, 0xa1, 0xdb, 0xf9, 0x98, 0x8d, 0x57, 0x89, 0x82, 0xcf, 0xc5, 0x94, 0xc7, 0x7a, 0x16, 0xba,
	0xc8, 0x2d, 0xa5, 0x43, 0x5b, 0xe4, 0xd2, 0x38, 0x6d, 0x28, 0x72, 0x15, 0xca, 0xef, 0x20, 0x4d,
	0x26, 0xd2, 0x10, 0x9f, 0x7c, 0xbb, 0x1f, 0xd2, 0x21, 0xa1, 0x38, 0xc5, 0x1f, 0x81, 0x76, 0x97,
	0x07, 0x19, 0x69, 0x01, 0x97, 0xe6, 0xa8, 0xc4, 0xe4, 0x31, 0x2e, 0x84, 0x7a, 0xdd, 0x20, 0xbd,
	0x9f, 0x85, 0x91, 0x7e, 0x25, 0xd8, 0x9b, 0xe5, 0x61, 0xb1, 0xce, 0x6c, 0x91, 0xd2, 0x85, 0x70,
	0x87, 0xcb, 0xe3, 0x4e, 0x8a, 0xa9, 0x34, 0x99, 0x04, 0x39, 0xbf, 0x6d, 0x9e, 0xb2, 0xfd, 0xaa,
	0x75, 0xee, 0xd7, 0xe6, 0x29, 0x5b, 0x10, 0xa0, 0x62, 0xba, 0x10, 0xee, 0xe3, 0xf2, 0x38, 0x30,
	0x22, 0xb3, 0xc7, 0xda, 0xcc, 0x04, 0x28, 0x9d, 0x55, 0x6e, 0x09, 0x7f, 0xb0, 0x3c, 0xfe, 0x58,
	0x14, 0x29, 0xa8, 0xec, 0x8b, 0x8c, 0x11, 0x8f, 0xb7, 0x44, 0xdd, 0x6f, 0xf0, 0xcf, 0x16, 0x20,
	0x53, 0x35, 0x53, 0x70, 0xb6, 0xf3, 0xd0, 0xe1, 0x35, 0xd6, 0xdf, 0x53, 0x16, 0x64, 0x26, 0xcd,
	0x5b, 0x1f, 0x3c, 0x4c, 0x58, 0xf0, 0x5e, 0x19, 0x28, 0x44, 0xba, 0xa3, 0x2d, 0xd0, 0xd3, 0xe0,
	0x1d, 0x5b, 0x73, 0xda, 0x23, 0x14, 0xdb, 0x41, 0xeb, 0x66, 0x6b, 0x73, 0xed, 0xc1, 0x3d, 0x3e,
	0xc7, 0xa3, 0x99, 0x9f, 0x1f, 0x76, 0xbe, 0x9b, 0x25, 0x46, 0x5a, 0x1b, 0x09, 0x90, 0x7b, 0x2e,
	0x2a, 0x62, 0xa6, 0xdc, 0xda, 0xe1, 0xef, 0x16, 0xeb, 0x45, 0xba, 0x00, 0x59, 0x9e, 0x11, 0xb3,
	0xfe, 0xf9, 0x7a, 0x96, 0xe7, 0x3c, 0xe6, 0xd5, 0x3a, 0x2f, 0x3a, 0x0c, 0x59, 0x07, 0xe7, 0x01,
	0x51, 0x95, 0x18, 0x3c, 0x67, 0x5d, 0xac, 0xaa, 0x1d, 0xb4, 0x91, 0x7d, 0x9b, 0x53, 0x91, 0x2f,
	0x44, 0xbe, 0xc4, 0xf0, 0x88, 0x64, 0xc3, 0x7f, 0x6d, 0xd6, 0x7f, 0x21, 0x2d, 0xa8, 0x0c, 0x89,
	0xfb, 0xb9, 0x8c, 0x83, 0x67, 0x6c, 0x45, 0x9c, 0x96, 0x5f, 0x3b, 0xe2, 0x68, 0x47, 0x8b, 0x78,
	0x15, 0xdd, 0xce, 0x95, 0xc8, 0xe9, 0x82, 0x09, 0x5b, 0xc5, 0x01, 0xa7, 0x4f, 0xba, 0xcb, 0x69,
	0xdc, 0xeb, 0x21, 0xbc, 0x36, 0xd8, 0x66, 0x1d, 0x67, 0x5e, 0x83, 0x15, 0x64, 0xdc, 0xe1, 0xde,
	0xc9, 0xea, 0x21, 0x50, 0xe9, 0x08, 0xce, 0xb3, 0x06, 0x1d, 0x22, 0x78, 0x03, 0xab, 0x49, 0x70,
	0xc1, 0x2e, 0x11, 0x1c, 0xdb, 0xc1, 0x2a, 0x25, 0x42, 0x43, 0x5c, 0x33, 0x11, 0x8c, 0x1e, 0x07,
	0xec, 0xea, 0xa7, 0xb3, 0x77, 0x47, 0xf0, 0x23, 0x97, 0xc3, 0xbf, 0x6d, 0xd6, 0x3b, 0xcc, 0x2d,
	0x18, 0x29, 0x66, 0x58, 0xf1, 0x09, 0xeb, 0x38, 0xcb, 0xa1, 0x92, 0x6f, 0xf1, 0x79, 0xff, 0x59,
	0x74, 0xda, 0xbc, 0xd8, 0x7d, 0xae, 0x8b, 0x0f, 0x26, 0xac, 0xeb, 0x7d, 0x81, 0x52, 0x1e, 0xf1,
	0xd2, 0x26, 0x6a, 0x20, 0x48, 0x1a, 0x3c, 0xf1, 0xbd, 0xf7, 0xad, 0xbb, 0x75, 0x71, 0xef, 0x2b,
	0x72, 0x6c, 0xfc, 0x76, 0xd9, 0x78, 0xdf, 0xb4, 0xcd, 0xcb, 0x1a, 0x5f, 0xd1, 0x53, 0xd7, 0x27,
	0xac, 0xeb, 0xdd, 0x9c, 0x4a, 0x3e, 0xe2, 0xa5, 0xb9, 0xd7, 0x49, 0xc1, 0xc7, 0x8e, 0xfb, 0x6c,
	0xbd, 0xa0, 0x37, 0x58, 0xee, 0xf1, 0xd3, 0x5f, 0x7f, 0x6e, 0xb4, 0x3e, 0x3c, 0xba, 0xd4, 0x71,
	0xf2, 0x93, 0x64, 0xa1, 0xeb, 0x4c, 0xbb, 0x68, 0x2d, 0x0f, 0xff, 0x07, 0x00, 0x00, 0xff, 0xff,
	0x8c, 0xeb, 0x89, 0xc0, 0x07, 0x08, 0x00, 0x00,
}
