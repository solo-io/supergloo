package mesh

import (
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/meshworkload/detector"
)

// the mesh-workload translator converts deployments with injected sidecars into MeshWorkload CRs
type Translator interface {
	TranslateMeshWorkloades(deployments appsv1sets.DeploymentSet) v1alpha1sets.MeshWorkloadSet
}

type translator struct {
	meshWorkloadDetector detector.MeshWorkloadDetector
}

func NewTranslator() Translator {
	return &translator{}
}

func (t *translator) TranslateMeshWorkloades(deployments appsv1sets.DeploymentSet) v1alpha1sets.MeshWorkloadSet {
	meshWorkloadSet := v1alpha1sets.NewMeshWorkloadSet()
	for _, deployment := range deployments.List() {
		meshWorkload := t.meshWorkloadDetector.DetectMeshWorkload(deployment)
		if meshWorkload == nil {
			continue
		}
		meshWorkloadSet.Insert(meshWorkload)
	}
	return meshWorkloadSet
}
