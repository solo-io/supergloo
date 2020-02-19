package common_config

import (
	"os"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"k8s.io/cli-runtime/pkg/resource"
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
	RESTClientGetter(path, context string) resource.RESTClientGetter
}

// only the pieces from a kube config that we need to operate on
// mainly just used to simplify from the complexity of the actual object
type KubeContext struct {
	CurrentContext string
	Contexts       map[string]*api.Context
	Clusters       map[string]*api.Cluster
}

// default KubeLoader
func DefaultKubeLoaderProvider(opts *options.Options) KubeLoader {
	return &kubeLoader{
		timeout: opts.Root.KubeTimeout.String(),
	}
}

type kubeLoader struct {
	timeout string
}

func (k *kubeLoader) GetRestConfigForContext(path string, context string) (*rest.Config, error) {
	cfg, err := getConfigWithContext("", path, context)
	if err != nil {
		return nil, err
	}

	return cfg.ClientConfig()
}

func getConfigWithContext(masterURL, kubeconfigPath, context string) (clientcmd.ClientConfig, error) {
	verifiedKubeConfigPath := clientcmd.RecommendedHomeFile
	if kubeconfigPath != "" {
		verifiedKubeConfigPath = kubeconfigPath
	}

	if err := assertKubeConfigExists(verifiedKubeConfigPath); err != nil {
		return nil, err
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = verifiedKubeConfigPath
	configOverrides := &clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterURL}}

	if context != "" {
		configOverrides.CurrentContext = context
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides), nil
}

// expects `path` to be nonempty
func assertKubeConfigExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}

	return nil
}

func (k *kubeLoader) GetRawConfigForContext(path, context string) (clientcmdapi.Config, error) {
	cfg, err := getConfigWithContext("", path, context)
	if err != nil {
		return clientcmdapi.Config{}, err
	}

	return cfg.RawConfig()
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

func (k *kubeLoader) RESTClientGetter(path, context string) resource.RESTClientGetter {
	return NewRESTClientGetter(k, path, context)
}
