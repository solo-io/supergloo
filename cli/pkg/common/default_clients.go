package common

import (
	"github.com/solo-io/go-utils/kubeerrutils"
	k8sapiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	k8sclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// default implementation of a SecretWriter
func DefaultSecretWriterProvider(client kubernetes.Interface, writeNamespace string) SecretWriter {
	return &secretWriter{client.CoreV1().Secrets(writeNamespace)}
}

type secretWriter struct {
	secretInterface k8sclientv1.SecretInterface
}

func (s *secretWriter) Apply(secret *k8sapiv1.Secret) error {
	_, err := s.secretInterface.Create(secret)
	if err != nil && kubeerrutils.IsAlreadyExists(err) {
		_, err = s.secretInterface.Update(secret)
	}
	return err
}
