package healthcheck

import (
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/internal"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	kubernetes_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/discovery"
)

var (
	KubernetesAPI = healthcheck_types.Category{
		Name:   "Kubernetes API",
		Weight: 10,
	}
	ManagementPlane = healthcheck_types.Category{
		Name:   "Service Mesh Hub Management Plane",
		Weight: 9,
	}
	ServiceFederation = healthcheck_types.Category{
		Name:   "Service Federation",
		Weight: 8,
	}
)

func DefaultHealthChecksProvider() healthcheck_types.HealthCheckSuite {
	return map[healthcheck_types.Category][]healthcheck_types.HealthCheck{
		KubernetesAPI: {
			internal.NewKubeConnectivityCheck(), // make this one first- doesn't make sense to do the others if we can't talk to the api server
			internal.NewK8sServerVersionCheck(),
		},
		ManagementPlane: {
			internal.NewInstallNamespaceExistenceCheck(),
			internal.NewSmhComponentsHealthCheck(),
		},
		ServiceFederation: {
			internal.NewFederationDecisionCheck(),
		},
	}
}

func ClientsProvider(
	namespaceClient kubernetes_core.NamespaceClient,
	serverVersionClient kubernetes_discovery.ServerVersionClient,
	podClient kubernetes_core.PodClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
) healthcheck_types.Clients {
	return healthcheck_types.Clients{
		NamespaceClient:     namespaceClient,
		ServerVersionClient: serverVersionClient,
		PodClient:           podClient,
		MeshServiceClient:   meshServiceClient,
	}
}
