package multicluster

import (
	ap_controllers "github.com/solo-io/mesh-projects/services/common/cluster/apps/v1/controller"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

//go:generate mockgen -source ./deployment_controller_factory.go -destination ./mocks/mock_deployment_controller_factory.go

// given a manager that can talk to a cluster and a name for that cluster, produce a `DeploymentController`
type DeploymentControllerFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (ap_controllers.DeploymentController, error)
}

func NewDeploymentControllerFactory() DeploymentControllerFactory {
	return &defaultDeploymentControllerFactory{}
}

type defaultDeploymentControllerFactory struct{}

func (d *defaultDeploymentControllerFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (ap_controllers.DeploymentController, error) {
	// just directly return the generated autopilot implementation
	return ap_controllers.NewDeploymentController(clusterName, mgr.Manager())
}
