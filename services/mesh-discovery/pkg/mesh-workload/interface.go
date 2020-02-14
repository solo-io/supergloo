package mesh_workload

import (
	"context"

	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./interface.go -destination mocks/mock_interface.go

type OwnerFetcher interface {
	GetDeployment(ctx context.Context, pod *corev1.Pod) (*appsv1.Deployment, error)
}

// check a pod to see if it represents a mesh workload
// if it does, produce the appropriate MeshWorkload CR instance corresponding to it
type MeshWorkloadScanner interface {
	ScanPod(context.Context, *corev1.Pod) (*discoveryv1alpha1.MeshWorkload, error)
}
