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
func (in *VirtualMeshCertificateSigningRequestSpec) DeepCopyInto(out *VirtualMeshCertificateSigningRequestSpec) {
	p := proto.Clone(in).(*VirtualMeshCertificateSigningRequestSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *VirtualMeshCertificateSigningRequestStatus) DeepCopyInto(out *VirtualMeshCertificateSigningRequestStatus) {
	p := proto.Clone(in).(*VirtualMeshCertificateSigningRequestStatus)
	*out = *p
}
