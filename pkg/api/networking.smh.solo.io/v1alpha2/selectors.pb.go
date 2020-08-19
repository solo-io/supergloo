// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/service-mesh-hub/api/networking/v1alpha2/selectors.proto

package v1alpha2

import (
	bytes "bytes"
	fmt "fmt"
	math "math"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
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
//Select Kubernetes services
//
//Only one of (labels + namespaces + cluster) or (resource refs) may be provided. If all four are provided, it will be
//considered an error, and the Status of the top level resource will be updated to reflect an IllegalSelection.
//
//Valid:
//1.
//selector:
//matcher:
//labels:
//foo: bar
//hello: world
//namespaces:
//- default
//cluster: "cluster-name"
//2.
//selector:
//matcher:
//refs:
//- name: foo
//namespace: bar
//
//Invalid:
//1.
//selector:
//matcher:
//labels:
//foo: bar
//hello: world
//namespaces:
//- default
//cluster: "cluster-name"
//refs:
//- name: foo
//namespace: bar
//
//By default labels will select across all namespaces, unless a list of namespaces is provided, in which case
//it will only select from those. An empty list is equal to AllNamespaces.
//
//If no labels are given, and only namespaces, all resources from the namespaces will be selected.
//
//The following selector will select all resources with the following labels in every namespace, in the local cluster:
//
//selector:
//matcher:
//labels:
//foo: bar
//hello: world
//
//Whereas the next selector will only select from the specified namespaces (foo, bar), in the local cluster:
//
//selector:
//matcher:
//labels:
//foo: bar
//hello: world
//namespaces
//- foo
//- bar
//
//This final selector will select all resources of a given type in the target namespace (foo), in the local cluster:
//
//selector
//matcher:
//namespaces
//- foo
//- bar
//labels:
//hello: world
//
//
type ServiceSelector struct {
	// a KubeServiceMatcher matches kubernetes services by the namespaces and clusters they belong to, as well
	// as the provided labels.
	KubeServiceMatcher *ServiceSelector_KubeServiceMatcher `protobuf:"bytes,1,opt,name=kube_service_matcher,json=kubeServiceMatcher,proto3" json:"kube_service_matcher,omitempty"`
	// Match individual k8s Services by direct reference.
	KubeServiceRefs      *ServiceSelector_KubeServiceRefs `protobuf:"bytes,2,opt,name=kube_service_refs,json=kubeServiceRefs,proto3" json:"kube_service_refs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                         `json:"-"`
	XXX_unrecognized     []byte                           `json:"-"`
	XXX_sizecache        int32                            `json:"-"`
}

func (m *ServiceSelector) Reset()         { *m = ServiceSelector{} }
func (m *ServiceSelector) String() string { return proto.CompactTextString(m) }
func (*ServiceSelector) ProtoMessage()    {}
func (*ServiceSelector) Descriptor() ([]byte, []int) {
	return fileDescriptor_08e8eca7185dcc82, []int{0}
}
func (m *ServiceSelector) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceSelector.Unmarshal(m, b)
}
func (m *ServiceSelector) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceSelector.Marshal(b, m, deterministic)
}
func (m *ServiceSelector) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceSelector.Merge(m, src)
}
func (m *ServiceSelector) XXX_Size() int {
	return xxx_messageInfo_ServiceSelector.Size(m)
}
func (m *ServiceSelector) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceSelector.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceSelector proto.InternalMessageInfo

func (m *ServiceSelector) GetKubeServiceMatcher() *ServiceSelector_KubeServiceMatcher {
	if m != nil {
		return m.KubeServiceMatcher
	}
	return nil
}

func (m *ServiceSelector) GetKubeServiceRefs() *ServiceSelector_KubeServiceRefs {
	if m != nil {
		return m.KubeServiceRefs
	}
	return nil
}

type ServiceSelector_KubeServiceMatcher struct {
	// If specified, all labels must exist on k8s Service, else match on any labels.
	Labels map[string]string `protobuf:"bytes,1,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// If specified, match k8s Services if they exist in one of the specified namespaces. If not specified, match on any namespace.
	Namespaces []string `protobuf:"bytes,2,rep,name=namespaces,proto3" json:"namespaces,omitempty"`
	// If specified, match k8s Services if they exist in one of the specified clusters. If not specified, match on any cluster.
	Clusters             []string `protobuf:"bytes,3,rep,name=clusters,proto3" json:"clusters,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ServiceSelector_KubeServiceMatcher) Reset()         { *m = ServiceSelector_KubeServiceMatcher{} }
func (m *ServiceSelector_KubeServiceMatcher) String() string { return proto.CompactTextString(m) }
func (*ServiceSelector_KubeServiceMatcher) ProtoMessage()    {}
func (*ServiceSelector_KubeServiceMatcher) Descriptor() ([]byte, []int) {
	return fileDescriptor_08e8eca7185dcc82, []int{0, 0}
}
func (m *ServiceSelector_KubeServiceMatcher) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceSelector_KubeServiceMatcher.Unmarshal(m, b)
}
func (m *ServiceSelector_KubeServiceMatcher) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceSelector_KubeServiceMatcher.Marshal(b, m, deterministic)
}
func (m *ServiceSelector_KubeServiceMatcher) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceSelector_KubeServiceMatcher.Merge(m, src)
}
func (m *ServiceSelector_KubeServiceMatcher) XXX_Size() int {
	return xxx_messageInfo_ServiceSelector_KubeServiceMatcher.Size(m)
}
func (m *ServiceSelector_KubeServiceMatcher) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceSelector_KubeServiceMatcher.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceSelector_KubeServiceMatcher proto.InternalMessageInfo

func (m *ServiceSelector_KubeServiceMatcher) GetLabels() map[string]string {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *ServiceSelector_KubeServiceMatcher) GetNamespaces() []string {
	if m != nil {
		return m.Namespaces
	}
	return nil
}

func (m *ServiceSelector_KubeServiceMatcher) GetClusters() []string {
	if m != nil {
		return m.Clusters
	}
	return nil
}

type ServiceSelector_KubeServiceRefs struct {
	// Match k8s Services by direct reference.
	Services             []*v1.ClusterObjectRef `protobuf:"bytes,1,rep,name=services,proto3" json:"services,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *ServiceSelector_KubeServiceRefs) Reset()         { *m = ServiceSelector_KubeServiceRefs{} }
func (m *ServiceSelector_KubeServiceRefs) String() string { return proto.CompactTextString(m) }
func (*ServiceSelector_KubeServiceRefs) ProtoMessage()    {}
func (*ServiceSelector_KubeServiceRefs) Descriptor() ([]byte, []int) {
	return fileDescriptor_08e8eca7185dcc82, []int{0, 1}
}
func (m *ServiceSelector_KubeServiceRefs) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceSelector_KubeServiceRefs.Unmarshal(m, b)
}
func (m *ServiceSelector_KubeServiceRefs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceSelector_KubeServiceRefs.Marshal(b, m, deterministic)
}
func (m *ServiceSelector_KubeServiceRefs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceSelector_KubeServiceRefs.Merge(m, src)
}
func (m *ServiceSelector_KubeServiceRefs) XXX_Size() int {
	return xxx_messageInfo_ServiceSelector_KubeServiceRefs.Size(m)
}
func (m *ServiceSelector_KubeServiceRefs) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceSelector_KubeServiceRefs.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceSelector_KubeServiceRefs proto.InternalMessageInfo

func (m *ServiceSelector_KubeServiceRefs) GetServices() []*v1.ClusterObjectRef {
	if m != nil {
		return m.Services
	}
	return nil
}

//
//Select Kubernetes workloads directly using label and/or namespace criteria. See comments on the fields for
//detailed semantics.
type WorkloadSelector struct {
	// If specified, all labels must exist on workloads, else match on any labels.
	Labels map[string]string `protobuf:"bytes,1,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// If specified, match workloads if they exist in one of the specified namespaces. If not specified, match on any namespace.
	Namespaces           []string `protobuf:"bytes,2,rep,name=namespaces,proto3" json:"namespaces,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WorkloadSelector) Reset()         { *m = WorkloadSelector{} }
func (m *WorkloadSelector) String() string { return proto.CompactTextString(m) }
func (*WorkloadSelector) ProtoMessage()    {}
func (*WorkloadSelector) Descriptor() ([]byte, []int) {
	return fileDescriptor_08e8eca7185dcc82, []int{1}
}
func (m *WorkloadSelector) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WorkloadSelector.Unmarshal(m, b)
}
func (m *WorkloadSelector) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WorkloadSelector.Marshal(b, m, deterministic)
}
func (m *WorkloadSelector) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WorkloadSelector.Merge(m, src)
}
func (m *WorkloadSelector) XXX_Size() int {
	return xxx_messageInfo_WorkloadSelector.Size(m)
}
func (m *WorkloadSelector) XXX_DiscardUnknown() {
	xxx_messageInfo_WorkloadSelector.DiscardUnknown(m)
}

