package kubeconfig

import (
	"os"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/kubeutils"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	FailedLoadingKubeConfig = func(err error) error {
		return eris.Wrap(err, "Failed to load the kube config")
	}
)

// only the pieces from a kube config that we need to operate on
// mainly just used to simplify from the complexity of the actual object
type KubeContext struct {
	CurrentContext string
	Contexts       map[string]*api.Context
	Clusters       map[string]*api.Cluster
}

func NewKubeLoader() KubeLoader {
	return &kubeLoader{}
}

type kubeLoader struct{}

func (k *kubeLoader) GetRestConfigForContext(path string, context string) (*rest.Config, error) {
	cfg, err := k.GetConfigWithContext("", path, context)
	if err != nil {
		return nil, FailedLoadingKubeConfig(err)
	}

	return cfg.ClientConfig()
}

func (k *kubeLoader) GetRestConfigFromBytes(config []byte) (*rest.Config, error) {
	return clientcmd.RESTConfigFromKubeConfig(config)
}

func (k *kubeLoader) GetConfigWithContext(masterURL, kubeconfigPath, context string) (clientcmd.ClientConfig, error) {
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
	cfg, err := k.GetConfigWithContext("", path, context)
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
	return NewFileBackedRESTClientGetter(k, path, context)
}
