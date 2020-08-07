package io

import (
	"github.com/solo-io/service-mesh-hub/codegen/constants"
	skv1alpha1 "github.com/solo-io/skv2/api/multicluster/v1alpha1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	NetworkingInputTypes = Snapshot{
		schema.GroupVersion{
			Group:   "discovery." + constants.ServiceMeshHubApiGroupSuffix,
			Version: "v1alpha2",
		}: {
			"Mesh",
			"MeshWorkload",
			"MeshService",
		},
		schema.GroupVersion{
			Group:   "networking." + constants.ServiceMeshHubApiGroupSuffix,
			Version: "v1alpha2",
		}: {
			"TrafficPolicy",
			"AccessPolicy",
			"VirtualMesh",
			"FailoverService",
		},
		skv1alpha1.Group.GroupVersion: {
			"KubernetesCluster",
		},
		corev1.SchemeGroupVersion: {
			"Secret",
		},
	}

	NetworkingOutputTypes = Snapshot{
		istionetworkingv1alpha3.SchemeGroupVersion: {
			"DestinationRule",
			"VirtualService",
			"EnvoyFilter",
			"ServiceEntry",
			"Gateway",
		},
		istiosecurityv1beta1.SchemeGroupVersion: {
			"AuthorizationPolicy",
		},
		schema.GroupVersion{
			Group:   "certificates." + constants.ServiceMeshHubApiGroupSuffix,
			Version: "v1alpha2",
		}: {
			"IssuedCertificate",
		},
		corev1.SchemeGroupVersion: {
			"Secret",
		},
	}
)
