package mesh

import (
	"context"
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
)

// the mesh translator translates Mesh CRs from
// a list of deployments.
// the deployments must have their ClusterName set
type Translator interface {
	TranslateMeshes(deployments appsv1sets.DeploymentSet) v1alpha1sets.MeshSet
}

type translator struct {
	ctx context.Context
	scanners scanners.DeploymentScanner
}

func NewTranslator(
	ctx context.Context,
	) Translator {
	return &translator{ctx: ctx}
}

func (t *translator) TranslateMeshes(deployments appsv1sets.DeploymentSet) v1alpha1sets.MeshSet {

}


func (t *translator) translateMesh(deployments appsv1sets.DeploymentSet) (*v1alpha1.Mesh, error) {

}

