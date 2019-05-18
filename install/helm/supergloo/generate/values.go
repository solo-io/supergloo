package generate

import "github.com/solo-io/gloo/install/helm/gloo/generate"

type Config struct {
	Supergloo     *Supergloo     `json:"supergloo,omitempty"`
	Discovery     *Discovery     `json:"discovery,omitempty"`
	MeshDiscovery *MeshDiscovery `json:"meshDiscovery,omitempty"`
	Rbac          *Rbac          `json:"rbac,omitempty"`
}

// Common

type Rbac struct {
	Create bool `json:"create"`
}

// Supergloo

type Supergloo struct {
	Deployment               *StandardDeployment `json:"deployment,omitempty"`
	DisablePrometheusBouncer bool                `json:"disablePrometheusBouncer,omitempty"`
}

type Discovery struct {
	Deployment *StandardDeployment `json:"deployment,omitempty"`
}

type MeshDiscovery struct {
	Deployment *StandardDeployment `json:"deployment,omitempty"`
}

type StandardDeployment struct {
	Image *generate.Image `json:"image,omitempty"`
	Stats bool            `json:"stats"`
}
