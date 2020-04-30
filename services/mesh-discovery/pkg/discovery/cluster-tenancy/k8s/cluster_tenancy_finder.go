package k8s_tenancy

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	"github.com/solo-io/skv2/pkg/utils"
	k8s_core_types "k8s.io/api/core/v1"
)

type clusterTenancyFinder struct {
	clusterName     string
	tenancyScanners []ClusterTenancyScanner
	podClient       k8s_core.PodClient
	localMeshClient zephyr_discovery.MeshClient
}

func NewClusterTenancyFinder(
	clusterName string,
	tenancyScanners []ClusterTenancyScanner,
	podClient k8s_core.PodClient,
	localMeshClient zephyr_discovery.MeshClient,
) ClusterTenancyFinder {
	return &clusterTenancyFinder{
		clusterName:     clusterName,
		tenancyScanners: tenancyScanners,
		podClient:       podClient,
		localMeshClient: localMeshClient,
	}
}

func (c *clusterTenancyFinder) StartDiscovery(
	ctx context.Context,
	podEventWatcher k8s_core_controller.PodEventWatcher,
	meshEventWatcher zephyr_discovery_controller.MeshEventWatcher,
) (err error) {
	if err = podEventWatcher.AddEventHandler(ctx, &k8s_core_controller.PodEventHandlerFuncs{
		OnCreate: func(pod *k8s_core_types.Pod) error {
			logging.BuildEventLogger(ctx, logging.CreateEvent, pod).
				Debugf("Handling for %s.%s", pod.GetName(), pod.GetNamespace())
			return c.reconcileTenancyForPodUpsert(ctx, pod)
		},
		OnUpdate: func(_, pod *k8s_core_types.Pod) error {
			logging.BuildEventLogger(ctx, logging.UpdateEvent, pod).
				Debugf("Handling for %s.%s", pod.GetName(), pod.GetNamespace())
			return c.reconcileTenancyForPodUpsert(ctx, pod)
		},
		OnDelete: func(pod *k8s_core_types.Pod) error {
			logging.BuildEventLogger(ctx, logging.DeleteEvent, pod).
				Debugf("Handling for %s.%s", pod.GetName(), pod.GetNamespace())
			return c.reconcileTenancyForCluster(ctx)
		},
	}); err != nil {
		return err
	}
	return meshEventWatcher.AddEventHandler(ctx, &zephyr_discovery_controller.MeshEventHandlerFuncs{
		OnCreate: func(mesh *zephyr_discovery.Mesh) error {
			logging.BuildEventLogger(ctx, logging.CreateEvent, mesh).
				Debugf("Handling for %s.%s", mesh.GetName(), mesh.GetNamespace())
			return c.reconcileTenancyForMesh(ctx, mesh)
		},
		OnUpdate: func(_, mesh *zephyr_discovery.Mesh) error {
			logging.BuildEventLogger(ctx, logging.UpdateEvent, mesh).
				Debugf("Handling for %s.%s", mesh.GetName(), mesh.GetNamespace())
			return c.reconcileTenancyForMesh(ctx, mesh)
		},
	})
}

// Register cluster to Mesh tenancy if needed
func (c *clusterTenancyFinder) reconcileTenancyForPodUpsert(
	ctx context.Context,
	pod *k8s_core_types.Pod,
) error {
	for _, tenancyScanner := range c.tenancyScanners {
		err := tenancyScanner.UpdateMeshTenancy(ctx, c.clusterName, pod)
		if err != nil {
			return err
		}
	}
	return nil
}

// Recompute this cluster's tenancy for the given Mesh
func (c *clusterTenancyFinder) reconcileTenancyForMesh(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
) error {
	// Currently multicluster tenancy is only support for AppMesh
	if mesh.Spec.GetAwsAppMesh() == nil {
		return nil
	}
	// Remove this cluster from tenancy list and recompute in case this cluster no longer contains any
	// Mesh injected workloads.
	mesh.Spec.GetAwsAppMesh().Clusters = utils.RemoveString(mesh.Spec.GetAwsAppMesh().GetClusters(), c.clusterName)
	podList, err := c.podClient.ListPod(ctx)
	if err != nil {
		return err
	}
	for _, pod := range podList.Items {
		pod := pod
		err := c.reconcileTenancyForPodUpsert(ctx, &pod)
		if err != nil {
			return err
		}
	}
	return nil
}

// For all Meshes hosted by this cluster, recompute tenancy
func (c *clusterTenancyFinder) reconcileTenancyForCluster(ctx context.Context) error {
	meshList, err := c.localMeshClient.ListMesh(ctx)
	if err != nil {
		return err
	}
	for _, mesh := range meshList.Items {
		mesh := mesh
		// If Mesh has this cluster registered, delete it and recompute in case this cluster no longer contains any
		// Mesh injected workloads.
		if utils.ContainsString(mesh.Spec.GetAwsAppMesh().GetClusters(), c.clusterName) {
			err = c.reconcileTenancyForMesh(ctx, &mesh)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