var xxx_messageInfo_WorkloadSelector proto.InternalMessageInfo

func (m *WorkloadSelector) GetLabels() map[string]string {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *WorkloadSelector) GetNamespaces() []string {
	if m != nil {
		return m.Namespaces
	}
	return nil
}

//
//Selector capable of selecting specific service identities. Useful for binding policy rules.
//Either (namespaces, cluster, service_account_names) or service_accounts can be specified.
//If all fields are omitted, any source identity is permitted.
type IdentitySelector struct {
	// A KubeIdentityMatcher matches request identities based on the k8s namespace and cluster.
	KubeIdentityMatcher *IdentitySelector_KubeIdentityMatcher `protobuf:"bytes,1,opt,name=kube_identity_matcher,json=kubeIdentityMatcher,proto3" json:"kube_identity_matcher,omitempty"`
	// KubeServiceAccountRefs matches request identities based on the k8s service account of request.
	KubeServiceAccountRefs *IdentitySelector_KubeServiceAccountRefs `protobuf:"bytes,2,opt,name=kube_service_account_refs,json=kubeServiceAccountRefs,proto3" json:"kube_service_account_refs,omitempty"`
	XXX_NoUnkeyedLiteral   struct{}                                 `json:"-"`
	XXX_unrecognized       []byte                                   `json:"-"`
	XXX_sizecache          int32                                    `json:"-"`
}

