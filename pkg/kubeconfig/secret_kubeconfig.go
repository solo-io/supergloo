package kubeconfig

import (
	"github.com/rotisserie/eris"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
)

func KubeConfigToSecret(name string, namespace string, kc *KubeConfig) (*kubev1.Secret, error) {
	return KubeConfigsToSecret(name, namespace, []*KubeConfig{kc})
}

func KubeConfigsToSecret(name string, namespace string, kcs []*KubeConfig) (*kubev1.Secret, error) {
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
	return &kubev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    map[string]string{KubeConfigSecretLabel: "true"},
			Name:      name,
			Namespace: namespace,
		},
		Type: kubev1.SecretTypeOpaque,
		Data: secretData,
	}, nil
}

func SecretToKubeConfigs(secret *kubev1.Secret) ([]*KubeConfig, error) {
	var kubeConfigs = []*KubeConfig{}
	for cluster, dataEntry := range secret.Data {
		config, err := clientcmd.Load(dataEntry)
		if err != nil {
			return nil, FailedToConvertSecretToKubeConfig(err)
		}
		kubeConfigs = append(kubeConfigs, &KubeConfig{
			Config:  *config,
			Cluster: cluster,
		})
	}
	return kubeConfigs, nil
}
