package k8s

import (
	"context"

	"github.com/hashicorp/go-multierror"
	k8s_apps "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/skv2/contrib/pkg/sets"
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
	localMeshClient smh_discovery.MeshClient,
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
	localMeshClient               smh_discovery.MeshClient
	ctx                           context.Context
	clusterClient                 client.Client
	clusterScopedDeploymentClient k8s_apps.DeploymentClient
}

func (m *meshFinder) Process(ctx context.Context, clusterName string) error {
	meshList, err := m.localMeshClient.ListMesh(m.ctx, client.MatchingLabels{kube.COMPUTE_TARGET: m.clusterName})
	if err != nil {
		return err
	}
	deploymentList, err := m.clusterScopedDeploymentClient.ListDeployment(m.ctx)
	if err != nil {
		return err
	}
	discoveredMeshes := smh_discovery_sets.NewMeshSet()
	for _, deployment := range deploymentList.Items {
		deployment := deployment
		discoveredMesh, err := m.discoverMesh(&deployment)
		if err != nil {
			return err
		}
		if discoveredMesh != nil {
			discoveredMeshes.Insert(discoveredMesh)
		}
	}
	existingMeshes := smh_discovery_sets.NewMeshSet()
	for _, mesh := range meshList.Items {
		mesh := mesh
		existingMeshes.Insert(&mesh)
	}
	existingMeshMap := existingMeshes.Map()
	// Create or upsert discovered Meshes
	for _, discoveredMesh := range discoveredMeshes.List() {
		existingMesh, ok := existingMeshMap[sets.Key(discoveredMesh)]
		if !ok || !existingMesh.Spec.Equal(discoveredMesh.Spec) {
			err := m.localMeshClient.UpsertMesh(ctx, discoveredMesh)
			if err != nil {
				return err
			}
		}
	}
	// Delete existing Meshes that no longer exist
	for _, existingMesh := range existingMeshes.Difference(discoveredMeshes).List() {
		err := m.localMeshClient.DeleteMesh(ctx, selection.ObjectMetaToObjectKey(existingMesh.ObjectMeta))
		if err != nil {
			return err
		}
	}
	return nil
}

// If both `discoveredMesh` and `err` are non-nil, then `err` should be considered a non-fatal error
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
	if discoveredMesh.Labels == nil {
		discoveredMesh.Labels = map[string]string{}
	}
	discoveredMesh.Labels[kube.COMPUTE_TARGET] = m.clusterName
	return discoveredMesh, multiErr.ErrorOrNil()
}
