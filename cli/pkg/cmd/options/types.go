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

// TODO(mitchdraft) Rename this NewSecret (to disambigute from secret ResourceRef)
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
	Secret ResourceRef
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

// All the cli-relevant resources keyed by namespace
type NsResourceMap map[string]*NsResource

// NsResource contains lists of the resources needed by the cli associated* with given namespace.
// *the association is by the namespace in which the CRD is installed, unless otherwise noted.
type NsResource struct {
	// keyed by namespace containing the CRD
	Meshes  []string
	Secrets []string

	// keyed by mesh installation namespace
	// purpose of this list: allows user to select a mesh by the namespace in which they installed the mesh
	// needs to be a resource ref so we can point back to the resource
	MeshesByInstallNs []ResourceRef
}
