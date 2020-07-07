package meshservice

import (
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/smh/pkg/mesh-discovery/translation/meshservice/detector"
)

//go:generate mockgen -source ./meshservice_translator.go -destination mocks/meshservice_translator.go

// the mesh-service translator converts deployments with injected sidecars into MeshService CRs
type Translator interface {
	TranslateMeshServices(services corev1sets.ServiceSet, meshWorkloads v1alpha1sets.MeshWorkloadSet) v1alpha1sets.MeshServiceSet
}

type translator struct {
	meshServiceDetector detector.MeshServiceDetector
}

func NewTranslator(meshServiceDetector detector.MeshServiceDetector) Translator {
	return &translator{meshServiceDetector:meshServiceDetector}
}

func (t *translator) TranslateMeshServices(services corev1sets.ServiceSet, meshWorkloads v1alpha1sets.MeshWorkloadSet) v1alpha1sets.MeshServiceSet {

	meshServiceSet := v1alpha1sets.NewMeshServiceSet()

	for _, service := range services.List() {
		meshService := t.meshServiceDetector.DetectMeshService(service, meshWorkloads)
		if meshService == nil {
			continue
		}
		meshServiceSet.Insert(meshService)
	}
	return meshServiceSet
}
