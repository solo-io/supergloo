package mesh

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"

	"github.com/solo-io/go-utils/contextutils"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./mesh_translator.go -destination mocks/mesh_translator.go

// the mesh translator converts deployments with control plane images into Mesh CRs
type Translator interface {
	TranslateMeshes(in input.Snapshot) v1alpha2sets.MeshSet
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

func (t *translator) TranslateMeshes(in input.Snapshot) v1alpha2sets.MeshSet {
	meshSet := v1alpha2sets.NewMeshSet()
	meshes, err := t.meshDetector.DetectMeshes(in)
	if err != nil {
		contextutils.LoggerFrom(t.ctx).Warnw("ecnountered error discovering meshes", "err", err)
	}
	for _, mesh := range meshes {
		contextutils.LoggerFrom(t.ctx).Debugf("detected mesh %v", sets.Key(mesh))
		meshSet.Insert(mesh)
	}
	return meshSet
}
