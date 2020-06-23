package failover

import (
	"context"

	"github.com/hashicorp/go-multierror"
	v1alpha3 "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/providers"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/skv2/pkg/reconcile"
	istio_client_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	FailoverServiceLabels = map[string]string{
		"created-by": "smh-failover-service",
	}
)

type failoverServiceReconciler struct {
	ctx                      context.Context
	failoverServiceProcessor FailoverServiceProcessor
	failoverServiceClient    smh_networking.FailoverServiceClient
	meshServiceClient        smh_discovery.MeshServiceClient
	meshClient               smh_discovery.MeshClient
	kubeClusterClient        smh_discovery.KubernetesClusterClient
	virtualMeshClient        smh_networking.VirtualMeshClient

	mcClient                  multicluster.Client
	serviceEntryClientFactory v1alpha3.ServiceEntryClientFactory
	envoyFilterClientFactory  v1alpha3.EnvoyFilterClientFactory
}

type InputSnapshot struct {
	FailoverServices []*smh_networking.FailoverService
	MeshServices     []*smh_discovery.MeshService
	// For validation
	KubeClusters  []*smh_discovery.KubernetesCluster
	Meshes        []*smh_discovery.Mesh
	VirtualMeshes []*smh_networking.VirtualMesh
}

type OutputSnapshot struct {
	FailoverServices []*smh_networking.FailoverService
	MeshOutputs      MeshOutputs
}

type MeshOutputs struct {
	// Istio
	ServiceEntries []*istio_client_networking.ServiceEntry
	EnvoyFilters   []*istio_client_networking.EnvoyFilter
}

// Append entries from the given OutputSnapshot to this OutputSnapshot
func (this OutputSnapshot) append(that OutputSnapshot) {
	this.FailoverServices = append(this.FailoverServices, that.FailoverServices...)
	this.MeshOutputs.append(that.MeshOutputs)
}

func (this MeshOutputs) append(that MeshOutputs) {
	this.ServiceEntries = append(this.ServiceEntries, that.ServiceEntries...)
	this.EnvoyFilters = append(this.EnvoyFilters, that.EnvoyFilters...)
}

func (f *failoverServiceReconciler) ReconcileFailoverService(_ *smh_networking.FailoverService) (reconcile.Result, error) {
	inputSnapshot, err := f.buildInputSnapshot()
	if err != nil {
		return reconcile.Result{}, err
	}
	outputSnapshot := f.failoverServiceProcessor.Process(f.ctx, inputSnapshot)
	// Update status on all FailoverServices, and ensure FailoverService mesh-specific translated resources on remote clusters.
	return reconcile.Result{}, f.ensureOutputSnapshot(outputSnapshot)
}

// TODO replace this with a generated builder
func (f *failoverServiceReconciler) buildInputSnapshot() (InputSnapshot, error) {
	inputSnapshot := InputSnapshot{}
	// FailoverService
	failoverServiceList, err := f.failoverServiceClient.ListFailoverService(f.ctx)
	if err != nil {
		return InputSnapshot{}, err
	}
	for _, failoverService := range failoverServiceList.Items {
		failoverService := failoverService
		inputSnapshot.FailoverServices = append(inputSnapshot.FailoverServices, &failoverService)
	}
	// MeshService
	meshServiceList, err := f.meshServiceClient.ListMeshService(f.ctx)
	if err != nil {
		return InputSnapshot{}, err
	}
	for _, meshService := range meshServiceList.Items {
		meshService := meshService
		inputSnapshot.MeshServices = append(inputSnapshot.MeshServices, &meshService)
	}
	// Mesh
	meshList, err := f.meshClient.ListMesh(f.ctx)
	if err != nil {
		return InputSnapshot{}, err
	}
	for _, mesh := range meshList.Items {
		mesh := mesh
		inputSnapshot.Meshes = append(inputSnapshot.Meshes, &mesh)
	}
	// KubernetesCluster
	kubeClusterList, err := f.kubeClusterClient.ListKubernetesCluster(f.ctx)
	if err != nil {
		return InputSnapshot{}, err
	}
	for _, kubeCluster := range kubeClusterList.Items {
		kubeCluster := kubeCluster
		inputSnapshot.KubeClusters = append(inputSnapshot.KubeClusters, &kubeCluster)
	}
	// VirtualMesh
	virtualMeshList, err := f.virtualMeshClient.ListVirtualMesh(f.ctx)
	if err != nil {
		return InputSnapshot{}, err
	}
	for _, virtualMesh := range virtualMeshList.Items {
		virtualMesh := virtualMesh
		inputSnapshot.VirtualMeshes = append(inputSnapshot.VirtualMeshes, &virtualMesh)
	}
	return inputSnapshot, nil
}

// Ensure that the actual state matches the desired state in the OutputSnapshot on each remote cluster.
func (f *failoverServiceReconciler) ensureOutputSnapshot(
	snapshot OutputSnapshot,
) error {
	var multierr *multierror.Error
	// Update FailoverService statuses
	for _, failoverService := range snapshot.FailoverServices {
		err := f.failoverServiceClient.UpdateFailoverServiceStatus(f.ctx, failoverService)
		if err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}
	// Upsert Istio resources
	if err := f.ensureServiceEntries(snapshot.MeshOutputs.ServiceEntries); err != nil {
		multierr = multierror.Append(multierr, err)
	}
	if err := f.ensureEnvoyFilters(snapshot.MeshOutputs.EnvoyFilters); err != nil {
		multierr = multierror.Append(multierr, err)
	}
	return multierr.ErrorOrNil()
}

