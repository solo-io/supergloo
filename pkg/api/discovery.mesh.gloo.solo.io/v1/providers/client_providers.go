// Code generated by skv2. DO NOT EDIT.

package v1

import (
	discovery_mesh_gloo_solo_io_v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
  The intention of these providers are to be used for Mocking.
  They expose the Clients as interfaces, as well as factories to provide mocked versions
  of the clients when they require building within a component.

  See package `github.com/solo-io/skv2/pkg/multicluster/register` for example
*/

// Provider for DestinationClient from Clientset
func DestinationClientFromClientsetProvider(clients discovery_mesh_gloo_solo_io_v1.Clientset) discovery_mesh_gloo_solo_io_v1.DestinationClient {
	return clients.Destinations()
}

// Provider for Destination Client from Client
func DestinationClientProvider(client client.Client) discovery_mesh_gloo_solo_io_v1.DestinationClient {
	return discovery_mesh_gloo_solo_io_v1.NewDestinationClient(client)
}

type DestinationClientFactory func(client client.Client) discovery_mesh_gloo_solo_io_v1.DestinationClient

func DestinationClientFactoryProvider() DestinationClientFactory {
	return DestinationClientProvider
}

type DestinationClientFromConfigFactory func(cfg *rest.Config) (discovery_mesh_gloo_solo_io_v1.DestinationClient, error)

func DestinationClientFromConfigFactoryProvider() DestinationClientFromConfigFactory {
	return func(cfg *rest.Config) (discovery_mesh_gloo_solo_io_v1.DestinationClient, error) {
		clients, err := discovery_mesh_gloo_solo_io_v1.NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.Destinations(), nil
	}
}

// Provider for WorkloadClient from Clientset
func WorkloadClientFromClientsetProvider(clients discovery_mesh_gloo_solo_io_v1.Clientset) discovery_mesh_gloo_solo_io_v1.WorkloadClient {
	return clients.Workloads()
}

// Provider for Workload Client from Client
func WorkloadClientProvider(client client.Client) discovery_mesh_gloo_solo_io_v1.WorkloadClient {
	return discovery_mesh_gloo_solo_io_v1.NewWorkloadClient(client)
}

type WorkloadClientFactory func(client client.Client) discovery_mesh_gloo_solo_io_v1.WorkloadClient

func WorkloadClientFactoryProvider() WorkloadClientFactory {
	return WorkloadClientProvider
}

type WorkloadClientFromConfigFactory func(cfg *rest.Config) (discovery_mesh_gloo_solo_io_v1.WorkloadClient, error)

func WorkloadClientFromConfigFactoryProvider() WorkloadClientFromConfigFactory {
	return func(cfg *rest.Config) (discovery_mesh_gloo_solo_io_v1.WorkloadClient, error) {
		clients, err := discovery_mesh_gloo_solo_io_v1.NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.Workloads(), nil
	}
}

// Provider for MeshClient from Clientset
func MeshClientFromClientsetProvider(clients discovery_mesh_gloo_solo_io_v1.Clientset) discovery_mesh_gloo_solo_io_v1.MeshClient {
	return clients.Meshes()
}

// Provider for Mesh Client from Client
func MeshClientProvider(client client.Client) discovery_mesh_gloo_solo_io_v1.MeshClient {
	return discovery_mesh_gloo_solo_io_v1.NewMeshClient(client)
}

type MeshClientFactory func(client client.Client) discovery_mesh_gloo_solo_io_v1.MeshClient

func MeshClientFactoryProvider() MeshClientFactory {
	return MeshClientProvider
}

type MeshClientFromConfigFactory func(cfg *rest.Config) (discovery_mesh_gloo_solo_io_v1.MeshClient, error)

func MeshClientFromConfigFactoryProvider() MeshClientFromConfigFactory {
	return func(cfg *rest.Config) (discovery_mesh_gloo_solo_io_v1.MeshClient, error) {
		clients, err := discovery_mesh_gloo_solo_io_v1.NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.Meshes(), nil
	}
}
