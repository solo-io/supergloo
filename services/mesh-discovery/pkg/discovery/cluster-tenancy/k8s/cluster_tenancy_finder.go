package k8s_tenancy

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	k8s_core_types "k8s.io/api/core/v1"
)

type clusterTenancyFinder struct {
	clusterName       string
	tenancyRegistrars []ClusterTenancyRegistrar
	podClient         k8s_core.PodClient
	localMeshClient   zephyr_discovery.MeshClient
}

func NewClusterTenancyFinder(
	clusterName string,
	tenancyRegistrars []ClusterTenancyRegistrar,
	podClient k8s_core.PodClient,
	localMeshClient zephyr_discovery.MeshClient,
) ClusterTenancyRegistrarLoop {
	return &clusterTenancyFinder{
		clusterName:       clusterName,
		tenancyRegistrars: tenancyRegistrars,
		podClient:         podClient,
		localMeshClient:   localMeshClient,
	}
}

func (c *clusterTenancyFinder) StartRegistration(
	ctx context.Context,
	podEventWatcher k8s_core_controller.PodEventWatcher,
	meshEventWatcher zephyr_discovery_controller.MeshEventWatcher,
) (err error) {
	if err = podEventWatcher.AddEventHandler(ctx, &k8s_core_controller.PodEventHandlerFuncs{
		OnCreate: func(pod *k8s_core_types.Pod) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, pod)
			logger.Debugf("Handling for %s.%s", pod.GetName(), pod.GetNamespace())
			err := c.reconcileTenancyForPod(ctx, pod)
			if err != nil {
				logger.Errorf("%+v", err)
			}
			return nil
		},
		OnUpdate: func(_, pod *k8s_core_types.Pod) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, pod)
			logger.Debugf("Handling for %s.%s", pod.GetName(), pod.GetNamespace())
			err := c.reconcileTenancyForPod(ctx, pod)
			if err != nil {
				logger.Errorf("%+v", err)
			}
			return nil
		},
		OnDelete: func(pod *k8s_core_types.Pod) error {
			logger := logging.BuildEventLogger(ctx, logging.DeleteEvent, pod)
			logger.Debugf("Handling for %s.%s", pod.GetName(), pod.GetNamespace())
			err := c.reconcileTenancyForCluster(ctx)
			if err != nil {
				logger.Errorf("%+v", err)
			}
			return nil
		},
	}); err != nil {
		return err
	}
	return meshEventWatcher.AddEventHandler(ctx, &zephyr_discovery_controller.MeshEventHandlerFuncs{
		OnCreate: func(mesh *zephyr_discovery.Mesh) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, mesh)
			logger.Debugf("Handling for %s.%s", mesh.GetName(), mesh.GetNamespace())
			err := c.reconcileTenancyForMesh(ctx, mesh)
			if err != nil {
				logger.Errorf("%+v", err)
			}
			return nil
		},
		OnUpdate: func(_, mesh *zephyr_discovery.Mesh) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, mesh)
			logger.Debugf("Handling for %s.%s", mesh.GetName(), mesh.GetNamespace())
			err := c.reconcileTenancyForMesh(ctx, mesh)
			if err != nil {
				logger.Errorf("%+v", err)
			}
			return nil
		},
	})
}

func (c *clusterTenancyFinder) updateClusterRegistryForMesh(
	ctx context.Context,
	registrar ClusterTenancyRegistrar,
	pod *k8s_core_types.Pod,
) error {
	mesh, err := registrar.MeshFromSidecar(ctx, pod)
	if err != nil {
		return err
	}
	if mesh == nil {
		return nil
	}
	return registrar.RegisterMesh(ctx, c.clusterName, mesh)
}

func (c *clusterTenancyFinder) reconcileTenancyForPod(
	ctx context.Context,
	pod *k8s_core_types.Pod,
) error {
	for _, registrar := range c.tenancyRegistrars {
		err := c.updateClusterRegistryForMesh(ctx, registrar, pod)
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
	// Remove this cluster from Mesh's cluster list and recompute in case this cluster no longer contains any
	// Mesh injected workloads.
	for _, registrar := range c.tenancyRegistrars {
		err := registrar.DeregisterMesh(ctx, c.clusterName, mesh)
		if err != nil {
			return err
		}
	}
	podList, err := c.podClient.ListPod(ctx)
	if err != nil {
		return err
	}
	for _, pod := range podList.Items {
		pod := pod
		err := c.reconcileTenancyForPod(ctx, &pod)
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
		if ClusterHostsMesh(c.clusterName, &mesh) {
			err = c.reconcileTenancyForMesh(ctx, &mesh)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
