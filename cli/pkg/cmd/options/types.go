package options

import (
	core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	superglooV1 "github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	Top         Top
	Install     Install
	Uninstall   Uninstall
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

	// Interactive only (not passable via flags)
	UseCustomSecret bool
}

type Uninstall struct {
	All       bool
	MeshNames string
	MeshType  string
}

type MeshTool struct {
	Mesh        core.ResourceRef
	ServiceId   string
	AddPolicy   AddPolicy
	RoutingRule superglooV1.RoutingRule
}

type AddPolicy struct {
	// (Do we care to support bulk entry in a form like this?)
	// PolicyCsv is a comma-separated-list in the form:
	// source_namespace,source_name,destination_namespace,destination_name (repeated)
	// PolicyCsv string

	Source      core.ResourceRef
	Destination core.ResourceRef
}

type IngressTool struct {
	IngressId string
	RouteId   string
}

type Get struct {
	Output string
}

type InputDuration struct {
	Seconds string
	Nanos   string
}

type InputRetry struct {
	Attempts      string
	PerTryTimeout InputDuration
}

// TODO(mitchdraft) Rename this NewSecret (to disambigute from secret ResourceRef)
type Secret struct {
	RootCa     string
	PrivateKey string
	CertChain  string
	Namespace  string
	Name       string
}

type Create struct {
	InputRoutingRule InputRoutingRule
	Secret           Secret
}

type Config struct {
	Ca ConfigCa
}

type ConfigCa struct {
	Mesh   core.ResourceRef
	Secret core.ResourceRef
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
	Meshes    []string
	Secrets   []string
	Upstreams []string

	// keyed by mesh installation namespace
	// purpose of this list: allows user to select a mesh by the namespace in which they installed the mesh
	// needs to be a resource ref so we can point back to the resource
	MeshesByInstallNs []core.ResourceRef
}
