package controllers

import (
	discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	apps_controllers "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1/controller"
	core_controllers "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

//go:generate mockgen -destination ./mocks/mock_deployment_controller.go -package mock_controllers github.com/solo-io/service-mesh-hub/services/common/cluster/apps/v1/controller DeploymentController
//go:generate mockgen -destination ./mocks/mock_pod_controller.go -package mock_controllers github.com/solo-io/service-mesh-hub/services/common/cluster/core/v1/controller PodController

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go -package mock_controllers

// given a manager that can talk to a cluster and a name for that cluster, produce a `DeploymentController`
type DeploymentEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (apps_controllers.DeploymentEventWatcher, error)
}

// given a manager that can talk to a cluster and a name for that cluster, produce a `PodController`
type PodEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (core_controllers.PodEventWatcher, error)
}

type MeshWorkloadEventWatcherFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (discovery_controller.MeshWorkloadEventWatcher, error)
}

type ServiceControllerFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (core_controllers.ServiceEventWatcher, error)
}
