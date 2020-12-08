// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/solo-io/gloo-mesh/api/networking/v1alpha2/access_policy.proto

package v1alpha2

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	types "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2/types"
	_ "github.com/solo-io/protoc-gen-ext/extproto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

//
//Access control policies apply ALLOW policies to communication in a mesh.
//Access control policies specify the following:
//ALLOW those requests that: originate from from **source workload**, target the **destination target**,
//and match the indicated request criteria (allowed_paths, allowed_methods, allowed_ports).
//Enforcement of access control is determined by the
//[VirtualMesh's GlobalAccessPolicy]({{% versioned_link_path fromRoot="/reference/api/virtual_mesh/#networking.mesh.gloo.solo.io.VirtualMeshSpec.GlobalAccessPolicy" %}})
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
	DestinationSelector []*TrafficTargetSelector `protobuf:"bytes,3,rep,name=destination_selector,json=destinationSelector,proto3" json:"destination_selector,omitempty"`
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
	AllowedMethods []types.HttpMethodValue `protobuf:"varint,5,rep,packed,name=allowed_methods,json=allowedMethods,proto3,enum=networking.mesh.gloo.solo.io.HttpMethodValue" json:"allowed_methods,omitempty"`
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
	return fileDescriptor_36a0dece7ff0ff65, []int{0}
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

