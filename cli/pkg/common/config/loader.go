package common_config

import (
	"os"

	"github.com/solo-io/go-utils/kubeutils"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// given a path to a kube config file, convert it into either creds for hitting the API server of the cluster it points to,
// or return the contexts/clusters it is aware of
//go:generate mockgen -destination ../../mocks/mock_kube_loader.go -package cli_mocks github.com/solo-io/mesh-projects/cli/pkg/common/config KubeLoader
type KubeLoader interface {
	GetRestConfigForContext(path string, context string) (*rest.Config, error)
	GetRawConfigForContext(path, context string) (clientcmdapi.Config, error)
}

// only the pieces from a kube config that we need to operate on
// mainly just used to simplify from the complexity of the actual object
type KubeContext struct {
	CurrentContext string
	Contexts       map[string]*api.Context
	Clusters       map[string]*api.Cluster
}

// default KubeLoader
func DefaultKubeLoaderProvider() KubeLoader {
	return &kubeLoader{}
}

type kubeLoader struct{}

func (k *kubeLoader) GetRestConfigForContext(path string, context string) (*rest.Config, error) {
	return getConfigWithContext("", path, context).ClientConfig()
}

func getConfigWithContext(masterURL, kubeconfigPath, context string) clientcmd.ClientConfig {
	if kubeconfigPath != "" {
		if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
			// the specified kubeconfig does not exist so fallback to default
			kubeconfigPath = ""
		}
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfigPath
	configOverrides := &clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterURL}}
	if context != "" {
		configOverrides.CurrentContext = context
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
}

func (k *kubeLoader) GetRawConfigForContext(path, context string) (clientcmdapi.Config, error) {
	return getConfigWithContext("", path, context).RawConfig()
}

func (k *kubeLoader) ParseContext(path string) (*KubeContext, error) {
	cfg, err := kubeutils.GetKubeConfig("", path)
	if err != nil {
		return nil, err
	}

	return &KubeContext{
		CurrentContext: cfg.CurrentContext,
		Contexts:       cfg.Contexts,
		Clusters:       cfg.Clusters,
	}, nil
}
