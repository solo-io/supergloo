package mesh

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/services/common/cluster/apps/v1/controller"
	k8s_apps_v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_mesh_interfaces.go -package mock_mesh

// once StartDiscovery is invoked, MeshFinder's DeploymentEventHandler callbacks will start receiving DeploymentEvents
type MeshFinder interface {
	// an event is only received by our callbacks if all the given predicates return true
	StartDiscovery(deploymentController controller.DeploymentController, predicates []predicate.Predicate) error

	controller.DeploymentEventHandler
}

// check a deployment to see if it represents a mesh installation
// if it does, produce the appropriate Mesh CR instance corresponding to it
type MeshScanner interface {
	ScanDeployment(context.Context, *k8s_apps_v1.Deployment) (*v1alpha1.Mesh, error)
}
