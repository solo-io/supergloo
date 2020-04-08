package controllers

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	apps_controllers "github.com/solo-io/service-mesh-hub/services/common/cluster/apps/v1/controller"
	core_controllers "github.com/solo-io/service-mesh-hub/services/common/cluster/core/v1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

//go:generate mockgen -destination ./mocks/mock_deployment_controller.go -package mock_controllers github.com/solo-io/service-mesh-hub/services/common/cluster/apps/v1/controller DeploymentController
//go:generate mockgen -destination ./mocks/mock_pod_controller.go -package mock_controllers github.com/solo-io/service-mesh-hub/services/common/cluster/core/v1/controller PodController

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go -package mock_controllers

// given a manager that can talk to a cluster and a name for that cluster, produce a `DeploymentController`
type DeploymentControllerFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (apps_controllers.DeploymentController, error)
}

// given a manager that can talk to a cluster and a name for that cluster, produce a `PodController`
type PodControllerFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (core_controllers.PodController, error)
}

type MeshWorkloadControllerFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (controller.MeshWorkloadController, error)
}

type MeshServiceControllerFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (controller.MeshServiceController, error)
}

type ServiceControllerFactory interface {
	Build(mgr mc_manager.AsyncManager, clusterName string) (core_controllers.ServiceController, error)
}
