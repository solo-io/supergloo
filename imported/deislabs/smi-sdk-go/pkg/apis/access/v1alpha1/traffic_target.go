package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TrafficTarget associates a set of traffic definitions (rules) with a service identity which is allocated to a group of pods.
// Access is controlled via referenced TrafficSpecs and by a list of source service identities.
// * If a pod which holds the referenced service identity makes a call to the destination on one of the defined routes then access
//   will be allowed
// * Any pod which attempts to connect and is not in the defined list of sources will be denied
// * Any pod which is in the defined list, but attempts to connect on a route which is not in the list of the
//   TrafficSpecs will be denied
type TrafficTarget struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Most recently observed status of the object.
	// This data may not be up to date.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Status Status `json:"status,omitempty" protobuf:"bytes,2,opt,name=status"`

	// Selector is the pod or group of pods to allow ingress traffic
	Destination IdentityBindingSubject `json:"destination" protobuf:"bytes,3,name=destination"`

	// Sources are the pod or group of pods to allow ingress traffic
	Sources []IdentityBindingSubject `json:"sources" protobuf:"bytes,4,name=sources"`

	// Rules are the traffic rules to allow (HTTPRoutes | TCPRoute),
	Specs []TrafficTargetSpec `json:"specs" protobuf:"bytes,5,name=specs"`
}

// TrafficTargetSpec is the TrafficSpec to allow for a TrafficTarget
type TrafficTargetSpec struct {
	// Kind is the kind of TrafficSpec to allow
	Kind string `json:"kind" protobuf:"bytes,1,name=kind"`
	// Name of the TrafficSpec to use
	Name string `json:"name" protobuf:"bytes,2,name=name"`
	// Matches is a list of TrafficSpec routes to allow traffic for
	Matches []string `json:"matches,omitempty" protobuf:"bytes,3,opt,name=matches"`
}

// IdentityBindingSubject is a Kubernetes objects which should be allowed access to the TrafficTarget
type IdentityBindingSubject struct {
	// Kind is the type of Subject to allow ingress (ServiceAccount | Group)
	Kind string `json:"kind" protobuf:"bytes,1,name=kind"`
	// Name of the Subject, i.e. ServiceAccountName
	Name string `json:"name" protobuf:"bytes,2,name=name"`
	// Namespace where the Subject is deployed
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
	// Port defines a TCP port to apply the TrafficTarget to
	Port string `json:"port,omitempty" protobuf:"bytes,4,opt,name=port"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TrafficTargetList satisfy K8s code gen requirements
type TrafficTargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TrafficTarget `json:"items"`
}
