package controllers

import (
	apps_controllers "github.com/solo-io/mesh-projects/services/common/cluster/apps/v1/controller"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

func NewDeploymentControllerFactory() DeploymentControllerFactory {
	return &defaultDeploymentControllerFactory{}
}

type defaultDeploymentControllerFactory struct{}

func (d *defaultDeploymentControllerFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (apps_controllers.DeploymentController, error) {
	// just directly return the generated autopilot implementation
	return apps_controllers.NewDeploymentController(clusterName, mgr.Manager())
}
