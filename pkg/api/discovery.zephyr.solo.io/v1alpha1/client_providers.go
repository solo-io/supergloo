package v1alpha1

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provider for the discovery.zephyr.solo.io/v1alpha1 Clientset from config
func ClientsetFromConfigProvider(cfg *rest.Config) (Clientset, error) {
	return NewClientsetFromConfig(cfg)
}

// Provider for the discovery.zephyr.solo.io/v1alpha1 Clientset from client
func ClientsProvider(client client.Client) Clientset {
	return NewClientset(client)
}

// Provider for KubernetesClusterClient from Clientset
func KubernetesClusterClientFromClientsetProvider(clients Clientset) KubernetesClusterClient {
	return clients.KubernetesClusters()
}

// Provider for KubernetesClusterClient from Client
func KubernetesClusterClientProvider(client client.Client) KubernetesClusterClient {
	return NewKubernetesClusterClient(client)
}

type KubernetesClusterClientFactory func(client client.Client) KubernetesClusterClient

func KubernetesClusterClientFactoryProvider() KubernetesClusterClientFactory {
	return KubernetesClusterClientProvider
}

type KubernetesClusterClientFromConfigFactory func(cfg *rest.Config) (KubernetesClusterClient, error)

func KubernetesClusterClientFromConfigFactoryProvider() KubernetesClusterClientFromConfigFactory {
	return func(cfg *rest.Config) (KubernetesClusterClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.KubernetesClusters(), nil
	}
}

// Provider for MeshServiceClient from Clientset
func MeshServiceClientFromClientsetProvider(clients Clientset) MeshServiceClient {
	return clients.MeshServices()
}

// Provider for MeshServiceClient from Client
func MeshServiceClientProvider(client client.Client) MeshServiceClient {
	return NewMeshServiceClient(client)
}

type MeshServiceClientFactory func(client client.Client) MeshServiceClient

func MeshServiceClientFactoryProvider() MeshServiceClientFactory {
	return MeshServiceClientProvider
}

type MeshServiceClientFromConfigFactory func(cfg *rest.Config) (MeshServiceClient, error)

func MeshServiceClientFromConfigFactoryProvider() MeshServiceClientFromConfigFactory {
	return func(cfg *rest.Config) (MeshServiceClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.MeshServices(), nil
	}
}

// Provider for MeshWorkloadClient from Clientset
func MeshWorkloadClientFromClientsetProvider(clients Clientset) MeshWorkloadClient {
	return clients.MeshWorkloads()
}

// Provider for MeshWorkloadClient from Client
func MeshWorkloadClientProvider(client client.Client) MeshWorkloadClient {
	return NewMeshWorkloadClient(client)
}

type MeshWorkloadClientFactory func(client client.Client) MeshWorkloadClient

func MeshWorkloadClientFactoryProvider() MeshWorkloadClientFactory {
	return MeshWorkloadClientProvider
}

type MeshWorkloadClientFromConfigFactory func(cfg *rest.Config) (MeshWorkloadClient, error)

func MeshWorkloadClientFromConfigFactoryProvider() MeshWorkloadClientFromConfigFactory {
	return func(cfg *rest.Config) (MeshWorkloadClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.MeshWorkloads(), nil
	}
}

// Provider for MeshClient from Clientset
func MeshClientFromClientsetProvider(clients Clientset) MeshClient {
	return clients.Meshes()
}

// Provider for MeshClient from Client
func MeshClientProvider(client client.Client) MeshClient {
	return NewMeshClient(client)
}

type MeshClientFactory func(client client.Client) MeshClient

func MeshClientFactoryProvider() MeshClientFactory {
	return MeshClientProvider
}

type MeshClientFromConfigFactory func(cfg *rest.Config) (MeshClient, error)

func MeshClientFromConfigFactoryProvider() MeshClientFromConfigFactory {
	return func(cfg *rest.Config) (MeshClient, error) {
		clients, err := NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.Meshes(), nil
	}
}
