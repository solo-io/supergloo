// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: github.com/solo-io/gloo-mesh/api/enterprise/networking/v1beta1/edge.proto

package v1beta1

import (
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	_ "github.com/solo-io/protoc-gen-ext/extproto"
	v11 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

//
//TODO(harveyxia): Explain and motivate.
type ServiceDependencySpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Select the traffic sources (i.e. Workloads) for this network ServiceDependency.
	SourceSelectors []*v1.WorkloadSelector `protobuf:"bytes,1,rep,name=source_selectors,json=sourceSelectors,proto3" json:"source_selectors,omitempty"`
	// Select the traffic destinations (i.e. Destination) for this network ServiceDependency.
	DestinationSelectors []*v1.DestinationSelector `protobuf:"bytes,2,rep,name=destination_selectors,json=destinationSelectors,proto3" json:"destination_selectors,omitempty"`
}

func (x *ServiceDependencySpec) Reset() {
	*x = ServiceDependencySpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceDependencySpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceDependencySpec) ProtoMessage() {}

func (x *ServiceDependencySpec) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceDependencySpec.ProtoReflect.Descriptor instead.
func (*ServiceDependencySpec) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDescGZIP(), []int{0}
}

func (x *ServiceDependencySpec) GetSourceSelectors() []*v1.WorkloadSelector {
	if x != nil {
		return x.SourceSelectors
	}
	return nil
}

func (x *ServiceDependencySpec) GetDestinationSelectors() []*v1.DestinationSelector {
	if x != nil {
		return x.DestinationSelectors
	}
	return nil
}

type ServiceDependencyStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The most recent generation observed in the the ServiceDependency metadata.
	// If the observedGeneration does not match `metadata.generation`, Gloo Mesh has not processed the most
	// recent version of this resource.
	ObservedGeneration int64 `protobuf:"varint,1,opt,name=observed_generation,json=observedGeneration,proto3" json:"observed_generation,omitempty"`
	// The state of the overall resource, will only show accepted if it has been successfully
	// applied to all exported to Meshes.
	State v1.ApprovalState `protobuf:"varint,2,opt,name=state,proto3,enum=common.mesh.gloo.solo.io.ApprovalState" json:"state,omitempty"`
	// The Workloads that this ServiceDependency applies to.
	Workloads []*v11.ClusterObjectRef `protobuf:"bytes,3,rep,name=workloads,proto3" json:"workloads,omitempty"`
	// Any errors found while processing this generation of the resource.
	Errors []string `protobuf:"bytes,4,rep,name=errors,proto3" json:"errors,omitempty"`
}

func (x *ServiceDependencyStatus) Reset() {
	*x = ServiceDependencyStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceDependencyStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceDependencyStatus) ProtoMessage() {}

func (x *ServiceDependencyStatus) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceDependencyStatus.ProtoReflect.Descriptor instead.
func (*ServiceDependencyStatus) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDescGZIP(), []int{1}
}

func (x *ServiceDependencyStatus) GetObservedGeneration() int64 {
	if x != nil {
		return x.ObservedGeneration
	}
	return 0
}

func (x *ServiceDependencyStatus) GetState() v1.ApprovalState {
	if x != nil {
		return x.State
	}
	return v1.ApprovalState_PENDING
}

func (x *ServiceDependencyStatus) GetWorkloads() []*v11.ClusterObjectRef {
	if x != nil {
		return x.Workloads
	}
	return nil
}

func (x *ServiceDependencyStatus) GetErrors() []string {
	if x != nil {
		return x.Errors
	}
	return nil
}

var File_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto protoreflect.FileDescriptor

