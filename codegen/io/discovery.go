package io

import (
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
