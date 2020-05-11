package kubeconfig

import (
	"context"

	k8s_core_v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

const KubeConfigSecretLabel = "solo.io/kubeconfig"

type KubeConfig struct {
	// the actual kubeconfig
	Config clientcmdapi.Config
	// expected to be used as an identifier string for a cluster
	// stored as the key for the kubeconfig data in a kubernetes secret
	Cluster string
}

// package up all forms of the config in one convenience struct
type ConvertedConfigs struct {
	ClientConfig clientcmd.ClientConfig
	ApiConfig    *clientcmdapi.Config
	RestConfig   *rest.Config
}

type Converter interface {
	// convert a kube config to the properly-formatted secret
	// If the kube config contains a reference to any files, those are read and the bytes moved to the in-memory secret
	ConfigToSecret(secretName string, secretNamespace string, config *KubeConfig) (*k8s_core_v1.Secret, error)
	// parse a secret out into the cluster it corresponds to as well as all the kube config formats you may need
	SecretToConfig(secret *k8s_core_v1.Secret) (clusterName string, config *ConvertedConfigs, err error)
}

// given a path to a kube config file, convert it into either creds for hitting the API server of the cluster it points to,
// or return the contexts/clusters it is aware of
type KubeLoader interface {
	GetConfigWithContext(masterURL, kubeconfigPath, context string) (clientcmd.ClientConfig, error)
	GetRestConfigForContext(path string, context string) (*rest.Config, error)
	GetRawConfigForContext(path, context string) (clientcmdapi.Config, error)
	RESTClientGetter(path, context string) resource.RESTClientGetter
	GetRestConfigFromBytes(config []byte) (*rest.Config, error)
}

type KubeConfigLookup interface {
	// get various pieces of config corresponding to a registered kube cluster
	FromCluster(ctx context.Context, clusterName string) (config *ConvertedConfigs, err error)
}
