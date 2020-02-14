// Definitions for the Kubernetes types
package v1alpha1

import (
	. "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status

// MeshGroupCertificateSigningRequest is the Schema for the meshGroupCertificateSigningRequest API
type MeshGroupCertificateSigningRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeshGroupCertificateSigningRequestSpec   `json:"spec,omitempty"`
	Status MeshGroupCertificateSigningRequestStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MeshGroupCertificateSigningRequestList contains a list of MeshGroupCertificateSigningRequest
type MeshGroupCertificateSigningRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeshGroupCertificateSigningRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MeshGroupCertificateSigningRequest{}, &MeshGroupCertificateSigningRequestList{})
}
