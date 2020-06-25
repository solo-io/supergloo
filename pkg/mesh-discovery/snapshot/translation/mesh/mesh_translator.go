package mesh

import (
	"context"
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	"github.com/solo-io/go-utils/contextutils"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/mesh/detector"
)

// the mesh translator converts deployments with control plane images into Mesh CRs
type Translator interface {
	TranslateMeshes(deployments appsv1sets.DeploymentSet) v1alpha1sets.MeshSet
}

type translator struct {
	ctx          context.Context
	meshDetector detector.MeshDetector
}

func NewTranslator(
	ctx context.Context,
	meshDetector detector.MeshDetector,
) Translator {
	ctx = contextutils.WithLogger(ctx, "mesh-translator")
	return &translator{ctx: ctx, meshDetector: meshDetector}
}

func (t *translator) TranslateMeshes(deployments appsv1sets.DeploymentSet) v1alpha1sets.MeshSet {
	meshSet := v1alpha1sets.NewMeshSet()
	for _, deployment := range deployments.List() {
		mesh, err := t.meshDetector.DetectMesh(deployment)
		if err != nil {
			contextutils.LoggerFrom(t.ctx).Warnw("failed to discover mesh for deployment ", "deployment", sets.Key(deployment))
		}
		if mesh == nil {
			continue
		}
		meshSet.Insert(mesh)
	}
	return meshSet
}
