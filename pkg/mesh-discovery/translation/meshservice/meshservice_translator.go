package meshservice

import (
	"context"

	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/go-utils/contextutils"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/meshservice/detector"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./meshservice_translator.go -destination mocks/meshservice_translator.go

// the mesh-service translator converts deployments with injected sidecars into MeshService CRs
type Translator interface {
	TranslateMeshServices(
		services corev1sets.ServiceSet,
		meshWorkloads v1alpha2sets.MeshWorkloadSet,
		meshes v1alpha2sets.MeshSet,
	) v1alpha2sets.MeshServiceSet
}

type translator struct {
	ctx                 context.Context
	meshServiceDetector detector.MeshServiceDetector
}

func NewTranslator(ctx context.Context, meshServiceDetector detector.MeshServiceDetector) Translator {
	return &translator{ctx: ctx, meshServiceDetector: meshServiceDetector}
}

func (t *translator) TranslateMeshServices(
	services corev1sets.ServiceSet,
	meshWorkloads v1alpha2sets.MeshWorkloadSet,
	meshes v1alpha2sets.MeshSet,
) v1alpha2sets.MeshServiceSet {

	meshServiceSet := v1alpha2sets.NewMeshServiceSet()

	for _, service := range services.List() {
		meshService := t.meshServiceDetector.DetectMeshService(service, meshWorkloads, meshes)
		if meshService == nil {
			continue
		}
		contextutils.LoggerFrom(t.ctx).Debugf("detected mesh service %v", sets.Key(meshService))
		meshServiceSet.Insert(meshService)
	}
	return meshServiceSet
}
