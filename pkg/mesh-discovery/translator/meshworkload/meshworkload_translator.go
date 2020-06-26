package meshworkload

import (
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/smh/pkg/mesh-discovery/translator/meshworkload/detector"
	"github.com/solo-io/smh/pkg/mesh-discovery/translator/meshworkload/types"
)

//go:generate mockgen -source ./meshworkload_translator.go -destination mocks/meshworkload_translator.go

// the mesh-workload translator converts deployments with injected sidecars into MeshWorkload CRs
type Translator interface {
	TranslateMeshWorkloads(deployments appsv1sets.DeploymentSet, daemonSets appsv1sets.DaemonSetSet, statefulSets appsv1sets.StatefulSetSet, meshes v1alpha1sets.MeshSet) v1alpha1sets.MeshWorkloadSet
}

type translator struct {
	meshWorkloadDetector detector.MeshWorkloadDetector
}

func NewTranslator(meshWorkloadDetector detector.MeshWorkloadDetector) Translator {
	return &translator{meshWorkloadDetector: meshWorkloadDetector}
}

func (t *translator) TranslateMeshWorkloads(deployments appsv1sets.DeploymentSet, daemonSets appsv1sets.DaemonSetSet, statefulSets appsv1sets.StatefulSetSet, meshes v1alpha1sets.MeshSet) v1alpha1sets.MeshWorkloadSet {
	var workloads []types.Workload
	for _, deployment := range deployments.List() {
		workloads = append(workloads, types.ToWorkload(deployment))
	}
	for _, daemonSet := range daemonSets.List() {
		workloads = append(workloads, types.ToWorkload(daemonSet))
	}
	for _, statefulSet := range statefulSets.List() {
		workloads = append(workloads, types.ToWorkload(statefulSet))
	}

	meshWorkloadSet := v1alpha1sets.NewMeshWorkloadSet()

	for _, workload := range workloads {
		meshWorkload := t.meshWorkloadDetector.DetectMeshWorkload(workload, meshes)
		if meshWorkload == nil {
			continue
		}
		meshWorkloadSet.Insert(meshWorkload)
	}
	return meshWorkloadSet
}
