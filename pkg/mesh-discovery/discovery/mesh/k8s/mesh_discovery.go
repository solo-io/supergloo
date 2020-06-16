package k8s

import (
	"context"

	"github.com/hashicorp/go-multierror"
	k8s_apps_clients "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	k8s_core_clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
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
func NewMeshDiscovery(
	meshScanners []MeshScanner,
	meshClient smh_discovery.MeshClient,
	appsMulticlusterClientset k8s_apps_clients.MulticlusterClientset,
	coreMulticlusterClientset k8s_core_clients.MulticlusterClientset,
) MeshDiscovery {
	return &meshDiscovery{
		meshScanners:              meshScanners,
		meshClient:                meshClient,
		appsMulticlusterClientset: appsMulticlusterClientset,
		coreMulticlusterClientset: coreMulticlusterClientset,
	}
}

type meshDiscovery struct {
	meshScanners              []MeshScanner
	meshClient                smh_discovery.MeshClient
	appsMulticlusterClientset k8s_apps_clients.MulticlusterClientset
	coreMulticlusterClientset k8s_core_clients.MulticlusterClientset
}

func (m *meshDiscovery) DiscoverMesh(ctx context.Context, clusterName string) error {
	appsClientset, err := m.appsMulticlusterClientset.Cluster(clusterName)
	if err != nil {
		return err
	}
	coreClientset, err := m.coreMulticlusterClientset.Cluster(clusterName)
	if err != nil {
		return err
	}
	deploymentClient := appsClientset.Deployments()
	configMapClient := coreClientset.ConfigMaps()
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
		discoveredMesh, err := m.discoverMesh(ctx, clusterName, configMapClient, &deployment)
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
	configMapClient k8s_core_clients.ConfigMapClient,
	deployment *apps_v1.Deployment,
) (discoveredMesh *smh_discovery.Mesh, err error) {
	var multiErr *multierror.Error
	for _, meshFinder := range m.meshScanners {
		discoveredMesh, err = meshFinder.ScanDeployment(ctx, clusterName, deployment, configMapClient)
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
