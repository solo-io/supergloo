package io

import (
	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	smiaccessv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	smispecsv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	smislpitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	"github.com/solo-io/gloo-mesh/codegen/constants"
	skv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	NetworkingInputTypes = Snapshot{
		schema.GroupVersion{
			Group:   "discovery." + constants.GlooMeshApiGroupSuffix,
			Version: "v1",
		}: {
			"Mesh",
			"Workload",
			"Destination",
		},
		schema.GroupVersion{
			Group:   "networking." + constants.GlooMeshApiGroupSuffix,
			Version: "v1",
		}: {
			"TrafficPolicy",
			"AccessPolicy",
			"VirtualMesh",
		},
		schema.GroupVersion{
			Group:   "settings." + constants.GlooMeshApiGroupSuffix,
			Version: "v1",
		}: {
			"Settings",
		},
		schema.GroupVersion{
			Group:   "certificates." + constants.GlooMeshApiGroupSuffix,
			Version: "v1",
		}: {
			"IssuedCertificate",
		},
		schema.GroupVersion{
			Group:   "networking.enterprise." + constants.GlooMeshApiGroupSuffix,
			Version: "v1beta1",
		}: {
			"WasmDeployment",
			"VirtualDestination",
			"ServiceDependency",
		},
		schema.GroupVersion{
			Group:   "observability.enterprise." + constants.GlooMeshApiGroupSuffix,
			Version: "v1",
		}: {
			"AccessLogRecord",
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
				"Sidecar",
			},
			istiosecurityv1beta1.SchemeGroupVersion: {
				"AuthorizationPolicy",
			},
			schema.GroupVersion{
				Group:   "certificates." + constants.GlooMeshApiGroupSuffix,
				Version: "v1",
			}: {
				"IssuedCertificate",
				"PodBounceDirective",
			},
			schema.GroupVersion{
				Group:   "xds.agent.enterprise." + constants.GlooMeshApiGroupSuffix,
				Version: "v1beta1",
			}: {
				"XdsConfig",
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
				"VirtualService",
			},
		},
	}
)
