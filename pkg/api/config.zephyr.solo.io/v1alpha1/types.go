// Definitions for the Kubernetes types
package v1alpha1

import (
	. "github.com/solo-io/mesh-projects/pkg/api/config.zephyr.solo.io/v1alpha1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status

// RoutingRule is the Schema for the routingRule API
type RoutingRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoutingRuleSpec   `json:"spec,omitempty"`
	Status RoutingRuleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoutingRuleList contains a list of RoutingRule
type RoutingRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoutingRule `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status

// SecurityRule is the Schema for the securityRule API
type SecurityRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecurityRuleSpec   `json:"spec,omitempty"`
	Status SecurityRuleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SecurityRuleList contains a list of SecurityRule
type SecurityRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RoutingRule{}, &RoutingRuleList{})
	SchemeBuilder.Register(&SecurityRule{}, &SecurityRuleList{})
}
