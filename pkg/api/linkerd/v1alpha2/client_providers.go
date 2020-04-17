package v1alpha2

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provider for the /v1alpha2 Clientset from config
func ClientsetFromConfigProvider(cfg *rest.Config) (Clientset, error) {
	return NewClientsetFromConfig(cfg)
}

// Provider for the /v1alpha2 Clientset from client
func ClientsProvider(client client.Client) Clientset {
	return NewClientset(client)
}

// Provider for ServiceProfileClient from Clientset
func ServiceProfileClientFromClientsetProvider(clients Clientset) ServiceProfileClient {
	return clients.ServiceProfiles()
}

// Provider for ServiceProfileClient from Client
func ServiceProfileClientProvider(client client.Client) ServiceProfileClient {
	return NewServiceProfileClient(client)
}

type ServiceProfileClientFactory func(client client.Client) ServiceProfileClient

func ServiceProfileClientFactoryProvider() ServiceProfileClientFactory {
	return ServiceProfileClientProvider
}

type ServiceProfileClientFromConfigFactory func(cfg *rest.Config) (ServiceProfileClient, error)

func ServiceProfileClientFromConfigFactoryProvider() ServiceProfileClientFromConfigFactory {
	return func(cfg *rest.Config) (ServiceProfileClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.ServiceProfiles(), nil
	}
}
