package util

import (
	"os"

	"github.com/solo-io/solo-kit/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Returns the namespace supergloo is installed to.
func GetSuperglooNamespace(client kubernetes.Interface) (string, error) {

	// First check if POD_NAMESPACE env is set
	if nsFromEnv := os.Getenv("POD_NAMESPACE"); nsFromEnv != "" {
		return nsFromEnv, nil
	}

	// Try to retrieve the supergloo deployment
	if client != nil {
		supergloo, err := client.AppsV1().Deployments(metav1.NamespaceAll).Get("supergloo", metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrapf(err, "could not get supergloo deployment")
		}
		return supergloo.Namespace, nil
	}

	return "", errors.Errorf("could not determine supergloo namespace")
}
