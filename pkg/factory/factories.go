package factory

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	istiosecret "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"

	"k8s.io/client-go/kubernetes"
)

func GetIstioCacertsSecretClient(clientset kubernetes.Interface) (istiosecret.IstioCacertsSecretClient, error) {
	return istiosecret.NewIstioCacertsSecretClient(&factory.KubeSecretClientFactory{
		Clientset:    clientset,
		PlainSecrets: true, // We need to use plain secrets for other systems (like istio) to be able to understand them
	})
}
