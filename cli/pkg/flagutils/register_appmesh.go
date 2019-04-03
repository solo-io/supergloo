package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func RegisterAwsAppMeshFlags(set *pflag.FlagSet, register *options.RegisterAppMesh) {
	set.Var(&register.Secret, "secret", "secret holding AWS access credentials. Format must be NAMESPACE.NAME")
	set.StringVar(&register.Region, "region", "", "AWS region the AWS App Mesh control plane resources (Virtual Nodes, Virtual Routers, etc.) will be created in")
	set.StringVar(&register.EnableAutoInjection, "auto-inject", "true", "determines whether auto-injection will be enabled for this mesh")
	set.Var(&register.ConfigMap, "configmap", "config map that contains the patch to be applied to the pods matching the selector. Format must be NAMESPACE.NAME")
	set.Var(&register.PodSelector.SelectedLabels, "select-labels", "auto-inject pods with these labels. Format must be KEY=VALUE")
	set.StringSliceVar(&register.PodSelector.SelectedNamespaces, "select-namespaces", nil, "auto-inject pods matching these labels")
	set.StringVar(&register.VirtualNodeLabel, "virtual-node-label", "", "If auto-injection is enabled, "+
		"the value of the pod label with this key will be used to calculate the value of APPMESH_VIRTUAL_NODE_NAME environment variable that is set on the injected sidecar proxy container.")
}
