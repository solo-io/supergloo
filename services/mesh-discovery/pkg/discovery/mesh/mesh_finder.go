package mesh

import (
	"context"
	"fmt"

	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/service-mesh-hub/services/common/cluster/apps/v1/controller"
	"go.uber.org/zap"
	apps_v1 "k8s.io/api/apps/v1"
)

// a `MeshFinder` receives deployment events from a controller that it gets attached to
// this is the abstraction that should be directly managing the Mesh CR instances, based on what
// its `meshScanners` determine
func NewMeshFinder(
	ctx context.Context,
	clusterName string,
	meshScanners []MeshScanner,
	localMeshClient zephyr_core.MeshClient,
	clusterClient client.Client,
) MeshFinder {
	return &meshFinder{
		clusterName:     clusterName,
		meshScanners:    meshScanners,
		localMeshClient: localMeshClient,
		ctx:             ctx,
		clusterClient:   clusterClient,
	}
}

type meshFinder struct {
	clusterName     string
	meshScanners    []MeshScanner
	localMeshClient zephyr_core.MeshClient
	ctx             context.Context
	clusterClient   client.Client
}

func (m *meshFinder) StartDiscovery(deploymentController controller.DeploymentController, predicates []predicate.Predicate) error {
	return deploymentController.AddEventHandler(
		m.ctx,
		m,
		predicates...,
	)
}

func (m *meshFinder) Create(deployment *apps_v1.Deployment) error {
	logger := logging.BuildEventLogger(m.ctx, logging.CreateEvent, deployment)
	return m.discoverAndUpsertMesh(deployment, logger)
}

func (m *meshFinder) Update(_, new *apps_v1.Deployment) error {
	logger := logging.BuildEventLogger(m.ctx, logging.UpdateEvent, new)
	return m.discoverAndUpsertMesh(new, logger)
}

func (m *meshFinder) Delete(deployment *apps_v1.Deployment) error {
	// TODO: Not deleting any entities for now
	return nil
}

func (m *meshFinder) Generic(deployment *apps_v1.Deployment) error {
	// not implemented- we haven't implemented generic events for this controller
	return nil
}

func (m *meshFinder) discoverAndUpsertMesh(deployment *apps_v1.Deployment, logger *zap.SugaredLogger) error {
	deployment.SetClusterName(m.clusterName)
	discoveredMesh, err := m.discoverMesh(deployment)
	if err != nil && discoveredMesh == nil {
		logger.Errorw("Error processing deployment for mesh discovery", zap.Error(err))
		return err
	} else if err != nil && discoveredMesh != nil {
		logger.Warnw("Non-fatal error occurred while scanning for mesh installations", zap.Error(err))
	} else if discoveredMesh == nil {
		return nil
	}
	err = m.localMeshClient.UpsertSpec(m.ctx, discoveredMesh)
	if err != nil {
		logger.Errorw(fmt.Sprintf("Error creating Mesh CR for deployment %+v", deployment), zap.Error(err))
	}
	return err
}

// if both `discoveredMesh` and `err` are non-nil, then `err` should be considered a non-fatal error
func (m *meshFinder) discoverMesh(deployment *apps_v1.Deployment) (discoveredMesh *discoveryv1alpha1.Mesh, err error) {
	var multiErr *multierror.Error
	for _, meshFinder := range m.meshScanners {
		discoveredMesh, err = meshFinder.ScanDeployment(m.ctx, deployment, m.clusterClient)
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
	return m.localMeshClient.Delete(m.ctx, mesh)
}
