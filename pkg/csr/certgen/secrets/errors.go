package cert_secrets

import (
	"github.com/rotisserie/eris"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	NoRootCertFoundError = func(meta metav1.ObjectMeta) error {
		return eris.Errorf("No root cert found for %s.%s", meta.Name, meta.Namespace)
	}
	NoPrivateKeyFoundError = func(meta metav1.ObjectMeta) error {
		return eris.Errorf("No private key found for %s.%s", meta.Name, meta.Namespace)
	}
	NoCertChainFoundError = func(meta metav1.ObjectMeta) error {
		return eris.Errorf("No cert chain found for %s.%s", meta.Name, meta.Namespace)
	}
	NoCaKeyFoundError = func(meta metav1.ObjectMeta) error {
		return eris.Errorf("No ca key found for %s.%s", meta.Name, meta.Namespace)
	}
	NoCaCertFoundError = func(meta metav1.ObjectMeta) error {
		return eris.Errorf("No ca cert found for %s.%s", meta.Name, meta.Namespace)
	}
)
