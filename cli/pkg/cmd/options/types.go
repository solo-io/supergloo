package options

import (
	core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	Top         Top
	Install     Install
	MeshTool    MeshTool
	IngressTool IngressTool
	Get         Get
	Create      Create
	Config      Config
	Cache       OptionsCache
}

type Top struct {
	Static bool
}

type Install struct {
	Filename            string
	MeshType            string
	Namespace           string
	Mtls                bool
	SecretRef           core.ResourceRef
	WatchNamespaces     []string
	ConsulServerAddress string
}

type MeshTool struct {
	MeshId    string
	ServiceId string
}

type IngressTool struct {
	IngressId string
	RouteId   string
}

type Get struct {
	Output string
}

type RoutingRule struct {
	Mesh             string
	Namespace        string
	Sources          string
	Destinations     string
	Matchers         []string
	OverrideExisting bool
}

// // Route Rule fields
// Status
// Metadata
// TargetMesh
// Sources
// Destinations
// RequestMatchers
// TrafficShifting
// FaultInjection
// Timeout
// Retries
// CorsPolicy
// Mirror
// HeaderManipulaition

type Secret struct {
	RootCa     string
	PrivateKey string
	CertChain  string
	Namespace  string
	Name       string
}

type Create struct {
	RoutingRule RoutingRule
	Secret      Secret
}

type Config struct {
	Ca ConfigCa
}

type ConfigCa struct {
	Mesh   ResourceRef
	Secret Secret
}

type ResourceRef struct {
	Name      string
	Namespace string
}

// OptionsCache holds resources that multiple commands need
// It should be initialized on start
type OptionsCache struct {
	Namespaces  []string
	KubeClient  *kubernetes.Clientset
	NsResources NsResourceMap
}

type NsResourceMap map[string]*NsResource

type NsResource struct {
	Meshes  []string
	Secrets []string
}