func (m *IdentitySelector) Reset()         { *m = IdentitySelector{} }
func (m *IdentitySelector) String() string { return proto.CompactTextString(m) }
func (*IdentitySelector) ProtoMessage()    {}
func (*IdentitySelector) Descriptor() ([]byte, []int) {
	return fileDescriptor_08e8eca7185dcc82, []int{2}
}
func (m *IdentitySelector) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IdentitySelector.Unmarshal(m, b)
}
func (m *IdentitySelector) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IdentitySelector.Marshal(b, m, deterministic)
}
func (m *IdentitySelector) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IdentitySelector.Merge(m, src)
}
func (m *IdentitySelector) XXX_Size() int {
	return xxx_messageInfo_IdentitySelector.Size(m)
}
func (m *IdentitySelector) XXX_DiscardUnknown() {
	xxx_messageInfo_IdentitySelector.DiscardUnknown(m)
}

var xxx_messageInfo_IdentitySelector proto.InternalMessageInfo

func (m *IdentitySelector) GetKubeIdentityMatcher() *IdentitySelector_KubeIdentityMatcher {
	if m != nil {
		return m.KubeIdentityMatcher
	}
	return nil
}

func (m *IdentitySelector) GetKubeServiceAccountRefs() *IdentitySelector_KubeServiceAccountRefs {
	if m != nil {
		return m.KubeServiceAccountRefs
	}
	return nil
}

