package utils

import (
	"github.com/solo-io/skv2/pkg/multicluster/register"
	"github.com/spf13/pflag"
)

func AddManagementKubeconfigFlags(kubeconfig, kubecontext *string, flags *pflag.FlagSet) {
	flags.StringVar(kubeconfig, "kubeconfig", "", "path to the kubeconfig from which the management cluster will be accessed")
	flags.StringVar(kubecontext, "kubecontext", "", "name of the kubeconfig context to use for the management cluster")
}

type MgmtRemoteKubeConfigOptions struct {
	kubeConfigPath string
	mgmtContext    string
	remoteContext  string
}

func (m *MgmtRemoteKubeConfigOptions) ConstructDiskKubeCfg() (mgmtKubeCfg *register.KubeCfg, remoteKubeCfg *register.KubeCfg) {
	mgmtKubeCfg = register.NewDiskKubeCfg(m.kubeConfigPath, m.mgmtContext)
	remoteKubeCfg = register.NewDiskKubeCfg(m.kubeConfigPath, m.remoteContext)
	return mgmtKubeCfg, remoteKubeCfg
}

func AddMgmtRemoteKubeConfigFlags(kubeConfigOptions *MgmtRemoteKubeConfigOptions, flags *pflag.FlagSet) {
	flags.StringVar(&kubeConfigOptions.kubeConfigPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	flags.StringVar(&kubeConfigOptions.mgmtContext, "mgmt-context", "", "name of the kubeconfig context to use for the management cluster")
	flags.StringVar(&kubeConfigOptions.remoteContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
}
