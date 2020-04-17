package v1alpha3

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provider for the networking/v1alpha3 Clientset from config
func ClientsetFromConfigProvider(cfg *rest.Config) (Clientset, error) {
	return NewClientsetFromConfig(cfg)
}

// Provider for the networking/v1alpha3 Clientset from client
func ClientsProvider(client client.Client) Clientset {
	return NewClientset(client)
}

// Provider for DestinationRuleClient from Clientset
func DestinationRuleClientFromClientsetProvider(clients Clientset) DestinationRuleClient {
	return clients.DestinationRules()
}

// Provider for DestinationRuleClient from Client
func DestinationRuleClientProvider(client client.Client) DestinationRuleClient {
	return NewDestinationRuleClient(client)
}

type DestinationRuleClientFactory func(client client.Client) DestinationRuleClient

func DestinationRuleClientFactoryProvider() DestinationRuleClientFactory {
	return DestinationRuleClientProvider
}

type DestinationRuleClientFromConfigFactory func(cfg *rest.Config) (DestinationRuleClient, error)

func DestinationRuleClientFromConfigFactoryProvider() DestinationRuleClientFromConfigFactory {
	return func(cfg *rest.Config) (DestinationRuleClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.DestinationRules(), nil
	}
}

// Provider for EnvoyFilterClient from Clientset
func EnvoyFilterClientFromClientsetProvider(clients Clientset) EnvoyFilterClient {
	return clients.EnvoyFilters()
}

// Provider for EnvoyFilterClient from Client
func EnvoyFilterClientProvider(client client.Client) EnvoyFilterClient {
	return NewEnvoyFilterClient(client)
}

type EnvoyFilterClientFactory func(client client.Client) EnvoyFilterClient

func EnvoyFilterClientFactoryProvider() EnvoyFilterClientFactory {
	return EnvoyFilterClientProvider
}

type EnvoyFilterClientFromConfigFactory func(cfg *rest.Config) (EnvoyFilterClient, error)

func EnvoyFilterClientFromConfigFactoryProvider() EnvoyFilterClientFromConfigFactory {
	return func(cfg *rest.Config) (EnvoyFilterClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.EnvoyFilters(), nil
	}
}

// Provider for GatewayClient from Clientset
func GatewayClientFromClientsetProvider(clients Clientset) GatewayClient {
	return clients.Gateways()
}

// Provider for GatewayClient from Client
func GatewayClientProvider(client client.Client) GatewayClient {
	return NewGatewayClient(client)
}

type GatewayClientFactory func(client client.Client) GatewayClient

func GatewayClientFactoryProvider() GatewayClientFactory {
	return GatewayClientProvider
}

type GatewayClientFromConfigFactory func(cfg *rest.Config) (GatewayClient, error)

func GatewayClientFromConfigFactoryProvider() GatewayClientFromConfigFactory {
	return func(cfg *rest.Config) (GatewayClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.Gateways(), nil
	}
}

// Provider for ServiceEntryClient from Clientset
func ServiceEntryClientFromClientsetProvider(clients Clientset) ServiceEntryClient {
	return clients.ServiceEntries()
}

// Provider for ServiceEntryClient from Client
func ServiceEntryClientProvider(client client.Client) ServiceEntryClient {
	return NewServiceEntryClient(client)
}

type ServiceEntryClientFactory func(client client.Client) ServiceEntryClient

func ServiceEntryClientFactoryProvider() ServiceEntryClientFactory {
	return ServiceEntryClientProvider
}

type ServiceEntryClientFromConfigFactory func(cfg *rest.Config) (ServiceEntryClient, error)

func ServiceEntryClientFromConfigFactoryProvider() ServiceEntryClientFromConfigFactory {
	return func(cfg *rest.Config) (ServiceEntryClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.ServiceEntries(), nil
	}
}

// Provider for VirtualServiceClient from Clientset
func VirtualServiceClientFromClientsetProvider(clients Clientset) VirtualServiceClient {
	return clients.VirtualServices()
}

// Provider for VirtualServiceClient from Client
func VirtualServiceClientProvider(client client.Client) VirtualServiceClient {
	return NewVirtualServiceClient(client)
}

type VirtualServiceClientFactory func(client client.Client) VirtualServiceClient

func VirtualServiceClientFactoryProvider() VirtualServiceClientFactory {
	return VirtualServiceClientProvider
}

type VirtualServiceClientFromConfigFactory func(cfg *rest.Config) (VirtualServiceClient, error)

func VirtualServiceClientFromConfigFactoryProvider() VirtualServiceClientFromConfigFactory {
	return func(cfg *rest.Config) (VirtualServiceClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.VirtualServices(), nil
	}
}