type IdentitySelector_KubeIdentityMatcher struct {
	// Namespaces to allow. If not set, any namespace is allowed.
	Namespaces []string `protobuf:"bytes,1,rep,name=namespaces,proto3" json:"namespaces,omitempty"`
	// Cluster to allow. If not set, any cluster is allowed.
	Clusters             []string `protobuf:"bytes,2,rep,name=clusters,proto3" json:"clusters,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IdentitySelector_KubeIdentityMatcher) Reset()         { *m = IdentitySelector_KubeIdentityMatcher{} }
func (m *IdentitySelector_KubeIdentityMatcher) String() string { return proto.CompactTextString(m) }
func (*IdentitySelector_KubeIdentityMatcher) ProtoMessage()    {}
func (*IdentitySelector_KubeIdentityMatcher) Descriptor() ([]byte, []int) {
	return fileDescriptor_08e8eca7185dcc82, []int{2, 0}
}
func (m *IdentitySelector_KubeIdentityMatcher) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IdentitySelector_KubeIdentityMatcher.Unmarshal(m, b)
}
func (m *IdentitySelector_KubeIdentityMatcher) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IdentitySelector_KubeIdentityMatcher.Marshal(b, m, deterministic)
}
func (m *IdentitySelector_KubeIdentityMatcher) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IdentitySelector_KubeIdentityMatcher.Merge(m, src)
}
func (m *IdentitySelector_KubeIdentityMatcher) XXX_Size() int {
	return xxx_messageInfo_IdentitySelector_KubeIdentityMatcher.Size(m)
}
func (m *IdentitySelector_KubeIdentityMatcher) XXX_DiscardUnknown() {
	xxx_messageInfo_IdentitySelector_KubeIdentityMatcher.DiscardUnknown(m)
}

var xxx_messageInfo_IdentitySelector_KubeIdentityMatcher proto.InternalMessageInfo

func (m *IdentitySelector_KubeIdentityMatcher) GetNamespaces() []string {
	if m != nil {
		return m.Namespaces
	}
	return nil
}

func (m *IdentitySelector_KubeIdentityMatcher) GetClusters() []string {
	if m != nil {
		return m.Clusters
	}
	return nil
}

type IdentitySelector_KubeServiceAccountRefs struct {
	// List of ServiceAccounts to allow. If not set, any ServiceAccount is allowed.
	ServiceAccounts      []*v1.ClusterObjectRef `protobuf:"bytes,1,rep,name=service_accounts,json=serviceAccounts,proto3" json:"service_accounts,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *IdentitySelector_KubeServiceAccountRefs) Reset() {
	*m = IdentitySelector_KubeServiceAccountRefs{}
}
func (m *IdentitySelector_KubeServiceAccountRefs) String() string { return proto.CompactTextString(m) }
func (*IdentitySelector_KubeServiceAccountRefs) ProtoMessage()    {}
func (*IdentitySelector_KubeServiceAccountRefs) Descriptor() ([]byte, []int) {
	return fileDescriptor_08e8eca7185dcc82, []int{2, 1}
}
func (m *IdentitySelector_KubeServiceAccountRefs) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IdentitySelector_KubeServiceAccountRefs.Unmarshal(m, b)
}
func (m *IdentitySelector_KubeServiceAccountRefs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IdentitySelector_KubeServiceAccountRefs.Marshal(b, m, deterministic)
}
func (m *IdentitySelector_KubeServiceAccountRefs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IdentitySelector_KubeServiceAccountRefs.Merge(m, src)
}
func (m *IdentitySelector_KubeServiceAccountRefs) XXX_Size() int {
	return xxx_messageInfo_IdentitySelector_KubeServiceAccountRefs.Size(m)
}
func (m *IdentitySelector_KubeServiceAccountRefs) XXX_DiscardUnknown() {
	xxx_messageInfo_IdentitySelector_KubeServiceAccountRefs.DiscardUnknown(m)
}

var xxx_messageInfo_IdentitySelector_KubeServiceAccountRefs proto.InternalMessageInfo

func (m *IdentitySelector_KubeServiceAccountRefs) GetServiceAccounts() []*v1.ClusterObjectRef {
	if m != nil {
		return m.ServiceAccounts
	}
	return nil
}

