package flags

import (
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/utils"
	"github.com/spf13/pflag"
)

type Options struct {
	Kubeconfig  string
	Kubecontext string
	Namespace   string
}

func (o *Options) AddToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.Kubeconfig, &o.Kubecontext, flags)
	flags.StringVar(&o.Namespace, "namespace", defaults.DefaultPodNamespace, "namespace that Service MeshService Hub is installed in")
}
