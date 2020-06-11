package kubeconfig

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/common/filesystem/files"
	k8s_core_v1 "k8s.io/api/core/v1"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	FailedToReadCAFile = func(err error, fileName string) error {
		return eris.Wrapf(err, "Failed to read kubeconfig CA file: %s", fileName)
	}
	DuplicateClusterName = func(repeatedClusterName string) error {
		return eris.Errorf("Error converting KubeConfigs to secret, duplicate cluster name found: %s", repeatedClusterName)
	}
	FailedToConvertSecretToKubeConfig = func(err error) error {
		return eris.Wrapf(err, "Could not deserialize string to KubeConfig while generating KubeConfig")
	}
	NoDataInKubeConfigSecret = func(secret *k8s_core_v1.Secret) error {
		return eris.Errorf("No data in kube config secret %s.%s", secret.ObjectMeta.Name, secret.ObjectMeta.Namespace)
	}
	FailedToConvertSecretToClientConfig = func(err error) error {
		return eris.Wrap(err, "Could not convert config to ClientConfig")
	}
)

func NewConverter(fileReader files.FileReader) Converter {
	return &converter{
		fileReader: fileReader,
	}
}

type converter struct {
	fileReader files.FileReader
}

func (c *converter) ConfigToSecret(secretName string, secretNamespace string, config *KubeConfig) (*k8s_core_v1.Secret, error) {
	// shuffle over the CA bytes from a file on-disk to in-memory bytes if necessary
	err := c.readCertAuthFileIfNecessary(config.Config)
	if err != nil {
		return nil, err
	}

	rawKubeConfig, err := clientcmd.Write(config.Config)
	if err != nil {
		return nil, eris.Wrap(err, "Could not serialize KubeConfig to yaml while generating secret.")
	}

	secretData := map[string][]byte{
		config.Cluster: rawKubeConfig,
	}

	return &k8s_core_v1.Secret{
		ObjectMeta: k8s_meta_v1.ObjectMeta{
			Labels:    map[string]string{KubeConfigSecretLabel: "true"},
			Name:      secretName,
			Namespace: secretNamespace,
		},
		Type: k8s_core_v1.SecretTypeOpaque,
		Data: secretData,
	}, nil
}

// https://github.com/solo-io/service-mesh-hub/issues/590
// If the user has a cert authority file set instead of the raw bytes in their kubeconfig, then
// we'll fail later when the pods in-cluster try to read that file path.
// We need to read the file right now, in a CLI context, and manually shuffle the bytes over to the CA data field
func (c *converter) readCertAuthFileIfNecessary(cfg api.Config) error {
	currentCluster := cfg.Clusters[cfg.Contexts[cfg.CurrentContext].Cluster]
	if len(currentCluster.CertificateAuthority) > 0 {
		fileContent, err := c.fileReader.Read(currentCluster.CertificateAuthority)
		if err != nil {
			return FailedToReadCAFile(err, currentCluster.CertificateAuthority)
		}

		currentCluster.CertificateAuthorityData = fileContent
		currentCluster.CertificateAuthority = "" // dont need to record the filename in the config; we have the data present
	}

	return nil
}

func (c *converter) SecretToConfig(secret *k8s_core_v1.Secret) (clusterName string, config *ConvertedConfigs, err error) {
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
		return clusterName, &ConvertedConfigs{
			ClientConfig: clientConfig,
			RestConfig:   restConfig,
			ApiConfig:    apiConfig,
		}, nil
	}

	return "", nil, NoDataInKubeConfigSecret(secret)
}
