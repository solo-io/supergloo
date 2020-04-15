package v1alpha1

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provider for the split/v1alpha1 Clientset from config
func ClientsetFromConfigProvider(cfg *rest.Config) (Clientset, error) {
	return NewClientsetFromConfig(cfg)
}

// Provider for the split/v1alpha1 Clientset from client
func ClientsProvider(client client.Client) Clientset {
	return NewClientset(client)
}

// Provider for TrafficSplitClient from Clientset
func TrafficSplitClientFromClientsetProvider(clients Clientset) TrafficSplitClient {
	return clients.TrafficSplits()
}

// Provider for TrafficSplitClient from Client
func TrafficSplitClientProvider(client client.Client) TrafficSplitClient {
	return NewTrafficSplitClient(client)
}

type TrafficSplitClientFactory func(client client.Client) TrafficSplitClient

func TrafficSplitClientFactoryProvider() TrafficSplitClientFactory {
	return TrafficSplitClientProvider
}

type TrafficSplitClientFromConfigFactory func(cfg *rest.Config) (TrafficSplitClient, error)

func TrafficSplitClientFromConfigFactoryProvider() TrafficSplitClientFromConfigFactory {
	return func(cfg *rest.Config) (TrafficSplitClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.TrafficSplits(), nil
	}
}
