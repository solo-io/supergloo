package io

import (
	"github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var (
	ClusterWatcherInputTypes = Snapshot{
		corev1.SchemeGroupVersion: {
			"Secret",
		},
		v1alpha1.SchemeGroupVersion: {
			"KubernetesCluster",
		},
	}
)
