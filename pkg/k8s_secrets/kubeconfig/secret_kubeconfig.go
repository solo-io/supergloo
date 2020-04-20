package kubeconfig

import (
	"github.com/rotisserie/eris"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const KubeConfigSecretLabel = "solo.io/kubeconfig"

type KubeConfig struct {
	// the actual kubeconfig
	Config api.Config
	// expected to be used as an identifier string for a cluster
	// stored as the key for the kubeconfig data in a kubernetes secret
	Cluster string
}

var (
	DuplicateClusterName = func(repeatedClusterName string) error {
		return eris.Errorf("Error converting KubeConfigs to secret, duplicate cluster name found: %s", repeatedClusterName)
	}
	FailedToConvertSecretToKubeConfig = func(err error) error {
		return eris.Wrapf(err, "Could not deserialize string to KubeConfig while generating KubeConfig")
	}
	NoDataInKubeConfigSecret = func(secret *k8s_core_types.Secret) error {
		return eris.Errorf("No data in kube config secret %s.%s", secret.ObjectMeta.Name, secret.ObjectMeta.Namespace)
	}
	FailedToConvertSecretToClientConfig = func(err error) error {
		return eris.Wrap(err, "Could not convert config to ClientConfig")
	}
)

func KubeConfigToSecret(name string, namespace string, kc *KubeConfig) (*k8s_core_types.Secret, error) {
	return KubeConfigsToSecret(name, namespace, []*KubeConfig{kc})
}

func KubeConfigsToSecret(name string, namespace string, kcs []*KubeConfig) (*k8s_core_types.Secret, error) {
	secretData := map[string][]byte{}
	for _, kc := range kcs {
		rawKubeConfig, err := clientcmd.Write(kc.Config)
		if err != nil {
			return nil, eris.Wrap(err, "Could not serialize KubeConfig to yaml while generating secret.")
		}
		if _, exists := secretData[kc.Cluster]; exists {
			return nil, DuplicateClusterName(kc.Cluster)
		}
		secretData[kc.Cluster] = rawKubeConfig
	}
	return &k8s_core_types.Secret{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Labels:    map[string]string{KubeConfigSecretLabel: "true"},
			Name:      name,
			Namespace: namespace,
		},
		Type: k8s_core_types.SecretTypeOpaque,
		Data: secretData,
	}, nil
}

type SecretToConfigConverter func(secret *k8s_core_types.Secret) (clusterName string, config *Config, err error)

func SecretToConfigConverterProvider() SecretToConfigConverter {
	return SecretToConfig
}

type Config struct {
	ClientConfig clientcmd.ClientConfig
	ApiConfig    *api.Config
	RestConfig   *rest.Config
}

func SecretToConfig(secret *k8s_core_types.Secret) (clusterName string, config *Config, err error) {
	if len(secret.Data) > 1 {
		return "", nil, eris.Errorf("kube config secret %s.%s has multiple keys in its data, this is unexpected",
			secret.ObjectMeta.Name, secret.ObjectMeta.Namespace)
	}
	for clusterName, dataEntry := range secret.Data {
		clientConfig, err := clientcmd.NewClientConfigFromBytes(dataEntry)
		if err != nil {
			return clusterName, nil, FailedToConvertSecretToClientConfig(err)
		}

		apiConfig, err := clientcmd.Load(dataEntry)
		if err != nil {
			return clusterName, nil, FailedToConvertSecretToKubeConfig(err)
		}

		restConfig, err := clientConfig.ClientConfig()
		if err != nil {
			return clusterName, nil, eris.Wrapf(err, "Failed to convert secret %s.%s to rest config",
				secret.ObjectMeta.Name, secret.ObjectMeta.Namespace)
		}
		return clusterName, &Config{
			ClientConfig: clientConfig,
			RestConfig:   restConfig,
			ApiConfig:    apiConfig,
		}, nil
	}

	return "", nil, NoDataInKubeConfigSecret(secret)
}
