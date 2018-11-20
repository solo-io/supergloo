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
	Cache       OptionsCache
}

type Top struct {
	Static bool
}

type Install struct {
	Filename  string
	MeshType  string
	Namespace string
	Mtls      bool
	SecretRef core.ResourceRef
	Consul    ConsulArgs
}

type ConsulArgs struct {
	Namespace string
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
	Matchers         string
	OverrideExisting bool
}

type Create struct {
	RoutingRule RoutingRule
}

// OptionsCache holds resources that multiple commands need
// It should be initialized on start
type OptionsCache struct {
	Namespaces []string
	KubeClient *kubernetes.Clientset
}
