package decorator

import (
	v1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/workload/types"
)

//go:generate mockgen -source ./workload_decorator.go -destination mocks/workload_director.go

type WorkloadDecorator interface {
	DecorateWorkload(
		discoveredWorkload *v1alpha2.Workload,
		kubeWorkload types.Workload,
		mesh *v1alpha2.Mesh,
		pods v1sets.PodSet,
	)
}

type WorkloadDecorators []WorkloadDecorator

func (w WorkloadDecorators) DecorateWorkload(discoveredWorkload *v1alpha2.Workload, kubeWorkload types.Workload, mesh *v1alpha2.Mesh, pods v1sets.PodSet) {
	for _, workloadDecorator := range w {
		workloadDecorator.DecorateWorkload(discoveredWorkload, kubeWorkload, mesh, pods)
	}
}
