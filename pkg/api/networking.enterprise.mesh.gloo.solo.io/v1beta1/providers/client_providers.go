// Code generated by skv2. DO NOT EDIT.

package v1beta1



import (
    networking_enterprise_mesh_gloo_solo_io_v1beta1 "github.com/solo-io/gloo-mesh/pkg/api/networking.enterprise.mesh.gloo.solo.io/v1beta1"

    "k8s.io/client-go/rest"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

/*
  The intention of these providers are to be used for Mocking.
  They expose the Clients as interfaces, as well as factories to provide mocked versions
  of the clients when they require building within a component.

  See package `github.com/solo-io/skv2/pkg/multicluster/register` for example
*/

// Provider for WasmDeploymentClient from Clientset
func WasmDeploymentClientFromClientsetProvider(clients networking_enterprise_mesh_gloo_solo_io_v1beta1.Clientset) networking_enterprise_mesh_gloo_solo_io_v1beta1.WasmDeploymentClient {
    return clients.WasmDeployments()
}

// Provider for WasmDeployment Client from Client
func WasmDeploymentClientProvider(client client.Client) networking_enterprise_mesh_gloo_solo_io_v1beta1.WasmDeploymentClient {
    return networking_enterprise_mesh_gloo_solo_io_v1beta1.NewWasmDeploymentClient(client)
}

type WasmDeploymentClientFactory func(client client.Client) networking_enterprise_mesh_gloo_solo_io_v1beta1.WasmDeploymentClient

func WasmDeploymentClientFactoryProvider() WasmDeploymentClientFactory {
    return WasmDeploymentClientProvider
}

type WasmDeploymentClientFromConfigFactory func(cfg *rest.Config) (networking_enterprise_mesh_gloo_solo_io_v1beta1.WasmDeploymentClient, error)

func WasmDeploymentClientFromConfigFactoryProvider() WasmDeploymentClientFromConfigFactory {
    return func(cfg *rest.Config) (networking_enterprise_mesh_gloo_solo_io_v1beta1.WasmDeploymentClient, error) {
        clients, err := networking_enterprise_mesh_gloo_solo_io_v1beta1.NewClientsetFromConfig(cfg)
        if err != nil {
            return nil, err
        }
        return clients.WasmDeployments(), nil
    }
}

// Provider for VirtualDestinationClient from Clientset
func VirtualDestinationClientFromClientsetProvider(clients networking_enterprise_mesh_gloo_solo_io_v1beta1.Clientset) networking_enterprise_mesh_gloo_solo_io_v1beta1.VirtualDestinationClient {
    return clients.VirtualDestinations()
}

// Provider for VirtualDestination Client from Client
func VirtualDestinationClientProvider(client client.Client) networking_enterprise_mesh_gloo_solo_io_v1beta1.VirtualDestinationClient {
    return networking_enterprise_mesh_gloo_solo_io_v1beta1.NewVirtualDestinationClient(client)
}

type VirtualDestinationClientFactory func(client client.Client) networking_enterprise_mesh_gloo_solo_io_v1beta1.VirtualDestinationClient

func VirtualDestinationClientFactoryProvider() VirtualDestinationClientFactory {
    return VirtualDestinationClientProvider
}

type VirtualDestinationClientFromConfigFactory func(cfg *rest.Config) (networking_enterprise_mesh_gloo_solo_io_v1beta1.VirtualDestinationClient, error)

func VirtualDestinationClientFromConfigFactoryProvider() VirtualDestinationClientFromConfigFactory {
	return func(cfg *rest.Config) (networking_enterprise_mesh_gloo_solo_io_v1beta1.VirtualDestinationClient, error) {
		clients, err := networking_enterprise_mesh_gloo_solo_io_v1beta1.NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.VirtualDestinations(), nil
	}
}

// Provider for ServiceDependencyClient from Clientset
func ServiceDependencyClientFromClientsetProvider(clients networking_enterprise_mesh_gloo_solo_io_v1beta1.Clientset) networking_enterprise_mesh_gloo_solo_io_v1beta1.ServiceDependencyClient {
	return clients.ServiceDependencies()
}

// Provider for ServiceDependency Client from Client
func ServiceDependencyClientProvider(client client.Client) networking_enterprise_mesh_gloo_solo_io_v1beta1.ServiceDependencyClient {
	return networking_enterprise_mesh_gloo_solo_io_v1beta1.NewServiceDependencyClient(client)
}

type ServiceDependencyClientFactory func(client client.Client) networking_enterprise_mesh_gloo_solo_io_v1beta1.ServiceDependencyClient

func ServiceDependencyClientFactoryProvider() ServiceDependencyClientFactory {
	return ServiceDependencyClientProvider
}

type ServiceDependencyClientFromConfigFactory func(cfg *rest.Config) (networking_enterprise_mesh_gloo_solo_io_v1beta1.ServiceDependencyClient, error)

func ServiceDependencyClientFromConfigFactoryProvider() ServiceDependencyClientFromConfigFactory {
	return func(cfg *rest.Config) (networking_enterprise_mesh_gloo_solo_io_v1beta1.ServiceDependencyClient, error) {
		clients, err := networking_enterprise_mesh_gloo_solo_io_v1beta1.NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.ServiceDependencies(), nil
	}
}
