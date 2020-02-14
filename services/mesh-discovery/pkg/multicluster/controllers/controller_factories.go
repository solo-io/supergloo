package controllers

import (
	apps_controllers "github.com/solo-io/mesh-projects/services/common/cluster/apps/v1/controller"
	core_controllers "github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

//go:generate mockgen -source ./controller_factories.go -destination ./mocks/mock_controller_factories.go

//go:generate mockgen -destination ./mocks/mock_pod_controller.go -package mock_controllers github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller PodController
// given a manager that can talk to a cluster and a name for that cluster, produce a `PodController`
type PodControllerFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (core_controllers.PodController, error)
}

func NewPodControllerFactory() PodControllerFactory {
	return &defaultPodControllerFactory{}
}

type defaultPodControllerFactory struct{}

func (d *defaultPodControllerFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (core_controllers.PodController, error) {
	return core_controllers.NewPodController(clusterName, mgr.Manager())
}

//go:generate mockgen -destination ./mocks/mock_deployment_controller.go -package mock_controllers github.com/solo-io/mesh-projects/services/common/cluster/apps/v1/controller DeploymentController
// given a manager that can talk to a cluster and a name for that cluster, produce a `DeploymentController`
type DeploymentControllerFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (apps_controllers.DeploymentController, error)
}

func NewDeploymentControllerFactory() DeploymentControllerFactory {
	return &defaultDeploymentControllerFactory{}
}

type defaultDeploymentControllerFactory struct{}

func (d *defaultDeploymentControllerFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (apps_controllers.DeploymentController, error) {
	// just directly return the generated autopilot implementation
	return apps_controllers.NewDeploymentController(clusterName, mgr.Manager())
}
