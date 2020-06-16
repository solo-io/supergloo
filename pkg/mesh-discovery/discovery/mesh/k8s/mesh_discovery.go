package k8s

import (
	"context"

	"github.com/hashicorp/go-multierror"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/providers"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	apps_v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// a `MeshFinder` receives deployment events from a controller that it gets attached to
// this is the abstraction that should be directly managing the Mesh CR instances, based on what
// its `meshScanners` determine
func NewMeshDiscovery(
	meshScanners []MeshScanner,
	meshClient smh_discovery.MeshClient,
	deploymentClientFactory v1.DeploymentClientFactory,
	dynamicClientGetter multicluster.DynamicClientGetter,
) MeshDiscovery {
	return &meshDiscovery{
		meshScanners:            meshScanners,
		meshClient:              meshClient,
		dynamicClientGetter:     dynamicClientGetter,
		deploymentClientFactory: deploymentClientFactory,
	}
}

type meshDiscovery struct {
	meshScanners            []MeshScanner
	meshClient              smh_discovery.MeshClient
	deploymentClientFactory v1.DeploymentClientFactory
	dynamicClientGetter     multicluster.DynamicClientGetter
}

func (m *meshDiscovery) DiscoverMesh(ctx context.Context, clusterName string) error {
	clusterClient, err := m.dynamicClientGetter.GetClientForCluster(ctx, clusterName)
	if err != nil {
		return err
	}
	deploymentClient := m.deploymentClientFactory(clusterClient)
	meshList, err := m.meshClient.ListMesh(ctx, client.MatchingLabels{kube.COMPUTE_TARGET: clusterName})
	if err != nil {
		return err
	}
	deploymentList, err := deploymentClient.ListDeployment(ctx)
	if err != nil {
		return err
	}
	discoveredMeshes := smh_discovery_sets.NewMeshSet()
	for _, deployment := range deploymentList.Items {
		deployment := deployment
		discoveredMesh, err := m.discoverMesh(ctx, clusterName, clusterClient, &deployment)
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
			err := m.meshClient.UpsertMesh(ctx, discoveredMesh)
			if err != nil {
				return err
			}
		}
	}
	// Delete existing Meshes that no longer exist
	for _, existingMesh := range existingMeshes.Difference(discoveredMeshes).List() {
		err := m.meshClient.DeleteMesh(ctx, selection.ObjectMetaToObjectKey(existingMesh.ObjectMeta))
		if err != nil {
			return err
		}
	}
	return nil
}

// If both `discoveredMesh` and `err` are non-nil, then `err` should be considered a non-fatal error
func (m *meshDiscovery) discoverMesh(
	ctx context.Context,
	clusterName string,
	clusterClient client.Client,
	deployment *apps_v1.Deployment,
) (discoveredMesh *smh_discovery.Mesh, err error) {
	var multiErr *multierror.Error
	for _, meshFinder := range m.meshScanners {
		discoveredMesh, err = meshFinder.ScanDeployment(ctx, clusterName, deployment, clusterClient)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		if discoveredMesh != nil {
			if discoveredMesh.Labels == nil {
				discoveredMesh.Labels = map[string]string{}
			}
			discoveredMesh.Labels[kube.COMPUTE_TARGET] = clusterName
			break
		}
	}
	return discoveredMesh, multiErr.ErrorOrNil()
}
