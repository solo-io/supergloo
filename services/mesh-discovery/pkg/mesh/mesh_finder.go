package mesh

import (
	"context"
	"fmt"

	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/pkg/logging"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/mesh-projects/services/common/cluster/apps/v1/controller"
	"go.uber.org/zap"
	k8s_api_v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// a `MeshFinder` receives deployment events from a controller that it gets attached to
// this is the abstraction that should be directly managing the Mesh CR instances, based on what
// its `meshScanners` determine
func DefaultMeshFinder(
	ctx context.Context,
	clusterName string,
	meshFinders []MeshScanner,
	localMeshClient zephyr_core.MeshClient,
) controller.DeploymentEventHandler {
	return &meshFinder{
		clusterName:     clusterName,
		meshScanners:    meshFinders,
		localMeshClient: localMeshClient,
		ctx:             ctx,
	}
}

type meshFinder struct {
	clusterName     string
	meshScanners    []MeshScanner
	localMeshClient zephyr_core.MeshClient
	ctx             context.Context
}

func (d *meshFinder) Create(deployment *k8s_api_v1.Deployment) error {
	logger := logging.BuildEventLogger(d.ctx, logging.CreateEvent, d.clusterName)

	deployment.SetClusterName(d.clusterName)

	discoveredMesh, err := d.discoverMesh(deployment)
	if err != nil && discoveredMesh == nil {
		logger.Errorw("Error processing deployment for mesh discovery", zap.Error(err))
		return err
	} else if err != nil && discoveredMesh != nil {
		logger.Warnw("Non-fatal error occurred while scanning for mesh installations", zap.Error(err))
	} else if discoveredMesh == nil {
		return nil
	}

	err = d.createIfNotExists(discoveredMesh)
	if err != nil {
		logger.Errorw(fmt.Sprintf("Error creating Mesh CR for deployment %+v", deployment), zap.Error(err))
	}

	return err
}

func (d *meshFinder) Update(old, new *k8s_api_v1.Deployment) error {
	logger := logging.BuildEventLogger(d.ctx, logging.UpdateEvent, d.clusterName)

	old.SetClusterName(d.clusterName)
	new.SetClusterName(d.clusterName)

	oldDiscoveredMesh, err := d.discoverMesh(old)
	if oldDiscoveredMesh == nil && err != nil {
		// if we don't know whether there was a mesh here previously, nothing else to do, so bail out
		//logger.Errorw("Failed to discover mesh from old deployment for update event", zap.Error(err))
		return err
	} else if oldDiscoveredMesh != nil && err != nil {
		logger.Warnw("Non-fatal error occurred while scanning old deployment", zap.Error(err))
	}

	newDiscoveredMesh, err := d.discoverMesh(new)
	if newDiscoveredMesh == nil && err != nil {
		logger.Errorw("Failed to discover mesh from new deployment", zap.Error(err))
		return err
	} else if newDiscoveredMesh != nil && err != nil {
		logger.Warnw("Non-fatal error occurred while scanning new deployment", zap.Error(err))
	}

	if oldDiscoveredMesh == nil && newDiscoveredMesh == nil {
		return nil
	} else if oldDiscoveredMesh == nil && newDiscoveredMesh != nil {
		return d.createIfNotExists(newDiscoveredMesh)
	} else if oldDiscoveredMesh != nil && newDiscoveredMesh == nil {
		// todo: delete
		return nil
	}

	// need to check both that the MeshSpec is equal, and that the ObjectMeta is equal
	// (eg, did the mesh move namespaces or change versions, both of which should be included in the name)
	meshesEqual := oldDiscoveredMesh.Spec.Equal(newDiscoveredMesh.Spec) && oldDiscoveredMesh.GetName() == newDiscoveredMesh.GetName()
	if meshesEqual {
		return nil
	}

	// TODO: What does the upgrade path for meshes look like? The way this is written,
	// this considers meshes of different versions as different entities
	// TODO: not deleting any entities for now
	//err = d.delete(oldDiscoveredMesh)
	//if err != nil {
	//	d.logger.Errorw("Failed to delete the old Mesh CR for update event", zap.Error(err))
	//	return err
	//}
	err = d.createIfNotExists(newDiscoveredMesh)
	if err != nil {
		logger.Errorw("Deleted the old Mesh CR, but failed to create the new Mesh CR", zap.Error(err))
		return err
	}

	return nil
}

func (d *meshFinder) Delete(deployment *k8s_api_v1.Deployment) error {
	// TODO: Not deleting any entities for now
	return nil
}

func (d *meshFinder) Generic(deployment *k8s_api_v1.Deployment) error {
	// not implemented- we haven't implemented generic events for this controller
	return nil
}

// if both `discoveredMesh` and `err` are non-nil, then `err` should be considered a non-fatal error
func (d *meshFinder) discoverMesh(deployment *k8s_api_v1.Deployment) (discoveredMesh *discoveryv1alpha1.Mesh, err error) {
	var multiErr *multierror.Error
	for _, meshFinder := range d.meshScanners {
		discoveredMesh, err = meshFinder.ScanDeployment(d.ctx, deployment)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		if discoveredMesh != nil {
			break
		}
	}

	return discoveredMesh, multiErr.ErrorOrNil()
}

func (d *meshFinder) createIfNotExists(discoveredMesh *discoveryv1alpha1.Mesh) error {
	objectKey, err := client.ObjectKeyFromObject(discoveredMesh)
	if err != nil {
		return err
	}
	_, err = d.localMeshClient.Get(d.ctx, objectKey)
	if errors.IsNotFound(err) {
		return d.localMeshClient.Create(d.ctx, discoveredMesh)
	} else if err != nil {
		return err
	}

	return nil
}

func (d *meshFinder) delete(mesh *discoveryv1alpha1.Mesh) error {
	return d.localMeshClient.Delete(d.ctx, mesh)
}