func init() {
	proto.RegisterType((*ServiceSelector)(nil), "networking.smh.solo.io.ServiceSelector")
	proto.RegisterType((*ServiceSelector_KubeServiceMatcher)(nil), "networking.smh.solo.io.ServiceSelector.KubeServiceMatcher")
	proto.RegisterMapType((map[string]string)(nil), "networking.smh.solo.io.ServiceSelector.KubeServiceMatcher.LabelsEntry")
	proto.RegisterType((*ServiceSelector_KubeServiceRefs)(nil), "networking.smh.solo.io.ServiceSelector.KubeServiceRefs")
	proto.RegisterType((*WorkloadSelector)(nil), "networking.smh.solo.io.WorkloadSelector")
	proto.RegisterMapType((map[string]string)(nil), "networking.smh.solo.io.WorkloadSelector.LabelsEntry")
	proto.RegisterType((*IdentitySelector)(nil), "networking.smh.solo.io.IdentitySelector")
	proto.RegisterType((*IdentitySelector_KubeIdentityMatcher)(nil), "networking.smh.solo.io.IdentitySelector.KubeIdentityMatcher")
	proto.RegisterType((*IdentitySelector_KubeServiceAccountRefs)(nil), "networking.smh.solo.io.IdentitySelector.KubeServiceAccountRefs")
}

func init() {
	proto.RegisterFile("github.com/solo-io/service-mesh-hub/api/networking/v1alpha2/selectors.proto", fileDescriptor_08e8eca7185dcc82)
}

