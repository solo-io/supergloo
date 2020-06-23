package meshservice

import (
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/meshservice/detector"
)

// the mesh-service translator converts deployments with injected sidecars into MeshService CRs
type Translator interface {
	TranslateMeshServices(services corev1sets.ServiceSet) v1alpha1sets.MeshServiceSet
}

type translator struct {
	meshServiceDetector detector.MeshServiceDetector
}

func NewTranslator() Translator {
	return &translator{}
}

func (t *translator) TranslateMeshServices(services corev1sets.ServiceSet) v1alpha1sets.MeshServiceSet {

	meshServiceSet := v1alpha1sets.NewMeshServiceSet()

	for _, service := range services.List() {
		meshService := t.meshServiceDetector.DetectMeshService(service)
		if meshService == nil {
			continue
		}
		meshServiceSet.Insert(meshService)
	}
	return meshServiceSet
}