func (m *AccessPolicySpec) GetDestinationSelector() []*TrafficTargetSelector {
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
	State ApprovalState `protobuf:"varint,2,opt,name=state,proto3,enum=networking.mesh.gloo.solo.io.ApprovalState" json:"state,omitempty"`
	// The status of the AccessPolicy for each TrafficTarget to which it has been applied.
	// An AccessPolicy may be Accepted for some TrafficTargets and rejected for others.
	TrafficTargets map[string]*ApprovalStatus `protobuf:"bytes,3,rep,name=traffic_targets,json=trafficTargets,proto3" json:"traffic_targets,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// The list of Workloads to which this policy has been applied.
	Workloads []string `protobuf:"bytes,4,rep,name=workloads,proto3" json:"workloads,omitempty"`
	// Any errors found while processing this generation of the resource.
	Errors               []string `protobuf:"bytes,5,rep,name=errors,proto3" json:"errors,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AccessPolicyStatus) Reset()         { *m = AccessPolicyStatus{} }
func (m *AccessPolicyStatus) String() string { return proto.CompactTextString(m) }
func (*AccessPolicyStatus) ProtoMessage()    {}
func (*AccessPolicyStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_36a0dece7ff0ff65, []int{1}
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

func (m *AccessPolicyStatus) GetErrors() []string {
	if m != nil {
		return m.Errors
	}
	return nil
}

func init() {
	proto.RegisterType((*AccessPolicySpec)(nil), "networking.mesh.gloo.solo.io.AccessPolicySpec")
	proto.RegisterType((*AccessPolicyStatus)(nil), "networking.mesh.gloo.solo.io.AccessPolicyStatus")
	proto.RegisterMapType((map[string]*ApprovalStatus)(nil), "networking.mesh.gloo.solo.io.AccessPolicyStatus.TrafficTargetsEntry")
}

func init() {
	proto.RegisterFile("github.com/solo-io/gloo-mesh/api/networking/v1alpha2/access_policy.proto", fileDescriptor_36a0dece7ff0ff65)
}

var fileDescriptor_36a0dece7ff0ff65 = []byte{
	// 522 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x93, 0xc1, 0x6a, 0xdb, 0x40,
	0x10, 0x86, 0xb1, 0x15, 0x1b, 0xbc, 0x69, 0xec, 0xb0, 0x2e, 0x45, 0x98, 0x1c, 0x44, 0x7a, 0x11,
	0xb4, 0x96, 0xa8, 0x73, 0x29, 0xbd, 0x14, 0x87, 0x94, 0xa6, 0x2d, 0x05, 0xa3, 0x84, 0x14, 0x7a,
	0x11, 0x6b, 0x69, 0x2c, 0x09, 0xcb, 0x1a, 0xb1, 0x3b, 0x76, 0xe2, 0x97, 0xe9, 0xb3, 0xf4, 0xa5,
	0x72, 0x2f, 0x5a, 0x49, 0x76, 0xdd, 0x14, 0xb7, 0xf8, 0xa6, 0xfd, 0x77, 0xe7, 0xfb, 0x47, 0xff,
	0x30, 0xec, 0x3a, 0x4a, 0x28, 0x5e, 0x4e, 0x9d, 0x00, 0x17, 0xae, 0xc2, 0x14, 0x87, 0x09, 0xba,
	0x51, 0x8a, 0x38, 0x5c, 0x80, 0x8a, 0x5d, 0x91, 0x27, 0x6e, 0x06, 0x74, 0x8f, 0x72, 0x9e, 0x64,
	0x91, 0xbb, 0x7a, 0x23, 0xd2, 0x3c, 0x16, 0x23, 0x57, 0x04, 0x01, 0x28, 0xe5, 0xe7, 0x98, 0x26,
	0xc1, 0xda, 0xc9, 0x25, 0x12, 0xf2, 0xb3, 0xed, 0x43, 0xa7, 0x28, 0x76, 0x0a, 0x8c, 0x53, 0x30,
	0x9d, 0x04, 0x07, 0xef, 0x0f, 0xf2, 0x89, 0x89, 0xf2, 0x12, 0x3f, 0xb8, 0x3a, 0x08, 0xa0, 0x20,
	0x85, 0x80, 0x50, 0xaa, 0x8a, 0xf2, 0xe5, 0x20, 0xca, 0x4a, 0xa4, 0x49, 0x28, 0x28, 0xc1, 0xcc,
	0x57, 0x24, 0x08, 0x2a, 0x18, 0x87, 0x07, 0xd2, 0x5f, 0x2e, 0x3c, 0x50, 0xa9, 0x9d, 0x3f, 0x36,
	0xd9, 0xe9, 0x58, 0xa7, 0x33, 0xd1, 0xe1, 0xdc, 0xe4, 0x10, 0xf0, 0x6f, 0xac, 0xa7, 0x70, 0x29,
	0x03, 0xf0, 0xeb, 0x7e, 0xcc, 0xa6, 0x65, 0xd8, 0xc7, 0x23, 0xc7, 0xd9, 0x17, 0x9a, 0xf3, 0x29,
	0x84, 0x8c, 0x12, 0x5a, 0xdf, 0x54, 0x55, 0x5e, 0xb7, 0xc4, 0xd4, 0x67, 0x3e, 0x63, 0xcf, 0x43,
	0x50, 0x94, 0x64, 0x55, 0x73, 0x35, 0xdd, 0xd0, 0xf4, 0x8b, 0xfd, 0xf4, 0x5b, 0x29, 0x66, 0xb3,
	0x24, 0xb8, 0x15, 0x32, 0x02, 0xda, 0x58, 0xf4, 0x7f, 0x03, 0x6e, 0x7c, 0x5e, 0xb2, 0x13, 0x91,
	0xa6, 0x78, 0x0f, 0xa1, 0x9f, 0x0b, 0x8a, 0x95, 0x79, 0x64, 0x19, 0x76, 0xc7, 0x7b, 0x56, 0x89,
	0x93, 0x42, 0xe3, 0x77, 0xac, 0x57, 0x3f, 0x5a, 0x00, 0xc5, 0x18, 0x2a, 0xb3, 0x65, 0x19, 0x76,
	0x77, 0x34, 0xdc, 0xdf, 0xc7, 0x35, 0x51, 0xfe, 0x55, 0x17, 0xdc, 0x89, 0x74, 0x09, 0x5e, 0xb7,
	0xa2, 0x94, 0x9a, 0xda, 0x31, 0x47, 0x49, 0xca, 0x6c, 0x5b, 0x86, 0x7d, 0xb2, 0x35, 0x2f, 0xb4,
	0xf3, 0x1f, 0x06, 0xe3, 0x3b, 0xb9, 0x93, 0xa0, 0xa5, 0xe2, 0x2e, 0xeb, 0xe3, 0x54, 0x81, 0x5c,
	0x41, 0xe8, 0x47, 0x90, 0x81, 0xd4, 0xff, 0x65, 0x36, 0xac, 0x86, 0x6d, 0x78, 0xbc, 0xbe, 0xfa,
	0xb8, 0xb9, 0xe1, 0x63, 0xd6, 0xd2, 0x23, 0x36, 0x9b, 0x56, 0xc3, 0xee, 0x8e, 0x5e, 0xed, 0x6f,
	0x7d, 0x9c, 0xe7, 0x12, 0x57, 0x22, 0x2d, 0xdc, 0xc0, 0x2b, 0x2b, 0xf9, 0x82, 0xf5, 0xa8, 0x8c,
	0xd6, 0x27, 0x9d, 0xad, 0xaa, 0xe6, 0x71, 0xf5, 0x0f, 0xd8, 0x93, 0xf6, 0x77, 0x47, 0xa4, 0x3e,
	0x64, 0x24, 0xd7, 0x5e, 0x97, 0x76, 0x44, 0x7e, 0xc6, 0x3a, 0x05, 0x33, 0x45, 0x11, 0xd6, 0x73,
	0xd9, 0x0a, 0xfc, 0x05, 0x6b, 0x83, 0x94, 0x28, 0xcb, 0x59, 0x74, 0xbc, 0xea, 0x34, 0x40, 0xd6,
	0xff, 0x0b, 0x9c, 0x9f, 0x32, 0x63, 0x0e, 0x6b, 0x9d, 0x4f, 0xc7, 0x2b, 0x3e, 0xf9, 0x25, 0x6b,
	0xad, 0x8a, 0xb1, 0xe8, 0x40, 0x8e, 0x47, 0xaf, 0xff, 0x3f, 0x90, 0xa5, 0xf2, 0xca, 0xd2, 0x77,
	0xcd, 0xb7, 0x8d, 0xcb, 0xc9, 0xcf, 0xc7, 0xa3, 0xc6, 0xf7, 0xcf, 0x7b, 0xf7, 0x2f, 0x9f, 0x47,
	0x7f, 0xec, 0xe0, 0x53, 0x8b, 0xcd, 0x56, 0x4e, 0xdb, 0x7a, 0xe3, 0x2e, 0x7e, 0x05, 0x00, 0x00,
	0xff, 0xff, 0x71, 0xfb, 0x15, 0x47, 0xc3, 0x04, 0x00, 0x00,
}