var fileDescriptor_08e8eca7185dcc82 = []byte{
	// 547 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0xd1, 0x6e, 0xd3, 0x30,
	0x14, 0x55, 0x56, 0x98, 0xd6, 0xdb, 0x87, 0x16, 0xaf, 0x54, 0x25, 0x48, 0xd5, 0x54, 0x5e, 0xfa,
	0xd2, 0x44, 0x2b, 0x48, 0xc0, 0x84, 0x34, 0x01, 0x02, 0x09, 0x6d, 0x80, 0x96, 0x3d, 0x20, 0xf1,
	0xc0, 0xe4, 0x78, 0xb7, 0x49, 0x48, 0x1a, 0x47, 0xb6, 0xd3, 0xa9, 0x7c, 0x10, 0xe2, 0x1f, 0x10,
	0x3f, 0xc3, 0x2b, 0x3f, 0x81, 0xea, 0xa4, 0xa1, 0x75, 0x3b, 0xc4, 0xca, 0x53, 0x9c, 0x63, 0xfb,
	0xdc, 0x73, 0x8f, 0x8f, 0x2e, 0x9c, 0x04, 0x91, 0x0a, 0x73, 0xdf, 0x61, 0x7c, 0xe2, 0x4a, 0x9e,
	0xf0, 0x61, 0xc4, 0x5d, 0x89, 0x62, 0x1a, 0x31, 0x1c, 0x4e, 0x50, 0x86, 0xc3, 0x30, 0xf7, 0x5d,
	0x9a, 0x45, 0x6e, 0x8a, 0xea, 0x8a, 0x8b, 0x38, 0x4a, 0x03, 0x77, 0x7a, 0x48, 0x93, 0x2c, 0xa4,
	0x23, 0x57, 0x62, 0x82, 0x4c, 0x71, 0x21, 0x9d, 0x4c, 0x70, 0xc5, 0x49, 0xe7, 0xcf, 0x21, 0x47,
	0x4e, 0x42, 0x67, 0x4e, 0xe8, 0x44, 0xdc, 0xee, 0x05, 0x9c, 0x07, 0x09, 0xba, 0xfa, 0x94, 0x9f,
	0x8f, 0xdd, 0x2b, 0x41, 0xb3, 0x0c, 0x17, 0xf7, 0xec, 0xfb, 0x32, 0x9e, 0x8e, 0x74, 0x15, 0xc6,
	0x05, 0xba, 0xd3, 0x43, 0xfd, 0x2d, 0x37, 0xdb, 0x01, 0x0f, 0xb8, 0x5e, 0xba, 0xf3, 0x55, 0x81,
	0xf6, 0xbf, 0xde, 0x82, 0xe6, 0x79, 0xa1, 0xf3, 0xbc, 0x54, 0x41, 0x12, 0x68, 0xc7, 0xb9, 0x8f,
	0x17, 0xa5, 0xfe, 0x8b, 0x09, 0x55, 0x2c, 0x44, 0xd1, 0xb5, 0x0e, 0xac, 0x41, 0x63, 0x74, 0xe4,
	0x6c, 0x56, 0xe7, 0x18, 0x34, 0xce, 0x49, 0xee, 0x63, 0x89, 0xbd, 0x2d, 0x18, 0x3c, 0x12, 0xaf,
	0x61, 0x84, 0xc1, 0x9d, 0x95, 0x6a, 0x02, 0xc7, 0xb2, 0xbb, 0xa3, 0x4b, 0x3d, 0xde, 0xa2, 0x94,
	0x87, 0x63, 0xe9, 0x35, 0xe3, 0x55, 0xc0, 0xfe, 0x65, 0x01, 0x59, 0xd7, 0x43, 0x3e, 0xc1, 0x6e,
	0x42, 0x7d, 0x4c, 0x64, 0xd7, 0x3a, 0xa8, 0x0d, 0x1a, 0xa3, 0xd7, 0xdb, 0xf7, 0xe6, 0x9c, 0x6a,
	0xa2, 0x57, 0xa9, 0x12, 0x33, 0xaf, 0x64, 0x25, 0x3d, 0x80, 0x94, 0x4e, 0x50, 0x66, 0x94, 0xe1,
	0xbc, 0xa9, 0xda, 0xa0, 0xee, 0x2d, 0x21, 0xc4, 0x86, 0x3d, 0x96, 0xe4, 0x52, 0xa1, 0x90, 0xdd,
	0x9a, 0xde, 0xad, 0xfe, 0xed, 0xa7, 0xd0, 0x58, 0xa2, 0x24, 0x2d, 0xa8, 0xc5, 0x38, 0xd3, 0x6f,
	0x50, 0xf7, 0xe6, 0x4b, 0xd2, 0x86, 0xdb, 0x53, 0x9a, 0xe4, 0xa8, 0xcd, 0xaa, 0x7b, 0xc5, 0xcf,
	0xd1, 0xce, 0x13, 0xcb, 0xf6, 0xa0, 0x69, 0x38, 0x42, 0x8e, 0x61, 0xaf, 0x34, 0x78, 0xd1, 0xeb,
	0x03, 0x47, 0x87, 0x63, 0x1e, 0x99, 0xaa, 0xcd, 0x97, 0x45, 0xf1, 0xf7, 0xfe, 0x67, 0x64, 0xca,
	0xc3, 0xb1, 0x57, 0x5d, 0xea, 0x7f, 0xb7, 0xa0, 0xf5, 0x81, 0x8b, 0x38, 0xe1, 0xf4, 0xb2, 0x4a,
	0xca, 0xa9, 0xe1, 0xdf, 0xa3, 0xeb, 0xfc, 0x33, 0x6f, 0x6e, 0xe3, 0xd6, 0x7f, 0x38, 0xd2, 0xff,
	0x51, 0x83, 0xd6, 0x9b, 0x4b, 0x4c, 0x55, 0xa4, 0x66, 0x95, 0xfa, 0x0c, 0xee, 0xea, 0xe4, 0x45,
	0xe5, 0x86, 0x11, 0xf4, 0x67, 0xd7, 0x35, 0x63, 0x12, 0xe9, 0x34, 0x2c, 0xc0, 0x45, 0xd4, 0xf7,
	0xe3, 0x75, 0x90, 0x7c, 0x81, 0x7b, 0x2b, 0x59, 0xa7, 0x8c, 0xf1, 0x3c, 0x55, 0xcb, 0x99, 0x3f,
	0xbe, 0x51, 0xd5, 0xf2, 0x89, 0x9f, 0x17, 0x3c, 0x3a, 0xfb, 0x9d, 0x78, 0x23, 0x6e, 0x9f, 0xc1,
	0xfe, 0x06, 0x9d, 0x86, 0xe9, 0xd6, 0x5f, 0x23, 0xba, 0x63, 0x44, 0x34, 0x84, 0xce, 0x66, 0x11,
	0xe4, 0x1d, 0xb4, 0x8c, 0x1e, 0x6f, 0x14, 0xbb, 0xa6, 0x5c, 0xa1, 0x94, 0x2f, 0xce, 0xbe, 0xfd,
	0xec, 0x59, 0x1f, 0xff, 0x69, 0xc8, 0x66, 0x71, 0x60, 0x0c, 0xda, 0x65, 0x1b, 0xab, 0xa1, 0xeb,
	0xef, 0xea, 0x01, 0xf8, 0xf0, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0xe1, 0x9e, 0xb3, 0x5f, 0xba,
	0x05, 0x00, 0x00,
}

