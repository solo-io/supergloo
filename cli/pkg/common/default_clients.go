package common

import (
	"github.com/solo-io/go-utils/kubeutils"
	k8sapiv1 "k8s.io/api/core/v1"
	k8sclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// default implementation of a SecretWriter
func DefaultSecretWriterProvider(client *k8sclientv1.CoreV1Client, writeNamespace string) SecretWriter {
	return &secretWriter{client.Secrets(writeNamespace)}
}

type secretWriter struct {
	secretInterface k8sclientv1.SecretInterface
}

func (s *secretWriter) Write(secret *k8sapiv1.Secret) error {
	_, err := s.secretInterface.Create(secret)
	return err
}

// default KubeLoader
func DefaultKubeLoaderProvider() KubeLoader {
	return &kubeLoader{}
}

type kubeLoader struct{}

func (k *kubeLoader) GetRestConfig(path string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", path)
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
