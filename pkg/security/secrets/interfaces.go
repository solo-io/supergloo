package cert_secrets

import (
	corev1 "k8s.io/api/core/v1"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

/*
	CertSecretBuilder is meant as an interface function to expose the ability to transform itself into a secret
	representing the underlying data

	For instance: the RootCaData will build a new secret containing all of the it's certificate related information
	in the name and namespace provided
*/
type CertSecretBuilder interface {
	BuildSecret(name, namespace string) *corev1.Secret
}
