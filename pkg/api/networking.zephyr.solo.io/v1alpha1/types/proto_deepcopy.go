// This file contains generated Deepcopy methods for
// Protobuf types with oneofs

package types

import (
	fmt "fmt"
	math "math"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *AccessControlPolicySpec) DeepCopyInto(out *AccessControlPolicySpec) {
	p := proto.Clone(in).(*AccessControlPolicySpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *AccessControlPolicyStatus) DeepCopyInto(out *AccessControlPolicyStatus) {
	p := proto.Clone(in).(*AccessControlPolicyStatus)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *TrafficPolicySpec) DeepCopyInto(out *TrafficPolicySpec) {
	p := proto.Clone(in).(*TrafficPolicySpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *TrafficPolicyStatus) DeepCopyInto(out *TrafficPolicyStatus) {
	p := proto.Clone(in).(*TrafficPolicyStatus)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *VirtualMeshSpec) DeepCopyInto(out *VirtualMeshSpec) {
	p := proto.Clone(in).(*VirtualMeshSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *VirtualMeshStatus) DeepCopyInto(out *VirtualMeshStatus) {
	p := proto.Clone(in).(*VirtualMeshStatus)
	*out = *p
}