func (this *ServiceSelector) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ServiceSelector)
	if !ok {
		that2, ok := that.(ServiceSelector)
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
	if !this.KubeServiceMatcher.Equal(that1.KubeServiceMatcher) {
		return false
	}
	if !this.KubeServiceRefs.Equal(that1.KubeServiceRefs) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *ServiceSelector_KubeServiceMatcher) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ServiceSelector_KubeServiceMatcher)
	if !ok {
		that2, ok := that.(ServiceSelector_KubeServiceMatcher)
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
	if len(this.Labels) != len(that1.Labels) {
		return false
	}
	for i := range this.Labels {
		if this.Labels[i] != that1.Labels[i] {
			return false
		}
	}
	if len(this.Namespaces) != len(that1.Namespaces) {
		return false
	}
	for i := range this.Namespaces {
		if this.Namespaces[i] != that1.Namespaces[i] {
			return false
		}
	}
	if len(this.Clusters) != len(that1.Clusters) {
		return false
	}
	for i := range this.Clusters {
		if this.Clusters[i] != that1.Clusters[i] {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *ServiceSelector_KubeServiceRefs) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ServiceSelector_KubeServiceRefs)
	if !ok {
		that2, ok := that.(ServiceSelector_KubeServiceRefs)
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
	if len(this.Services) != len(that1.Services) {
		return false
	}
	for i := range this.Services {
		if !this.Services[i].Equal(that1.Services[i]) {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *WorkloadSelector) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*WorkloadSelector)
	if !ok {
		that2, ok := that.(WorkloadSelector)
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
	if len(this.Labels) != len(that1.Labels) {
		return false
	}
	for i := range this.Labels {
		if this.Labels[i] != that1.Labels[i] {
			return false
		}
	}
	if len(this.Namespaces) != len(that1.Namespaces) {
		return false
	}
	for i := range this.Namespaces {
		if this.Namespaces[i] != that1.Namespaces[i] {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *IdentitySelector) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*IdentitySelector)
	if !ok {
		that2, ok := that.(IdentitySelector)
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
	if !this.KubeIdentityMatcher.Equal(that1.KubeIdentityMatcher) {
		return false
	}
	if !this.KubeServiceAccountRefs.Equal(that1.KubeServiceAccountRefs) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *IdentitySelector_KubeIdentityMatcher) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*IdentitySelector_KubeIdentityMatcher)
	if !ok {
		that2, ok := that.(IdentitySelector_KubeIdentityMatcher)
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
	if len(this.Namespaces) != len(that1.Namespaces) {
		return false
	}
	for i := range this.Namespaces {
		if this.Namespaces[i] != that1.Namespaces[i] {
			return false
		}
	}
	if len(this.Clusters) != len(that1.Clusters) {
		return false
	}
	for i := range this.Clusters {
		if this.Clusters[i] != that1.Clusters[i] {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *IdentitySelector_KubeServiceAccountRefs) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*IdentitySelector_KubeServiceAccountRefs)
	if !ok {
		that2, ok := that.(IdentitySelector_KubeServiceAccountRefs)
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
	if len(this.ServiceAccounts) != len(that1.ServiceAccounts) {
		return false
	}
	for i := range this.ServiceAccounts {
		if !this.ServiceAccounts[i].Equal(that1.ServiceAccounts[i]) {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
