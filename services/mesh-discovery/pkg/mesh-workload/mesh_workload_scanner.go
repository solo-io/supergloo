package mesh_workload

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	core_v1 "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./mesh_workload_scanner.go -destination mocks/mock_mesh_workload_scanner.go

// check a pod to see if it represents a mesh workload
// if it does, produce the appropriate MeshWorkload CR instance corresponding to it
type MeshWorkloadScanner interface {
	ScanPod(context.Context, *core_v1.Pod) (*v1alpha1.MeshWorkload, error)
}
