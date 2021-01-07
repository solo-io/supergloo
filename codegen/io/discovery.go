package io

import (
	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/solo-io/gloo-mesh/codegen/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	DiscoveryRemoteInputTypes = Snapshot{
		corev1.SchemeGroupVersion: {
			"Pod",
			"Service",
			"ConfigMap",
			"Node",
			"Endpoints",
		},
		appsv1.SchemeGroupVersion: {
			"Deployment",
			"ReplicaSet",
			"DaemonSet",
			"StatefulSet",
		},
		appmeshv1beta2.GroupVersion: {
			"Mesh",
		},
	}

	DiscoveryLocalInputTypes = Snapshot{
		schema.GroupVersion{
			Group:   "settings." + constants.GlooMeshApiGroupSuffix,
			Version: "v1alpha2",
		}: {
			"Settings",
		},
		schema.GroupVersion{
			Group:   "networking." + constants.GlooMeshApiGroupSuffix,
			Version: "v1alpha2",
		}: {
			"VirtualMesh",
		},
	}

	DiscoveryOutputTypes = OutputSnapshot{
		Name: "discovery",
		Snapshot: Snapshot{
			schema.GroupVersion{
				Group:   "discovery." + constants.GlooMeshApiGroupSuffix,
				Version: "v1alpha2",
			}: {
				"Mesh",
				"Workload",
				"TrafficTarget",
			},
		},
	}
)
