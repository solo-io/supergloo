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
func (in *SecurityRuleSpec) DeepCopyInto(out *SecurityRuleSpec) {
	p := proto.Clone(in).(*SecurityRuleSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *SecurityRuleStatus) DeepCopyInto(out *SecurityRuleStatus) {
	p := proto.Clone(in).(*SecurityRuleStatus)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *RoutingRuleSpec) DeepCopyInto(out *RoutingRuleSpec) {
	p := proto.Clone(in).(*RoutingRuleSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *RoutingRuleStatus) DeepCopyInto(out *RoutingRuleStatus) {
	p := proto.Clone(in).(*RoutingRuleStatus)
	*out = *p
}
