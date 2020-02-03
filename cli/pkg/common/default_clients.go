package common

import (
	k8sapiv1 "k8s.io/api/core/v1"
	k8sclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
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
