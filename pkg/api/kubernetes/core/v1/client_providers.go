package v1

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provider for the core/v1 Clientset from config
func ClientsetFromConfigProvider(cfg *rest.Config) (Clientset, error) {
	return NewClientsetFromConfig(cfg)
}

// Provider for the core/v1 Clientset from client
func ClientsProvider(client client.Client) Clientset {
	return NewClientset(client)
}

// Provider for SecretClient from Clientset
func SecretClientFromClientsetProvider(clients Clientset) SecretClient {
	return clients.Secrets()
}

// Provider for SecretClient from Client
func SecretClientProvider(client client.Client) SecretClient {
	return NewSecretClient(client)
}

type SecretClientFactory func(client client.Client) SecretClient

func SecretClientFactoryProvider() SecretClientFactory {
	return SecretClientProvider
}

type SecretClientFromConfigFactory func(cfg *rest.Config) (SecretClient, error)

func SecretClientFromConfigFactoryProvider() SecretClientFromConfigFactory {
	return func(cfg *rest.Config) (SecretClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.Secrets(), nil
	}
}

// Provider for ServiceAccountClient from Clientset
func ServiceAccountClientFromClientsetProvider(clients Clientset) ServiceAccountClient {
	return clients.ServiceAccounts()
}

// Provider for ServiceAccountClient from Client
func ServiceAccountClientProvider(client client.Client) ServiceAccountClient {
	return NewServiceAccountClient(client)
}

type ServiceAccountClientFactory func(client client.Client) ServiceAccountClient

func ServiceAccountClientFactoryProvider() ServiceAccountClientFactory {
	return ServiceAccountClientProvider
}

type ServiceAccountClientFromConfigFactory func(cfg *rest.Config) (ServiceAccountClient, error)

func ServiceAccountClientFromConfigFactoryProvider() ServiceAccountClientFromConfigFactory {
	return func(cfg *rest.Config) (ServiceAccountClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.ServiceAccounts(), nil
	}
}

// Provider for ConfigMapClient from Clientset
func ConfigMapClientFromClientsetProvider(clients Clientset) ConfigMapClient {
	return clients.ConfigMaps()
}

// Provider for ConfigMapClient from Client
func ConfigMapClientProvider(client client.Client) ConfigMapClient {
	return NewConfigMapClient(client)
}

type ConfigMapClientFactory func(client client.Client) ConfigMapClient

func ConfigMapClientFactoryProvider() ConfigMapClientFactory {
	return ConfigMapClientProvider
}

type ConfigMapClientFromConfigFactory func(cfg *rest.Config) (ConfigMapClient, error)

func ConfigMapClientFromConfigFactoryProvider() ConfigMapClientFromConfigFactory {
	return func(cfg *rest.Config) (ConfigMapClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.ConfigMaps(), nil
	}
}

// Provider for ServiceClient from Clientset
func ServiceClientFromClientsetProvider(clients Clientset) ServiceClient {
	return clients.Services()
}

// Provider for ServiceClient from Client
func ServiceClientProvider(client client.Client) ServiceClient {
	return NewServiceClient(client)
}

type ServiceClientFactory func(client client.Client) ServiceClient

func ServiceClientFactoryProvider() ServiceClientFactory {
	return ServiceClientProvider
}

type ServiceClientFromConfigFactory func(cfg *rest.Config) (ServiceClient, error)

func ServiceClientFromConfigFactoryProvider() ServiceClientFromConfigFactory {
	return func(cfg *rest.Config) (ServiceClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.Services(), nil
	}
}

// Provider for PodClient from Clientset
func PodClientFromClientsetProvider(clients Clientset) PodClient {
	return clients.Pods()
}

// Provider for PodClient from Client
func PodClientProvider(client client.Client) PodClient {
	return NewPodClient(client)
}

type PodClientFactory func(client client.Client) PodClient

func PodClientFactoryProvider() PodClientFactory {
	return PodClientProvider
}

type PodClientFromConfigFactory func(cfg *rest.Config) (PodClient, error)

func PodClientFromConfigFactoryProvider() PodClientFromConfigFactory {
	return func(cfg *rest.Config) (PodClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.Pods(), nil
	}
}

// Provider for NamespaceClient from Clientset
func NamespaceClientFromClientsetProvider(clients Clientset) NamespaceClient {
	return clients.Namespaces()
}

// Provider for NamespaceClient from Client
func NamespaceClientProvider(client client.Client) NamespaceClient {
	return NewNamespaceClient(client)
}

type NamespaceClientFactory func(client client.Client) NamespaceClient

func NamespaceClientFactoryProvider() NamespaceClientFactory {
	return NamespaceClientProvider
}

type NamespaceClientFromConfigFactory func(cfg *rest.Config) (NamespaceClient, error)

func NamespaceClientFromConfigFactoryProvider() NamespaceClientFromConfigFactory {
	return func(cfg *rest.Config) (NamespaceClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.Namespaces(), nil
	}
}

// Provider for NodeClient from Clientset
func NodeClientFromClientsetProvider(clients Clientset) NodeClient {
	return clients.Nodes()
}

// Provider for NodeClient from Client
func NodeClientProvider(client client.Client) NodeClient {
	return NewNodeClient(client)
}

type NodeClientFactory func(client client.Client) NodeClient

func NodeClientFactoryProvider() NodeClientFactory {
	return NodeClientProvider
}

type NodeClientFromConfigFactory func(cfg *rest.Config) (NodeClient, error)

func NodeClientFromConfigFactoryProvider() NodeClientFromConfigFactory {
	return func(cfg *rest.Config) (NodeClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.Nodes(), nil
	}
}
