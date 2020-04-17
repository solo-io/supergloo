package mesh

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1/controller"
	k8s_apps_v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_mesh_interfaces.go -package mock_mesh

// once StartDiscovery is invoked, MeshFinder's DeploymentEventHandler callbacks will start receiving DeploymentEvents
type MeshFinder interface {
	// an event is only received by our callbacks if all the given predicates return true
	StartDiscovery(deploymentEventWatcher k8s_controller.DeploymentEventWatcher, predicates []predicate.Predicate) error

	k8s_controller.DeploymentEventHandler
}

// check a deployment to see if it represents a mesh installation
// if it does, produce the appropriate Mesh CR instance corresponding to it
type MeshScanner interface {
	ScanDeployment(context.Context, *k8s_apps_v1.Deployment, client.Client) (*zephyr_discovery.Mesh, error)
}
