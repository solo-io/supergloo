package mesh_workload

import (
	"context"

	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_controllers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	core_controllers "github.com/solo-io/service-mesh-hub/services/common/cluster/core/v1/controller"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_mesh_interfaces.go -package mock_mesh_workload

type MeshWorkloadScannerImplementations map[core_types.MeshType]MeshWorkloadScanner

type MeshWorkloadFinder interface {
	StartDiscovery(
		podController core_controllers.PodController,
		meshController discovery_controllers.MeshController,
	) error
}

// get a resource's controller- i.e., in the case of a pod, get its deployment
type OwnerFetcher interface {
	GetDeployment(ctx context.Context, pod *corev1.Pod) (*appsv1.Deployment, error)
}

// check a pod to see if it represents a mesh workload
// if it does, produce the appropriate controller reference, and object meta corresponding to it
type MeshWorkloadScanner interface {
	ScanPod(context.Context, *corev1.Pod) (*core_types.ResourceRef, metav1.ObjectMeta, error)
}

// these need to be constructed on the fly when a cluster is added, because the ownerFetcher will need to talk to that cluster
type MeshWorkloadScannerFactory func(ownerFetcher OwnerFetcher) MeshWorkloadScanner
