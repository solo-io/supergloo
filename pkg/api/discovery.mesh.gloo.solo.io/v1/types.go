// Code generated by skv2. DO NOT EDIT.

// Definitions for the Kubernetes types
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status

// GroupVersionKind for Destination
var DestinationGVK = schema.GroupVersionKind{
	Group:   "discovery.mesh.gloo.solo.io",
	Version: "v1",
	Kind:    "Destination",
}

// Destination is the Schema for the destination API
type Destination struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DestinationSpec   `json:"spec,omitempty"`
	Status DestinationStatus `json:"status,omitempty"`
}

// GVK returns the GroupVersionKind associated with the resource type.
func (Destination) GVK() schema.GroupVersionKind {
	return DestinationGVK
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DestinationList contains a list of Destination
type DestinationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Destination `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status

// GroupVersionKind for Workload
var WorkloadGVK = schema.GroupVersionKind{
	Group:   "discovery.mesh.gloo.solo.io",
	Version: "v1",
	Kind:    "Workload",
}

// Workload is the Schema for the workload API
type Workload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkloadSpec   `json:"spec,omitempty"`
	Status WorkloadStatus `json:"status,omitempty"`
}

// GVK returns the GroupVersionKind associated with the resource type.
func (Workload) GVK() schema.GroupVersionKind {
	return WorkloadGVK
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkloadList contains a list of Workload
type WorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workload `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status

// GroupVersionKind for Mesh
var MeshGVK = schema.GroupVersionKind{
	Group:   "discovery.mesh.gloo.solo.io",
	Version: "v1",
	Kind:    "Mesh",
}

// Mesh is the Schema for the mesh API
type Mesh struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeshSpec   `json:"spec,omitempty"`
	Status MeshStatus `json:"status,omitempty"`
}

// GVK returns the GroupVersionKind associated with the resource type.
func (Mesh) GVK() schema.GroupVersionKind {
	return MeshGVK
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MeshList contains a list of Mesh
type MeshList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mesh `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Destination{}, &DestinationList{})
	SchemeBuilder.Register(&Workload{}, &WorkloadList{})
	SchemeBuilder.Register(&Mesh{}, &MeshList{})
}
