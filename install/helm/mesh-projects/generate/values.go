package generate

import (
	"github.com/solo-io/gloo/install/helm/gloo/generate"
)

type HelmConfig struct {
	Config
	Global *Global `json:"global,omitempty"`
}

type Config struct {
	MeshDiscovery *MeshDiscovery `json:"meshDiscovery,omitempty"`
	MeshBridge    *MeshBridge    `json:"meshBridge,omitempty"`
}

type Global struct {
	Image *generate.Image `json:"image,omitempty"`
	Rbac  *Rbac           `json:"rbac,omitempty"`
	Crds  *Crds           `json:"crds,omitempty"`
}

type Namespace struct {
	Create bool `json:"create" desc:"create the installation namespace"`
}

type Rbac struct {
	Create     bool `json:"create" desc:"create rbac rules for the service-mesh-hub service account"`
	Namespaced bool `json:"Namespaced" desc:"use Roles instead of ClusterRoles"`
}

type Crds struct {
	Create bool `json:"create" desc:"create CRDs for MeshDiscovery (turn off if installing with Helm to a cluster that already has MeshDiscovery CRDs)"`
}

type MeshDiscovery struct {
	Disabled   bool                     `json:"disabled"`
	Deployment *MeshDiscoveryDeployment `json:"deployment,omitempty"`
}

type MeshDiscoveryDeployment struct {
	Image *generate.Image `json:"image,omitempty"`
	Stats bool            `json:"stats" desc:"enable prometheus stats"`
	*generate.DeploymentSpec
}

type MeshBridge struct {
	Disabled   bool                  `json:"disabled"`
	Deployment *MeshBridgeDeployment `json:"deployment,omitempty"`
}

type MeshBridgeDeployment struct {
	Image *generate.Image `json:"image,omitempty"`
	Stats bool            `json:"stats" desc:"enable prometheus stats"`
	*generate.DeploymentSpec
}
