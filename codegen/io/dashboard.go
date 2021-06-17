package io

import (
	"github.com/solo-io/gloo-mesh/codegen/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	DashboardInputTypes = Snapshot{
		schema.GroupVersion{
			Group:   "settings." + constants.GlooMeshApiGroupSuffix,
			Version: "v1",
		}: {
			"Dashboard",
		},
		corev1.SchemeGroupVersion: {
			"Secret",
		},
	}
)
