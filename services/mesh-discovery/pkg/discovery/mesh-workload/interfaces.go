package mesh_workload

import (
	"context"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	k8s_apps_types "k8s.io/api/apps/v1"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_mesh_interfaces.go -package mock_mesh_workload

// once StartDiscovery is invoked, MeshWorkloadFinder's PodEventHandler callbacks will start receiving PodEvents
type MeshWorkloadFinder interface {
	// an event is only received by our callbacks if all the given predicates return true
	StartDiscovery(podEventWatcher k8s_core_controller.PodEventWatcher, predicates []predicate.Predicate) error

	k8s_core_controller.PodEventHandler
}

// get a resource's controller- i.e., in the case of a pod, get its deployment
type OwnerFetcher interface {
	GetDeployment(ctx context.Context, pod *k8s_core_types.Pod) (*k8s_apps_types.Deployment, error)
}

// check a pod to see if it represents a mesh workload
// if it does, produce the appropriate controller reference, and object meta corresponding to it
type MeshWorkloadScanner interface {
	ScanPod(context.Context, *k8s_core_types.Pod) (*zephyr_core_types.ResourceRef, k8s_meta_types.ObjectMeta, error)
}

// these need to be constructed on the fly when a cluster is added, because the ownerFetcher will need to talk to that cluster
type MeshWorkloadScannerFactory func(ownerFetcher OwnerFetcher) MeshWorkloadScanner