var file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDesc = []byte{
	0x0a, 0x49, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c,
	0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x67, 0x6c, 0x6f, 0x6f, 0x2d, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x65, 0x6e, 0x74, 0x65, 0x72, 0x70, 0x72, 0x69, 0x73, 0x65, 0x2f, 0x6e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31,
	0x2f, 0x65, 0x64, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x27, 0x6e, 0x65, 0x74,
	0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x2e, 0x65, 0x6e, 0x74, 0x65, 0x72, 0x70, 0x72, 0x69,
	0x73, 0x65, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x67, 0x6c, 0x6f, 0x6f, 0x2e, 0x73, 0x6f, 0x6c,
	0x6f, 0x2e, 0x69, 0x6f, 0x1a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6b, 0x76, 0x32, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x67, 0x6c, 0x6f, 0x6f, 0x2d, 0x6d, 0x65,
	0x73, 0x68, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31,
	0x2f, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x41, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c,
	0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x67, 0x6c, 0x6f, 0x6f, 0x2d, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x12, 0x65, 0x78, 0x74, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x78,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd2, 0x01, 0x0a, 0x15, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x44, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63, 0x79, 0x53, 0x70, 0x65,
	0x63, 0x12, 0x55, 0x0a, 0x10, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x73, 0x65, 0x6c, 0x65,
	0x63, 0x74, 0x6f, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x67, 0x6c, 0x6f, 0x6f, 0x2e, 0x73,
	0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x53,
	0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x0f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53,
	0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x12, 0x62, 0x0a, 0x15, 0x64, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x67, 0x6c, 0x6f, 0x6f, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e,
	0x69, 0x6f, 0x2e, 0x44, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65,
	0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x14, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x22, 0xe4, 0x01, 0x0a,
	0x17, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x44, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e,
	0x63, 0x79, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x2f, 0x0a, 0x13, 0x6f, 0x62, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x64, 0x5f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x12, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x47,
	0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3d, 0x0a, 0x05, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x27, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x67, 0x6c, 0x6f, 0x6f, 0x2e, 0x73, 0x6f, 0x6c, 0x6f,
	0x2e, 0x69, 0x6f, 0x2e, 0x41, 0x70, 0x70, 0x72, 0x6f, 0x76, 0x61, 0x6c, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x41, 0x0a, 0x09, 0x77, 0x6f, 0x72, 0x6b,
	0x6c, 0x6f, 0x61, 0x64, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x63, 0x6f,
	0x72, 0x65, 0x2e, 0x73, 0x6b, 0x76, 0x32, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e,
	0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x66,
	0x52, 0x09, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x73, 0x42, 0x5a, 0x5a, 0x54, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x67, 0x6c, 0x6f, 0x6f, 0x2d, 0x6d,
	0x65, 0x73, 0x68, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x2e, 0x65, 0x6e, 0x74, 0x65, 0x72, 0x70, 0x72, 0x69, 0x73,
	0x65, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x67, 0x6c, 0x6f, 0x6f, 0x2e, 0x73, 0x6f, 0x6c, 0x6f,
	0x2e, 0x69, 0x6f, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0xc0, 0xf5, 0x04, 0x01, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDescOnce sync.Once
	file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDescData = file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDesc
)

func file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDescGZIP() []byte {
	file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDescOnce.Do(func() {
		file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDescData)
	})
	return file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDescData
}

var file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_goTypes = []interface{}{
	(*ServiceDependencySpec)(nil),   // 0: networking.enterprise.mesh.gloo.solo.io.ServiceDependencySpec
	(*ServiceDependencyStatus)(nil), // 1: networking.enterprise.mesh.gloo.solo.io.ServiceDependencyStatus
	(*v1.WorkloadSelector)(nil),     // 2: common.mesh.gloo.solo.io.WorkloadSelector
	(*v1.DestinationSelector)(nil),  // 3: common.mesh.gloo.solo.io.DestinationSelector
	(v1.ApprovalState)(0),           // 4: common.mesh.gloo.solo.io.ApprovalState
	(*v11.ClusterObjectRef)(nil),    // 5: core.skv2.solo.io.ClusterObjectRef
}
var file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_depIdxs = []int32{
	2, // 0: networking.enterprise.mesh.gloo.solo.io.ServiceDependencySpec.source_selectors:type_name -> common.mesh.gloo.solo.io.WorkloadSelector
	3, // 1: networking.enterprise.mesh.gloo.solo.io.ServiceDependencySpec.destination_selectors:type_name -> common.mesh.gloo.solo.io.DestinationSelector
	4, // 2: networking.enterprise.mesh.gloo.solo.io.ServiceDependencyStatus.state:type_name -> common.mesh.gloo.solo.io.ApprovalState
	5, // 3: networking.enterprise.mesh.gloo.solo.io.ServiceDependencyStatus.workloads:type_name -> core.skv2.solo.io.ClusterObjectRef
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_init() }
func file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_init() {
	if File_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceDependencySpec); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceDependencyStatus); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_goTypes,
		DependencyIndexes: file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_depIdxs,
		MessageInfos:      file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_msgTypes,
	}.Build()
	File_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto = out.File
	file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_rawDesc = nil
	file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_goTypes = nil
	file_github_com_solo_io_gloo_mesh_api_enterprise_networking_v1beta1_edge_proto_depIdxs = nil
}
