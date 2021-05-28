package context

import (
	"istio.io/istio/pkg/test/framework/components/cluster"
	"istio.io/istio/pkg/test/framework/components/echo"
	"istio.io/istio/pkg/test/framework/components/namespace"
	"istio.io/istio/pkg/test/framework/resource"
)

type DeploymentContext struct {
	EchoContext *EchoDeploymentContext
	Meshes      []GlooMeshInstance
}

type EchoDeploymentContext struct {
	Deployments     echo.Instances
	AppNamespace    namespace.Instance
	SubsetNamespace namespace.Instance
	NoMeshNamespace namespace.Instance
}

// GlooMeshInstance is a component that provides access to a deployed echo service.
type GlooMeshInstance interface {
	resource.Resource
	GetKubeConfig() string
	IsManagementPlane() bool
	// for management plane only
	GetRelayServerAddress() (string, error)
	GetCluster() cluster.Cluster
}

func (d *DeploymentContext) GetManagementPlaneCluster() cluster.Cluster {
	for _, m := range d.Meshes {
		if m.IsManagementPlane() {
			return m.GetCluster()
		}
	}
	return nil
}
