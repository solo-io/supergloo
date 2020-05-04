package k8s

import (
	"context"

	"github.com/hashicorp/go-multierror"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	"github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	"go.uber.org/zap"
	apps_v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// a `MeshFinder` receives deployment events from a controller that it gets attached to
// this is the abstraction that should be directly managing the Mesh CR instances, based on what
// its `meshScanners` determine
func NewMeshFinder(
	ctx context.Context,
	clusterName string,
	meshScanners []MeshScanner,
	localMeshClient zephyr_discovery.MeshClient,
	clusterClient client.Client,
	clusterScopedDeploymentClient k8s_apps.DeploymentClient,
) MeshFinder {
	return &meshFinder{
		clusterName:                   clusterName,
		meshScanners:                  meshScanners,
		localMeshClient:               localMeshClient,
		ctx:                           ctx,
		clusterClient:                 clusterClient,
		clusterScopedDeploymentClient: clusterScopedDeploymentClient,
	}
}

type meshFinder struct {
	clusterName                   string
	meshScanners                  []MeshScanner
	localMeshClient               zephyr_discovery.MeshClient
	ctx                           context.Context
	clusterClient                 client.Client
	clusterScopedDeploymentClient k8s_apps.DeploymentClient
}

func (m *meshFinder) StartDiscovery(deploymentEventWatcher controller.DeploymentEventWatcher) error {
	// reconcile before adding the event handler
	err := m.reconcileExistingState()
	if err != nil {
		return err
	}

	return deploymentEventWatcher.AddEventHandler(
		m.ctx,
		m,
	)
}

func (m *meshFinder) CreateDeployment(deployment *apps_v1.Deployment) error {
	logger := logging.BuildEventLogger(m.ctx, logging.CreateEvent, deployment)
	return m.discoverAndUpsertMesh(deployment, logger)
}

func (m *meshFinder) UpdateDeployment(_, new *apps_v1.Deployment) error {
	logger := logging.BuildEventLogger(m.ctx, logging.UpdateEvent, new)
	return m.discoverAndUpsertMesh(new, logger)
}

func (m *meshFinder) DeleteDeployment(deployment *apps_v1.Deployment) error {
	logger := logging.BuildEventLogger(m.ctx, logging.DeleteEvent, deployment)

	discoveredMesh, err := m.discoverMesh(deployment)
	if err != nil {
		logger.Errorf("Error while attempting to discover mesh during delete: %+v", err)
		return nil
	}

	if discoveredMesh != nil {
		err = m.localMeshClient.DeleteMesh(m.ctx, clients.ObjectMetaToObjectKey(discoveredMesh.ObjectMeta))
		if err != nil {
			logger.Errorf("Error while deleting mesh: %+v", err)
		}
	}
	return nil
}

func (m *meshFinder) GenericDeployment(deployment *apps_v1.Deployment) error {
	// not implemented- we haven't implemented generic events for this controller
	return nil
}

// When the pod starts up, we reconcile the existing state of discovered resources with a newly-computed set of discovered resources.
// If the newly-computed set is missing entries from the current state, we must have missed an event, and we must reconcile the two.
func (m *meshFinder) reconcileExistingState() error {
	allMeshesOnCluster, err := m.localMeshClient.ListMesh(m.ctx, client.MatchingLabels{constants.COMPUTE_TARGET: m.clusterName})
	if err != nil {
		return err
	}

	if len(allMeshesOnCluster.Items) == 0 {
		// we have not discovered anything here yet, nothing to reconcile
		return nil
	}

	allDeployments, err := m.clusterScopedDeploymentClient.ListDeployment(m.ctx)
	if err != nil {
		return err
	}

	var recomputedMeshes []*discoveryv1alpha1.Mesh
	for _, deployment := range allDeployments.Items {
		discoveredMesh, err := m.discoverMesh(&deployment)
		if err != nil {
			return err
		}
		if discoveredMesh != nil {
			recomputedMeshes = append(recomputedMeshes, discoveredMesh)
		}
	}

	// ignore meshes that are in "newly computed" but not "recorded meshes"
	// those will be picked up by Create events rolling in when we start the handler after this function call
	for _, recordedMesh := range allMeshesOnCluster.Items {
		shouldDelete := true
		shouldUpdate := false
		for _, newlyComputedMesh := range recomputedMeshes {
			if newlyComputedMesh.GetName() == recordedMesh.GetName() {
				shouldDelete = false

				if !newlyComputedMesh.Spec.Equal(recordedMesh.Spec) {
					shouldUpdate = true
					recordedMesh.Spec = newlyComputedMesh.Spec
				}
			}
		}

		if shouldDelete {
			// we missed a delete event - clean up the state
			err := m.localMeshClient.DeleteMesh(m.ctx, clients.ObjectMetaToObjectKey(recordedMesh.ObjectMeta))
			if err != nil {
				return err
			}
		} else if shouldUpdate {
			// we missed an update event - reconcile the Mesh
			err := m.localMeshClient.UpdateMesh(m.ctx, &recordedMesh)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *meshFinder) discoverAndUpsertMesh(deployment *apps_v1.Deployment, logger *zap.SugaredLogger) error {
	discoveredMesh, err := m.discoverMesh(deployment)
	if err != nil && discoveredMesh == nil {
		logger.Errorw("Error processing deployment for mesh discovery",
			zap.Any("deployment", deployment),
			zap.Error(err),
		)
		return err
	} else if err != nil && discoveredMesh != nil {
		logger.Warnw("Non-fatal error occurred while scanning for mesh installations",
			zap.Any("deployment", deployment),
			zap.Error(err),
		)
	} else if discoveredMesh == nil {
		return nil
	}

	if discoveredMesh.Labels == nil {
		discoveredMesh.Labels = map[string]string{}
	}

	discoveredMesh.Labels[constants.COMPUTE_TARGET] = m.clusterName

	err = m.localMeshClient.UpsertMeshSpec(m.ctx, discoveredMesh)
	if err != nil {
		logger.Errorw("could not create Mesh CR for deployment",
			zap.Any("deployment", deployment),
			zap.Error(err),
		)
	}
	return err
}

// if both `discoveredMesh` and `err` are non-nil, then `err` should be considered a non-fatal error
func (m *meshFinder) discoverMesh(deployment *apps_v1.Deployment) (discoveredMesh *discoveryv1alpha1.Mesh, err error) {
	var multiErr *multierror.Error
	for _, meshFinder := range m.meshScanners {
		discoveredMesh, err = meshFinder.ScanDeployment(m.ctx, m.clusterName, deployment, m.clusterClient)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		if discoveredMesh != nil {
			break
		}
	}

	return discoveredMesh, multiErr.ErrorOrNil()
}

func (m *meshFinder) delete(mesh *discoveryv1alpha1.Mesh) error {
	return m.localMeshClient.DeleteMesh(m.ctx, client.ObjectKey{Name: mesh.GetName(), Namespace: mesh.GetNamespace()})
}
