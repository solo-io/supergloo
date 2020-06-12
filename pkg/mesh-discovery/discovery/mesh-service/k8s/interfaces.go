package k8s

import (
	k8s_core_controller "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go -package service_discovery_mocks

type MeshServiceFinder interface {
	StartDiscovery(serviceEventWatcher k8s_core_controller.ServiceEventWatcher, meshWorkloadEventWatcher smh_discovery_controller.MeshWorkloadEventWatcher) error
}
