package io

import (
	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/solo-io/service-mesh-hub/codegen/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	DiscoveryInputTypes = Snapshot{
		corev1.SchemeGroupVersion: {
			"Pod",
			"Service",
			"ConfigMap",
			"Node",
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

	DiscoveryOutputTypes = OutputSnapshot{
		Name: "discovery",
		Snapshot: Snapshot{
			schema.GroupVersion{
				Group:   "discovery." + constants.ServiceMeshHubApiGroupSuffix,
				Version: "v1alpha2",
			}: {
				"Mesh",
				"Workload",
				"TrafficTarget",
			},
		},
	}
)
