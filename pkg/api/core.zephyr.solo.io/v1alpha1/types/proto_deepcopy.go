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
func (in *KubernetesClusterSpec) DeepCopyInto(out *KubernetesClusterSpec) {
	p := proto.Clone(in).(*KubernetesClusterSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *MeshSpec) DeepCopyInto(out *MeshSpec) {
	p := proto.Clone(in).(*MeshSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *MeshStatus) DeepCopyInto(out *MeshStatus) {
	p := proto.Clone(in).(*MeshStatus)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *MeshGroupSpec) DeepCopyInto(out *MeshGroupSpec) {
	p := proto.Clone(in).(*MeshGroupSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *MeshGroupStatus) DeepCopyInto(out *MeshGroupStatus) {
	p := proto.Clone(in).(*MeshGroupStatus)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *MeshServiceSpec) DeepCopyInto(out *MeshServiceSpec) {
	p := proto.Clone(in).(*MeshServiceSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *MeshServiceStatus) DeepCopyInto(out *MeshServiceStatus) {
	p := proto.Clone(in).(*MeshServiceStatus)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *MeshWorkloadSpec) DeepCopyInto(out *MeshWorkloadSpec) {
	p := proto.Clone(in).(*MeshWorkloadSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *MeshWorkloadStatus) DeepCopyInto(out *MeshWorkloadStatus) {
	p := proto.Clone(in).(*MeshWorkloadStatus)
	*out = *p
}
