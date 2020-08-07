package flags

import (
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Options struct {
	Kubeconfig    string
	Kubecontext   string
	MeshName      string
	MeshNamespace string
}

func (o *Options) AddToFlags(set *pflag.FlagSet) {
	set.StringVar(&o.Kubeconfig, "kubeconfig", "", "path to the kubeconfig from which the management cluster will be accessed")
	set.StringVar(&o.Kubecontext, "kubecontext", "", "name of the kubeconfig context to use for the management cluster")
	set.StringVar(&o.MeshName, "mesh-name", "", "name of the Mesh object representing the service mesh being operated on")
	set.StringVar(&o.MeshNamespace, "mesh-namespace", defaults.DefaultPodNamespace, "namespace of the Mesh object representing the service mesh being operated on")

	cobra.MarkFlagRequired(set, "mesh-name")
}
