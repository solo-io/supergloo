package mesh

import (
	"context"
	"github.com/hashicorp/go-multierror"
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/mesh/detector"
)

// the mesh translator translates Mesh CRs from
// a list of deployments.
// the deployments must have their ClusterName set
type Translator interface {
	TranslateMeshes(deployments appsv1sets.DeploymentSet) (v1alpha1sets.MeshSet, error)
}

type translator struct {
	ctx          context.Context
	meshDetector detector.MeshDetector
}

func NewTranslator(
	ctx context.Context,
) Translator {
	return &translator{ctx: ctx}
}

func (t *translator) TranslateMeshes(deployments appsv1sets.DeploymentSet) (v1alpha1sets.MeshSet, error) {
	meshSet := v1alpha1sets.NewMeshSet()
	var errs error
	for _, deployment := range deployments.List() {
		mesh, err := t.meshDetector.DetectMesh(deployment)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		if mesh == nil {
			continue
		}
		meshSet.Insert(mesh)
	}
	return meshSet, nil
}
