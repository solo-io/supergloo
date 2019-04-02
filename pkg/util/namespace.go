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
		deployments, err := client.AppsV1().Deployments(metav1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			return "", errors.Wrapf(err, "could not list deployments")
		}
		for _, deployment := range deployments.Items {
			if deployment.Name == "supergloo" {
				return deployment.Namespace, nil
			}
		}

	}

	return "", errors.Errorf("could not determine supergloo namespace")
}
