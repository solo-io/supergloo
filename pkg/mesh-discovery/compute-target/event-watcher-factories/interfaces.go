package event_watcher_factories

import (
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	k8s_apps_controller "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/controller"
	k8s_core_controller "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
)

//go:generate mockgen -destination ./mocks/mock_deployment_event_watcher.go -package mock_controllers "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/controller" DeploymentEventWatcher
//go:generate mockgen -destination ./mocks/mock_pod_event_watcher.go -package mock_controllers github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller PodEventWatcher

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go -package mock_controllers

// given a manager that can talk to a cluster and a name for that cluster, produce a `DeploymentEventWatcher`
type DeploymentEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) k8s_apps_controller.DeploymentEventWatcher
}

// given a manager that can talk to a cluster and a name for that cluster, produce a `PodEventWatcher`
type PodEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) k8s_core_controller.PodEventWatcher
}

type MeshEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) smh_discovery_controller.MeshEventWatcher
}

type MeshWorkloadEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) smh_discovery_controller.MeshWorkloadEventWatcher
}

type ServiceEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) k8s_core_controller.ServiceEventWatcher
}
