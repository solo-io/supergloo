package v1beta1

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provider for the security/v1beta1 Clientset from config
func ClientsetFromConfigProvider(cfg *rest.Config) (Clientset, error) {
	return NewClientsetFromConfig(cfg)
}

// Provider for the security/v1beta1 Clientset from client
func ClientsProvider(client client.Client) Clientset {
	return NewClientset(client)
}

// Provider for AuthorizationPolicyClient from Clientset
func AuthorizationPolicyClientFromClientsetProvider(clients Clientset) AuthorizationPolicyClient {
	return clients.AuthorizationPolicies()
}

// Provider for AuthorizationPolicyClient from Client
func AuthorizationPolicyClientProvider(client client.Client) AuthorizationPolicyClient {
	return NewAuthorizationPolicyClient(client)
}

type AuthorizationPolicyClientFactory func(client client.Client) AuthorizationPolicyClient

func AuthorizationPolicyClientFactoryProvider() AuthorizationPolicyClientFactory {
	return AuthorizationPolicyClientProvider
}

type AuthorizationPolicyClientFromConfigFactory func(cfg *rest.Config) (AuthorizationPolicyClient, error)

func AuthorizationPolicyClientFromConfigFactoryProvider() AuthorizationPolicyClientFromConfigFactory {
	return func(cfg *rest.Config) (AuthorizationPolicyClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.AuthorizationPolicies(), nil
	}
}
