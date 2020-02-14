package mesh

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_apps_v1 "k8s.io/api/apps/v1"
)

// check a deployment to see if it represents a mesh installation
// if it does, produce the appropriate Mesh CR instance corresponding to it
//go:generate mockgen -source ./mesh_scanner.go -destination mocks/mock_mesh_scanner.go -package mock_mesh
type MeshScanner interface {
	ScanDeployment(context.Context, *k8s_apps_v1.Deployment) (*v1alpha1.Mesh, error)
}
