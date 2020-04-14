package controllers

import (
	discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	apps_controllers "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1/controller"
	core_controllers "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

//go:generate mockgen -destination ./mocks/mock_deployment_event_watcher.go -package mock_controllers "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1/controller" DeploymentEventWatcher
//go:generate mockgen -destination ./mocks/mock_pod_event_watcher.go -package mock_controllers github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller PodEventWatcher

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go -package mock_controllers

// given a manager that can talk to a cluster and a name for that cluster, produce a `DeploymentEventWatcher`
type DeploymentEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) apps_controllers.DeploymentEventWatcher
}

// given a manager that can talk to a cluster and a name for that cluster, produce a `PodEventWatcher`
type PodEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) core_controllers.PodEventWatcher
}

type MeshWorkloadEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) discovery_controller.MeshWorkloadEventWatcher
}

type ServiceEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) core_controllers.ServiceEventWatcher
}
