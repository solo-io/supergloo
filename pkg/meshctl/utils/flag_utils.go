package utils

import (
	"time"

	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

// Set kubeconfig path and context flags for the management cluster.
func AddManagementKubeconfigFlags(kubeconfig, kubecontext *string, flags *pflag.FlagSet) {
	flags.StringVar(kubeconfig, "kubeconfig", "", "path to the kubeconfig from which the management cluster will be accessed")
	flags.StringVar(kubecontext, "kubecontext", "", "name of the kubeconfig context to use for the management cluster")
}

// Kubeconfig options for both management and remote clusters.
type MgmtRemoteKubeConfigOptions struct {
	kubeConfigPath string
	mgmtContext    string
	// We need to explicitly pass the remote context because of this open issue: https://github.com/kubernetes/client-go/issues/735
	RemoteContext string
}

// Initialize a ClientConfig for the management and remote clusters from the options.
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

// Set kubeconfig path and context flags for both the management cluster and remote cluster.
func (m *MgmtRemoteKubeConfigOptions) AddMgmtRemoteKubeConfigFlags(flags *pflag.FlagSet) {
	flags.StringVar(&m.kubeConfigPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	flags.StringVar(&m.mgmtContext, "mgmt-context", "", "name of the kubeconfig context to use for the management cluster")
	flags.StringVar(&m.RemoteContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
}
