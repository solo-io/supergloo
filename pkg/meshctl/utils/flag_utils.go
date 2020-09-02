package utils

import (
	"time"

	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

func AddManagementKubeconfigFlags(kubeconfig, kubecontext *string, flags *pflag.FlagSet) {
	flags.StringVar(kubeconfig, "kubeconfig", "", "path to the kubeconfig from which the management cluster will be accessed")
	flags.StringVar(kubecontext, "kubecontext", "", "name of the kubeconfig context to use for the management cluster")
}

type MgmtRemoteKubeConfigOptions struct {
	kubeConfigPath string
	mgmtContext    string
	RemoteContext  string
}

func (m *MgmtRemoteKubeConfigOptions) ConstructClientConfigs() (mgmtKubeCfg clientcmd.ClientConfig, remoteKubeCfg clientcmd.ClientConfig, err error) {
	loader := kubeconfig.NewKubeLoader(5 * time.Second)
	mgmtKubeCfg, err = loader.GetClientConfigForContext(m.kubeConfigPath, m.mgmtContext)
	if err != nil {
		return nil, nil, err
	}
	remoteKubeCfg, err = loader.GetClientConfigForContext(m.kubeConfigPath, m.RemoteContext)
	if err != nil {
		return nil, nil, err
	}
	return mgmtKubeCfg, remoteKubeCfg, nil
}

func AddMgmtRemoteKubeConfigFlags(kubeConfigOptions *MgmtRemoteKubeConfigOptions, flags *pflag.FlagSet) {
	flags.StringVar(&kubeConfigOptions.kubeConfigPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	flags.StringVar(&kubeConfigOptions.mgmtContext, "mgmt-context", "", "name of the kubeconfig context to use for the management cluster")
	flags.StringVar(&kubeConfigOptions.RemoteContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
}
