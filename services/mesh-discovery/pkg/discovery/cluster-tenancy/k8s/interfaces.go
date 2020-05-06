package k8s_tenancy

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_controllers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	k8s_core "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type ClusterTenancyRegistrarLoop interface {
	StartRegistration(
		ctx context.Context,
		podEventWatcher k8s_core_controller.PodEventWatcher,
		meshEventWatcher discovery_controllers.MeshEventWatcher,
	) error
}

type ClusterTenancyRegistrar interface {
	// Scan the pod for existence of a Mesh, return Mesh if found.
	MeshFromSidecar(ctx context.Context, pod *k8s_core.Pod) (*zephyr_discovery.Mesh, error)
	ClusterHostsMesh(clusterName string, mesh *zephyr_discovery.Mesh) bool
	RegisterMesh(ctx context.Context, clusterName string, mesh *zephyr_discovery.Mesh) error
	DeregisterMesh(ctx context.Context, clusterName string, mesh *zephyr_discovery.Mesh) error
}

type ClusterTenancyScannerFactory func(meshClient zephyr_discovery.MeshClient) ClusterTenancyRegistrar
