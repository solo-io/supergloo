// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/service-mesh-hub/api/networking/v1alpha2/access_policy.proto

package v1alpha2

import (
	bytes "bytes"
	fmt "fmt"
	math "math"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/types"
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
//Access control policies apply ALLOW policies to communication in a mesh.
//Access control policies specify the following:
//ALLOW those requests that: originate from from **source workload**, target the **destination target**,
//and match the indicated request criteria (allowed_paths, allowed_methods, allowed_ports).
//Enforcement of access control is determined by the
//[VirtualMesh's GlobalAccessPolicy]({{% versioned_link_path fromRoot="/reference/api/virtual_mesh/#networking.smh.solo.io.VirtualMeshSpec.GlobalAccessPolicy" %}})
type AccessPolicySpec struct {
	//
	//Requests originating from these pods will have the rule applied.
	//Leave empty to have all pods in the mesh apply these policies.
	//
	//Note that access control policies are mapped to source pods by their
	//service account. If other pods share the same service account,
	//this access control rule will apply to those pods as well.
	//
	//For fine-grained access control policies, ensure that your
	//service accounts properly reflect the desired
	//boundary for your access control policies.
	SourceSelector []*IdentitySelector `protobuf:"bytes,2,rep,name=source_selector,json=sourceSelector,proto3" json:"source_selector,omitempty"`
	//
	//Requests destined for these pods will have the rule applied.
	//Leave empty to apply to all destination pods in the mesh.
	DestinationSelector []*ServiceSelector `protobuf:"bytes,3,rep,name=destination_selector,json=destinationSelector,proto3" json:"destination_selector,omitempty"`
	//
	//Optional. A list of HTTP paths or gRPC methods to allow.
	//gRPC methods must be presented as fully-qualified name in the form of
	//"/packageName.serviceName/methodName" and are case sensitive.
	//Exact match, prefix match, and suffix match are supported for paths.
	//For example, the path "/books/review" matches
	//"/books/review" (exact match), "*books/" (suffix match), or "/books*" (prefix match).
	//
	//If not specified, allow any path.
	AllowedPaths []string `protobuf:"bytes,4,rep,name=allowed_paths,json=allowedPaths,proto3" json:"allowed_paths,omitempty"`
	//
	//Optional. A list of HTTP methods to allow (e.g., "GET", "POST").
	//It is ignored in gRPC case because the value is always "POST".
	//If not specified, allows any method.
	AllowedMethods []types.HttpMethodValue `protobuf:"varint,5,rep,packed,name=allowed_methods,json=allowedMethods,proto3,enum=networking.smh.solo.io.HttpMethodValue" json:"allowed_methods,omitempty"`
	//
	//Optional. A list of ports which to allow.
	//If not set any port is allowed.
	AllowedPorts         []uint32 `protobuf:"varint,6,rep,packed,name=allowed_ports,json=allowedPorts,proto3" json:"allowed_ports,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AccessPolicySpec) Reset()         { *m = AccessPolicySpec{} }
func (m *AccessPolicySpec) String() string { return proto.CompactTextString(m) }
func (*AccessPolicySpec) ProtoMessage()    {}
func (*AccessPolicySpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_a607a654bf8f02aa, []int{0}
}
func (m *AccessPolicySpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccessPolicySpec.Unmarshal(m, b)
}
func (m *AccessPolicySpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccessPolicySpec.Marshal(b, m, deterministic)
}
func (m *AccessPolicySpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccessPolicySpec.Merge(m, src)
}
func (m *AccessPolicySpec) XXX_Size() int {
	return xxx_messageInfo_AccessPolicySpec.Size(m)
}
func (m *AccessPolicySpec) XXX_DiscardUnknown() {
	xxx_messageInfo_AccessPolicySpec.DiscardUnknown(m)
}

var xxx_messageInfo_AccessPolicySpec proto.InternalMessageInfo

func (m *AccessPolicySpec) GetSourceSelector() []*IdentitySelector {
	if m != nil {
		return m.SourceSelector
	}
	return nil
}

func (m *AccessPolicySpec) GetDestinationSelector() []*ServiceSelector {
	if m != nil {
		return m.DestinationSelector
	}
	return nil
}

func (m *AccessPolicySpec) GetAllowedPaths() []string {
	if m != nil {
		return m.AllowedPaths
	}
	return nil
}

func (m *AccessPolicySpec) GetAllowedMethods() []types.HttpMethodValue {
	if m != nil {
		return m.AllowedMethods
	}
	return nil
}

func (m *AccessPolicySpec) GetAllowedPorts() []uint32 {
	if m != nil {
		return m.AllowedPorts
	}
	return nil
}

type AccessPolicyStatus struct {
	// The most recent generation observed in the the AccessPolicy metadata.
	// If the observedGeneration does not match generation, the controller has not received the most
	// recent version of this resource.
	ObservedGeneration int64 `protobuf:"varint,1,opt,name=observed_generation,json=observedGeneration,proto3" json:"observed_generation,omitempty"`
	// The state of the overall resource.
	// It will only show accepted if it has been successfully
	// applied to all target meshes.
	State ApprovalState `protobuf:"varint,2,opt,name=state,proto3,enum=networking.smh.solo.io.ApprovalState" json:"state,omitempty"`
	// The status of the AccessPolicy for each TrafficTarget to which it has been applied.
	// An AccessPolicy may be Accepted for some TrafficTargets and rejected for others.
	TrafficTargets map[string]*ApprovalStatus `protobuf:"bytes,3,rep,name=traffic_targets,json=trafficTargets,proto3" json:"traffic_targets,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// The list of Workloads to which this policy has been applied.
	Workloads []string `protobuf:"bytes,4,rep,name=workloads,proto3" json:"workloads,omitempty"`
	// A list of errors pertaining to the configuration targets of this resource (e.g. referencing a traffic target that does not exist).
	ConfigTargetErrors   []string `protobuf:"bytes,5,rep,name=config_target_errors,json=configTargetErrors,proto3" json:"config_target_errors,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AccessPolicyStatus) Reset()         { *m = AccessPolicyStatus{} }
func (m *AccessPolicyStatus) String() string { return proto.CompactTextString(m) }
func (*AccessPolicyStatus) ProtoMessage()    {}
func (*AccessPolicyStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_a607a654bf8f02aa, []int{1}
}
func (m *AccessPolicyStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccessPolicyStatus.Unmarshal(m, b)
}
func (m *AccessPolicyStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccessPolicyStatus.Marshal(b, m, deterministic)
}
func (m *AccessPolicyStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccessPolicyStatus.Merge(m, src)
}
func (m *AccessPolicyStatus) XXX_Size() int {
	return xxx_messageInfo_AccessPolicyStatus.Size(m)
}
func (m *AccessPolicyStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_AccessPolicyStatus.DiscardUnknown(m)
}

var xxx_messageInfo_AccessPolicyStatus proto.InternalMessageInfo

func (m *AccessPolicyStatus) GetObservedGeneration() int64 {
	if m != nil {
		return m.ObservedGeneration
	}
	return 0
}

func (m *AccessPolicyStatus) GetState() ApprovalState {
	if m != nil {
		return m.State
	}
	return ApprovalState_PENDING
}

func (m *AccessPolicyStatus) GetTrafficTargets() map[string]*ApprovalStatus {
	if m != nil {
		return m.TrafficTargets
	}
	return nil
}

func (m *AccessPolicyStatus) GetWorkloads() []string {
	if m != nil {
		return m.Workloads
	}
	return nil
}

func (m *AccessPolicyStatus) GetConfigTargetErrors() []string {
	if m != nil {
		return m.ConfigTargetErrors
	}
	return nil
}

func init() {
	proto.RegisterType((*AccessPolicySpec)(nil), "networking.smh.solo.io.AccessPolicySpec")
	proto.RegisterType((*AccessPolicyStatus)(nil), "networking.smh.solo.io.AccessPolicyStatus")
	proto.RegisterMapType((map[string]*ApprovalStatus)(nil), "networking.smh.solo.io.AccessPolicyStatus.TrafficTargetsEntry")
}

func init() {
	proto.RegisterFile("github.com/solo-io/service-mesh-hub/api/networking/v1alpha2/access_policy.proto", fileDescriptor_a607a654bf8f02aa)
}

var fileDescriptor_a607a654bf8f02aa = []byte{
	// 539 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0xcf, 0x6a, 0xdb, 0x4e,
	0x10, 0xc7, 0xb1, 0xf5, 0x4b, 0xc0, 0x9b, 0x5f, 0x9c, 0xb0, 0x36, 0x45, 0x98, 0x52, 0x44, 0x4a,
	0x5b, 0x5d, 0x2c, 0xb5, 0xce, 0x25, 0xb4, 0xa5, 0x25, 0x85, 0xd0, 0x96, 0x52, 0xea, 0xc8, 0xa1,
	0x87, 0x5c, 0xc4, 0x5a, 0x5a, 0x4b, 0x8b, 0x65, 0x8d, 0xd8, 0x1d, 0x39, 0xf8, 0x85, 0x4a, 0x5f,
	0xaa, 0x97, 0x3e, 0x49, 0xd1, 0xae, 0x64, 0xc7, 0x6d, 0x0c, 0xbe, 0xad, 0xbe, 0x33, 0xf3, 0x99,
	0x7f, 0x8c, 0xc8, 0xb7, 0x44, 0x60, 0x5a, 0x4e, 0xbd, 0x08, 0x16, 0xbe, 0x82, 0x0c, 0x86, 0x02,
	0x7c, 0xc5, 0xe5, 0x52, 0x44, 0x7c, 0xb8, 0xe0, 0x2a, 0x1d, 0xa6, 0xe5, 0xd4, 0x67, 0x85, 0xf0,
	0x73, 0x8e, 0x77, 0x20, 0xe7, 0x22, 0x4f, 0xfc, 0xe5, 0x2b, 0x96, 0x15, 0x29, 0x1b, 0xf9, 0x2c,
	0x8a, 0xb8, 0x52, 0x61, 0x01, 0x99, 0x88, 0x56, 0x5e, 0x21, 0x01, 0x81, 0x3e, 0xda, 0x38, 0x7a,
	0x6a, 0x91, 0x7a, 0x15, 0xd4, 0x13, 0x30, 0x38, 0xdf, 0x9b, 0x9a, 0x22, 0x16, 0x06, 0x36, 0xb8,
	0xd8, 0x3b, 0x48, 0xf1, 0x8c, 0x47, 0x08, 0x52, 0xd5, 0x91, 0xef, 0xf7, 0x8e, 0x5c, 0xb2, 0x4c,
	0xc4, 0x0c, 0x05, 0xe4, 0xa1, 0x42, 0x86, 0xbc, 0x06, 0xf4, 0x13, 0x48, 0x40, 0x3f, 0xfd, 0xea,
	0x65, 0xd4, 0xb3, 0x5f, 0x6d, 0x72, 0x7a, 0xa9, 0xbb, 0x1e, 0xeb, 0xa6, 0x27, 0x05, 0x8f, 0xe8,
	0x35, 0x39, 0x51, 0x50, 0xca, 0x88, 0x87, 0x4d, 0x15, 0x76, 0xdb, 0xb1, 0xdc, 0xa3, 0x91, 0xeb,
	0x3d, 0x3c, 0x0c, 0xef, 0x73, 0xcc, 0x73, 0x14, 0xb8, 0x9a, 0xd4, 0xfe, 0x41, 0xd7, 0x00, 0x9a,
	0x6f, 0x7a, 0x4b, 0xfa, 0x31, 0x57, 0x28, 0xf2, 0xba, 0xb0, 0x86, 0x6b, 0x69, 0xee, 0x8b, 0x5d,
	0xdc, 0x89, 0x69, 0x7a, 0x8d, 0xed, 0xdd, 0x83, 0xac, 0xd9, 0x4f, 0xc9, 0x31, 0xcb, 0x32, 0xb8,
	0xe3, 0x71, 0x58, 0x30, 0x4c, 0x95, 0xfd, 0x9f, 0x63, 0xb9, 0x9d, 0xe0, 0xff, 0x5a, 0x1c, 0x57,
	0x1a, 0x1d, 0x93, 0x93, 0xc6, 0x69, 0xc1, 0x31, 0x85, 0x58, 0xd9, 0x07, 0x8e, 0xe5, 0x76, 0x77,
	0xe7, 0xfe, 0x84, 0x58, 0x7c, 0xd5, 0xae, 0xdf, 0x59, 0x56, 0xf2, 0xa0, 0x5b, 0xc7, 0x1b, 0x4d,
	0x6d, 0xa5, 0x05, 0x89, 0xca, 0x3e, 0x74, 0x2c, 0xf7, 0x78, 0x93, 0xb6, 0xd2, 0xce, 0x7e, 0x58,
	0x84, 0x6e, 0xcd, 0x17, 0x19, 0x96, 0x8a, 0xfa, 0xa4, 0x07, 0xd3, 0x6a, 0xa3, 0x3c, 0x0e, 0x13,
	0x9e, 0x73, 0xa9, 0x3b, 0xb2, 0x5b, 0x4e, 0xcb, 0xb5, 0x02, 0xda, 0x98, 0x3e, 0xae, 0x2d, 0xf4,
	0x0d, 0x39, 0xd0, 0xcb, 0xb4, 0xdb, 0x4e, 0xcb, 0xed, 0x8e, 0x9e, 0xed, 0x2a, 0xfa, 0xb2, 0x28,
	0x24, 0x2c, 0x59, 0x56, 0xe5, 0xe1, 0x81, 0x89, 0xa1, 0x09, 0x39, 0x41, 0xc9, 0x66, 0x33, 0x11,
	0x85, 0xc8, 0x64, 0xc2, 0x51, 0xd5, 0x73, 0x7f, 0xb7, 0x13, 0xf3, 0x4f, 0xc9, 0xde, 0x8d, 0x21,
	0xdc, 0x18, 0xc0, 0x55, 0x8e, 0x72, 0x15, 0x74, 0x71, 0x4b, 0xa4, 0x8f, 0x49, 0xa7, 0xa2, 0x65,
	0xc0, 0xe2, 0x66, 0x0b, 0x1b, 0x81, 0xbe, 0x24, 0xfd, 0x08, 0xf2, 0x99, 0x48, 0xea, 0x2a, 0x42,
	0x2e, 0x25, 0x48, 0xb3, 0x87, 0x4e, 0x40, 0x8d, 0xcd, 0xa0, 0xae, 0xb4, 0x65, 0x20, 0x48, 0xef,
	0x81, 0xb4, 0xf4, 0x94, 0x58, 0x73, 0xbe, 0xd2, 0xd3, 0xea, 0x04, 0xd5, 0x93, 0xbe, 0x25, 0x07,
	0xcb, 0x6a, 0x49, 0x7a, 0x3c, 0x47, 0xa3, 0xe7, 0xfb, 0x8c, 0xa7, 0x54, 0x81, 0x09, 0x7a, 0xdd,
	0xbe, 0x68, 0x7d, 0xb8, 0xfe, 0xf9, 0xfb, 0x49, 0xeb, 0xf6, 0xcb, 0x3e, 0x7f, 0x8f, 0x62, 0x9e,
	0xfc, 0x75, 0x7c, 0xf7, 0x73, 0xac, 0x0f, 0x71, 0x7a, 0xa8, 0x4f, 0xec, 0xfc, 0x4f, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x62, 0x90, 0x68, 0x42, 0x93, 0x04, 0x00, 0x00,
}

func (this *AccessPolicySpec) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*AccessPolicySpec)
	if !ok {
		that2, ok := that.(AccessPolicySpec)
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
	if len(this.SourceSelector) != len(that1.SourceSelector) {
		return false
	}
	for i := range this.SourceSelector {
		if !this.SourceSelector[i].Equal(that1.SourceSelector[i]) {
			return false
		}
	}
	if len(this.DestinationSelector) != len(that1.DestinationSelector) {
		return false
	}
	for i := range this.DestinationSelector {
		if !this.DestinationSelector[i].Equal(that1.DestinationSelector[i]) {
			return false
		}
	}
	if len(this.AllowedPaths) != len(that1.AllowedPaths) {
		return false
	}
	for i := range this.AllowedPaths {
		if this.AllowedPaths[i] != that1.AllowedPaths[i] {
			return false
		}
	}
	if len(this.AllowedMethods) != len(that1.AllowedMethods) {
		return false
	}
	for i := range this.AllowedMethods {
		if this.AllowedMethods[i] != that1.AllowedMethods[i] {
			return false
		}
	}
	if len(this.AllowedPorts) != len(that1.AllowedPorts) {
		return false
	}
	for i := range this.AllowedPorts {
		if this.AllowedPorts[i] != that1.AllowedPorts[i] {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *AccessPolicyStatus) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*AccessPolicyStatus)
	if !ok {
		that2, ok := that.(AccessPolicyStatus)
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
	if this.ObservedGeneration != that1.ObservedGeneration {
		return false
	}
	if this.State != that1.State {
		return false
	}
	if len(this.TrafficTargets) != len(that1.TrafficTargets) {
		return false
	}
	for i := range this.TrafficTargets {
		if !this.TrafficTargets[i].Equal(that1.TrafficTargets[i]) {
			return false
		}
	}
	if len(this.Workloads) != len(that1.Workloads) {
		return false
	}
	for i := range this.Workloads {
		if this.Workloads[i] != that1.Workloads[i] {
			return false
		}
	}
	if len(this.ConfigTargetErrors) != len(that1.ConfigTargetErrors) {
		return false
	}
	for i := range this.ConfigTargetErrors {
		if this.ConfigTargetErrors[i] != that1.ConfigTargetErrors[i] {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