func (f *failoverServiceReconciler) ensureServiceEntries(
	serviceEntries []*istio_client_networking.ServiceEntry,
) error {
	serviceEntriesByCluster := map[string][]*istio_client_networking.ServiceEntry{}
	for _, serviceEntry := range serviceEntries {
		f.addScopeLabels(serviceEntry)
		_, ok := serviceEntriesByCluster[serviceEntry.GetClusterName()]
		if !ok {
			serviceEntriesByCluster[serviceEntry.GetClusterName()] = []*istio_client_networking.ServiceEntry{}
		}
		serviceEntriesByCluster[serviceEntry.GetClusterName()] = append(serviceEntriesByCluster[serviceEntry.GetClusterName()], serviceEntry)
	}
	var multierr *multierror.Error
	// Reconcile per cluster
	for clusterName, serviceEntries := range serviceEntriesByCluster {
		clusterClient, err := f.mcClient.Cluster(clusterName)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		serviceEntryClient := f.serviceEntryClientFactory(clusterClient)
		desiredServiceEntries := v1alpha3sets.NewServiceEntrySet(serviceEntries...)
		existingServiceEntries := v1alpha3sets.NewServiceEntrySet()
		existingServiceEntriesList, err := serviceEntryClient.ListServiceEntry(f.ctx, client.MatchingLabels(FailoverServiceLabels))
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		for _, serviceEntry := range existingServiceEntriesList.Items {
			serviceEntry := serviceEntry
			existingServiceEntries.Insert(&serviceEntry)
		}
		// Upsert
		for _, desiredServiceEntry := range desiredServiceEntries.List() {
			err := serviceEntryClient.UpsertServiceEntry(f.ctx, desiredServiceEntry)
			if err != nil {
				multierr = multierror.Append(multierr, err)
				continue
			}
		}
		// Delete
		for _, existingServiceEntry := range existingServiceEntries.Difference(desiredServiceEntries).List() {
			err := serviceEntryClient.DeleteServiceEntry(f.ctx, selection.ObjectMetaToObjectKey(existingServiceEntry.ObjectMeta))
			if err != nil {
				multierr = multierror.Append(multierr, err)
				continue
			}
		}
	}
	return multierr.ErrorOrNil()
}

func (f *failoverServiceReconciler) ensureEnvoyFilters(
	envoyFilters []*istio_client_networking.EnvoyFilter,
) error {
	envoyFiltersByCluster := map[string][]*istio_client_networking.EnvoyFilter{}
	for _, envoyFilter := range envoyFilters {
		f.addScopeLabels(envoyFilter)
		_, ok := envoyFiltersByCluster[envoyFilter.GetClusterName()]
		if !ok {
			envoyFiltersByCluster[envoyFilter.GetClusterName()] = []*istio_client_networking.EnvoyFilter{}
		}
		envoyFiltersByCluster[envoyFilter.GetClusterName()] = append(envoyFiltersByCluster[envoyFilter.GetClusterName()], envoyFilter)
	}
	var multierr *multierror.Error
	// Reconcile per cluster
	for clusterName, envoyFilters := range envoyFiltersByCluster {
		clusterClient, err := f.mcClient.Cluster(clusterName)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		envoyFilterClient := f.envoyFilterClientFactory(clusterClient)
		desiredEnvoyFilters := v1alpha3sets.NewEnvoyFilterSet(envoyFilters...)
		existingEnvoyFilters := v1alpha3sets.NewEnvoyFilterSet()
		existingEnvoyFiltersList, err := envoyFilterClient.ListEnvoyFilter(f.ctx, client.MatchingLabels(FailoverServiceLabels))
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		for _, envoyFilter := range existingEnvoyFiltersList.Items {
			envoyFilter := envoyFilter
			existingEnvoyFilters.Insert(&envoyFilter)
		}
		// Upsert
		for _, desiredEnvoyFilter := range desiredEnvoyFilters.List() {
			err := envoyFilterClient.UpsertEnvoyFilter(f.ctx, desiredEnvoyFilter)
			if err != nil {
				multierr = multierror.Append(multierr, err)
				continue
			}
		}
		// Delete
		for _, existingEnvoyFilter := range existingEnvoyFilters.Difference(desiredEnvoyFilters).List() {
			err := envoyFilterClient.DeleteEnvoyFilter(f.ctx, selection.ObjectMetaToObjectKey(existingEnvoyFilter.ObjectMeta))
			if err != nil {
				multierr = multierror.Append(multierr, err)
				continue
			}
		}
	}
	return multierr.ErrorOrNil()
}

func (f *failoverServiceReconciler) addScopeLabels(obj v1.Object) {
	labels := obj.GetLabels()
	for k, v := range FailoverServiceLabels {
		labels[k] = v
	}
	obj.SetLabels(labels)
}
