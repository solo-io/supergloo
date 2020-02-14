package controllers

import (
	ap_controllers "github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

//go:generate mockgen -destination ./mocks/mock_pod_controller.go -package mock_controllers github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller PodController
//go:generate mockgen -source ./pod_controller_factory.go -destination ./mocks/mock_pod_controller_factory.go

// given a manager that can talk to a cluster and a name for that cluster, produce a `PodController`
type PodControllerFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (ap_controllers.PodController, error)
}

func NewPodControllerFactory() PodControllerFactory {
	return &defaultPodControllerFactory{}
}

type defaultPodControllerFactory struct{}

func (d *defaultPodControllerFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (ap_controllers.PodController, error) {
	return ap_controllers.NewPodController(clusterName, mgr.Manager())
}
