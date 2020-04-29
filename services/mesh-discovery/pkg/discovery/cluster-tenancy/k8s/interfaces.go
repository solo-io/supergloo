package k8s_tenancy

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_controllers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	k8s_core "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type ClusterTenancyFinder interface {
	StartDiscovery(
		ctx context.Context,
		podEventWatcher k8s_core_controller.PodEventWatcher,
		meshEventWatcher discovery_controllers.MeshEventWatcher,
	) error
}

type ClusterTenancyScanner interface {
	// Scan the pod for existence of a Mesh, update the relevant Mesh CRD only if it already exists. Otherwise do nothing.
	UpdateMeshTenancy(
		ctx context.Context,
		clusterName string,
		pod *k8s_core.Pod,
	) error
}

type ClusterTenancyScannerFactory func(meshClient zephyr_discovery.MeshClient) ClusterTenancyScanner
