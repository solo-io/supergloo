package mesh_service

import (
	corev1_controllers "UpdateMeshService"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go -package service_discovery_mocks

type MeshServiceFinder interface {
	StartDiscovery(serviceEventWatcher corev1_controllers.ServiceEventWatcher, meshWorkloadEventWatcher controller.MeshWorkloadEventWatcher) error
}
