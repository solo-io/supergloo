// Code generated by skv2. DO NOT EDIT.

package v1alpha1

import (
	observability_enterprise_mesh_gloo_solo_io_v1alpha1 "github.com/solo-io/gloo-mesh/pkg/api/observability.enterprise.mesh.gloo.solo.io/v1alpha1"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
  The intention of these providers are to be used for Mocking.
  They expose the Clients as interfaces, as well as factories to provide mocked versions
  of the clients when they require building within a component.

  See package `github.com/solo-io/skv2/pkg/multicluster/register` for example
*/

// Provider for AccessLogRecordClient from Clientset
func AccessLogRecordClientFromClientsetProvider(clients observability_enterprise_mesh_gloo_solo_io_v1alpha1.Clientset) observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogRecordClient {
	return clients.AccessLogRecords()
}

// Provider for AccessLogRecord Client from Client
func AccessLogRecordClientProvider(client client.Client) observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogRecordClient {
	return observability_enterprise_mesh_gloo_solo_io_v1alpha1.NewAccessLogRecordClient(client)
}

type AccessLogRecordClientFactory func(client client.Client) observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogRecordClient

func AccessLogRecordClientFactoryProvider() AccessLogRecordClientFactory {
	return AccessLogRecordClientProvider
}

type AccessLogRecordClientFromConfigFactory func(cfg *rest.Config) (observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogRecordClient, error)

func AccessLogRecordClientFromConfigFactoryProvider() AccessLogRecordClientFromConfigFactory {
	return func(cfg *rest.Config) (observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogRecordClient, error) {
		clients, err := observability_enterprise_mesh_gloo_solo_io_v1alpha1.NewClientsetFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		return clients.AccessLogRecords(), nil
	}
}
