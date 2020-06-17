package k8s_tenancy

import (
	"context"

	k8s_core_controller "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discovery_controllers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
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
	MeshFromSidecar(ctx context.Context, pod *k8s_core.Pod, clusterName string) (*smh_discovery.Mesh, error)
	ClusterHostsMesh(clusterName string, mesh *smh_discovery.Mesh) bool
	RegisterMesh(ctx context.Context, clusterName string, mesh *smh_discovery.Mesh) error
	DeregisterMesh(ctx context.Context, clusterName string, mesh *smh_discovery.Mesh) error
}

type ClusterTenancyScannerFactory func(
	meshClient smh_discovery.MeshClient,
) ClusterTenancyRegistrar
