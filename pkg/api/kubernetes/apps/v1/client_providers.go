package v1

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provider for the apps/v1 Clientset from config
func ClientsetFromConfigProvider(cfg *rest.Config) (Clientset, error) {
	return NewClientsetFromConfig(cfg)
}

// Provider for the apps/v1 Clientset from client
func ClientsProvider(client client.Client) Clientset {
	return NewClientset(client)
}

// Provider for DeploymentClient from Clientset
func DeploymentClientFromClientsetProvider(clients Clientset) DeploymentClient {
	return clients.Deployments()
}

// Provider for DeploymentClient from Client
func DeploymentClientProvider(client client.Client) DeploymentClient {
	return NewDeploymentClient(client)
}

type DeploymentClientFactory func(client client.Client) DeploymentClient

func DeploymentClientFactoryProvider() DeploymentClientFactory {
	return DeploymentClientProvider
}

type DeploymentClientFromConfigFactory func(cfg *rest.Config) (DeploymentClient, error)

func DeploymentClientFromConfigFactoryProvider() DeploymentClientFromConfigFactory {
	return func(cfg *rest.Config) (DeploymentClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.Deployments(), nil
	}
}

// Provider for ReplicaSetClient from Clientset
func ReplicaSetClientFromClientsetProvider(clients Clientset) ReplicaSetClient {
	return clients.ReplicaSets()
}

// Provider for ReplicaSetClient from Client
func ReplicaSetClientProvider(client client.Client) ReplicaSetClient {
	return NewReplicaSetClient(client)
}

type ReplicaSetClientFactory func(client client.Client) ReplicaSetClient

func ReplicaSetClientFactoryProvider() ReplicaSetClientFactory {
	return ReplicaSetClientProvider
}

type ReplicaSetClientFromConfigFactory func(cfg *rest.Config) (ReplicaSetClient, error)

func ReplicaSetClientFromConfigFactoryProvider() ReplicaSetClientFromConfigFactory {
	return func(cfg *rest.Config) (ReplicaSetClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.ReplicaSets(), nil
	}
}
