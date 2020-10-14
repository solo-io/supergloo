package io

import (
	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	smiaccessv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	smispecsv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	smislpitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	"github.com/solo-io/service-mesh-hub/codegen/constants"
	skv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
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
			"Workload",
			"TrafficTarget",
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
		schema.GroupVersion{
			Group:   "settings." + constants.ServiceMeshHubApiGroupSuffix,
			Version: "v1alpha2",
		}: {
			"Settings",
		},
		skv1alpha1.SchemeGroupVersion: {
			"KubernetesCluster",
		},
		corev1.SchemeGroupVersion: {
			"Secret",
		},
	}

	IstioNetworkingOutputTypes = OutputSnapshot{
		Name: "istio",
		Snapshot: Snapshot{
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
				"PodBounceDirective",
			},
		},
	}

	LocalNetworkingOutputTypes = OutputSnapshot{
		Name: "local",
		Snapshot: Snapshot{
			corev1.SchemeGroupVersion: {
				"Secret",
			},
		},
	}

	SmiNetworkingOutputTypes = OutputSnapshot{
		Name: "smi",
		Snapshot: Snapshot{
			smislpitv1alpha2.SchemeGroupVersion: {
				"TrafficSplit",
			},
			smiaccessv1alpha2.SchemeGroupVersion: {
				"TrafficTarget",
			},
			smispecsv1alpha3.SchemeGroupVersion: {
				"HTTPRouteGroup",
			},
		},
	}

	AppMeshNetworkingOutputTypes = OutputSnapshot{
		Name: "appmesh",
		Snapshot: Snapshot{
			appmeshv1beta2.GroupVersion: {
				"VirtualNode",
				"VirtualRouter",
				"Route",
				"VirtualService",
			},
		},
	}
)
