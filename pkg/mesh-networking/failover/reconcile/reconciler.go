package reconcile

import (
	"context"

	"github.com/hashicorp/go-multierror"
	v1alpha32 "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3"
	v1alpha3 "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/providers"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	"github.com/solo-io/go-utils/contextutils"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/controller"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/sets"
	multicluster2 "github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover"
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

	// TODO(harveyxia) replace with multicluster.Client after refactor
	dynamicClientGetter       multicluster2.DynamicClientGetter
	serviceEntryClientFactory v1alpha3.ServiceEntryClientFactory
	envoyFilterClientFactory  v1alpha3.EnvoyFilterClientFactory
}

func NewFailoverServiceReconciler(
	ctx context.Context,
	failoverServiceProcessor FailoverServiceProcessor,
	failoverServiceClient smh_networking.FailoverServiceClient,
	meshServiceClient smh_discovery.MeshServiceClient,
	meshClient smh_discovery.MeshClient,
	kubeClusterClient smh_discovery.KubernetesClusterClient,
	virtualMeshClient smh_networking.VirtualMeshClient,
	dynamicClientGetter multicluster2.DynamicClientGetter,
	serviceEntryClientFactory v1alpha3.ServiceEntryClientFactory,
	envoyFilterClientFactory v1alpha3.EnvoyFilterClientFactory,
) controller.FailoverServiceReconciler {
	return &failoverServiceReconciler{
		ctx:                       ctx,
		failoverServiceProcessor:  failoverServiceProcessor,
		failoverServiceClient:     failoverServiceClient,
		meshServiceClient:         meshServiceClient,
		meshClient:                meshClient,
		kubeClusterClient:         kubeClusterClient,
		virtualMeshClient:         virtualMeshClient,
		dynamicClientGetter:       dynamicClientGetter,
		serviceEntryClientFactory: serviceEntryClientFactory,
		envoyFilterClientFactory:  envoyFilterClientFactory,
	}
}

func (f *failoverServiceReconciler) ReconcileFailoverService(_ *smh_networking.FailoverService) (reconcile.Result, error) {
	inputSnapshot, err := f.buildInputSnapshot()
	if err != nil {
		return reconcile.Result{}, err
	}
	var clusterNames []string
	for _, kubeCluster := range inputSnapshot.KubeClusters.List() {
		clusterNames = append(clusterNames, kubeCluster.GetName())
	}
	outputSnapshot := f.failoverServiceProcessor.Process(f.ctx, inputSnapshot)
	// Update status on all FailoverServices, and ensure FailoverService mesh-specific translated resources on remote clusters.
	return reconcile.Result{}, f.ensureOutputSnapshot(f.ctx, clusterNames, outputSnapshot)
}

func (f *failoverServiceReconciler) ReconcileFailoverServiceDeletion(req reconcile.Request) {
	logger := contextutils.LoggerFrom(f.ctx)
	inputSnapshot, err := f.buildInputSnapshot()
	if err != nil {
		logger.Error("Error while reconciling FailoverService deletion: %+v.", err)
		return
	}
	var clusterNames []string
	for _, kubeCluster := range inputSnapshot.KubeClusters.List() {
		clusterNames = append(clusterNames, kubeCluster.GetName())
	}
	outputSnapshot := f.failoverServiceProcessor.Process(f.ctx, inputSnapshot)
	err = f.ensureOutputSnapshot(f.ctx, clusterNames, outputSnapshot)
	if err != nil {
		logger.Error("Error while reconciling FailoverService deletion: %+v.", err)
	}
}

// TODO replace this with a generated builder
func (f *failoverServiceReconciler) buildInputSnapshot() (failover.InputSnapshot, error) {
	inputSnapshot := failover.InputSnapshot{
		FailoverServices: v1alpha1sets.NewFailoverServiceSet(),
		MeshServices:     v1alpha1sets2.NewMeshServiceSet(),
		KubeClusters:     v1alpha1sets2.NewKubernetesClusterSet(),
		Meshes:           v1alpha1sets2.NewMeshSet(),
		VirtualMeshes:    v1alpha1sets.NewVirtualMeshSet(),
	}
	// FailoverService
	failoverServiceList, err := f.failoverServiceClient.ListFailoverService(f.ctx)
	if err != nil {
		return failover.InputSnapshot{}, err
	}
	for _, failoverService := range failoverServiceList.Items {
		failoverService := failoverService
		inputSnapshot.FailoverServices.Insert(&failoverService)
	}
	// MeshService
	meshServiceList, err := f.meshServiceClient.ListMeshService(f.ctx)
	if err != nil {
		return failover.InputSnapshot{}, err
	}
	for _, meshService := range meshServiceList.Items {
		meshService := meshService
		inputSnapshot.MeshServices.Insert(&meshService)
	}
	// Mesh
	meshList, err := f.meshClient.ListMesh(f.ctx)
	if err != nil {
		return failover.InputSnapshot{}, err
	}
	for _, mesh := range meshList.Items {
		mesh := mesh
		inputSnapshot.Meshes.Insert(&mesh)
	}
	// KubernetesCluster
	kubeClusterList, err := f.kubeClusterClient.ListKubernetesCluster(f.ctx)
	if err != nil {
		return failover.InputSnapshot{}, err
	}
	for _, kubeCluster := range kubeClusterList.Items {
		kubeCluster := kubeCluster
		inputSnapshot.KubeClusters.Insert(&kubeCluster)
	}
	// VirtualMesh
	virtualMeshList, err := f.virtualMeshClient.ListVirtualMesh(f.ctx)
	if err != nil {
		return failover.InputSnapshot{}, err
	}
	for _, virtualMesh := range virtualMeshList.Items {
		virtualMesh := virtualMesh
		inputSnapshot.VirtualMeshes.Insert(&virtualMesh)
	}
	return inputSnapshot, nil
}

// Ensure that the actual state matches the desired state in the OutputSnapshot on each remote cluster.
func (f *failoverServiceReconciler) ensureOutputSnapshot(
	ctx context.Context,
	clusterNames []string,
	snapshot failover.OutputSnapshot,
) error {
	var multierr *multierror.Error
	// Update FailoverService statuses
	for _, failoverService := range snapshot.FailoverServices.List() {
		err := f.failoverServiceClient.UpdateFailoverServiceStatus(f.ctx, failoverService)
		if err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}
	// Upsert Istio resources
	if err := f.ensureServiceEntries(ctx, clusterNames, snapshot.MeshOutputs.ServiceEntries); err != nil {
		multierr = multierror.Append(multierr, err)
	}
	if err := f.ensureEnvoyFilters(ctx, clusterNames, snapshot.MeshOutputs.EnvoyFilters); err != nil {
		multierr = multierror.Append(multierr, err)
	}
	return multierr.ErrorOrNil()
}

func (f *failoverServiceReconciler) ensureServiceEntries(
	ctx context.Context,
	clusterNames []string,
	serviceEntries v1alpha3sets.ServiceEntrySet,
) error {
	desiredServiceEntriesByCluster := map[string][]*istio_client_networking.ServiceEntry{}
	for _, serviceEntry := range serviceEntries.List() {
		f.addScopeLabels(serviceEntry)
		_, ok := desiredServiceEntriesByCluster[serviceEntry.GetClusterName()]
		if !ok {
			desiredServiceEntriesByCluster[serviceEntry.GetClusterName()] = []*istio_client_networking.ServiceEntry{}
		}
		desiredServiceEntriesByCluster[serviceEntry.GetClusterName()] = append(desiredServiceEntriesByCluster[serviceEntry.GetClusterName()], serviceEntry)
	}
	var multierr *multierror.Error
	// Reconcile per cluster
	for clusterName, desiredServiceEntries := range desiredServiceEntriesByCluster {
		clusterClient, err := f.dynamicClientGetter.GetClientForCluster(ctx, clusterName)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		serviceEntryClient := f.serviceEntryClientFactory(clusterClient)
		desiredServiceEntries := v1alpha3sets.NewServiceEntrySet(desiredServiceEntries...)
		// Upsert
		for _, desiredServiceEntry := range desiredServiceEntries.List() {
			err := serviceEntryClient.UpsertServiceEntry(f.ctx, desiredServiceEntry)
			if err != nil {
				multierr = multierror.Append(multierr, err)
				continue
			}
		}
	}
	// Delete
	for _, clusterName := range clusterNames {
		clusterClient, err := f.dynamicClientGetter.GetClientForCluster(ctx, clusterName)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		serviceEntryClient := f.serviceEntryClientFactory(clusterClient)
		existingServiceEntries, err := f.fetchExistingServiceEntriesByCluster(clusterName, serviceEntryClient)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		var desiredServiceEntries v1alpha3sets.ServiceEntrySet
		desiredServiceEntriesList, ok := desiredServiceEntriesByCluster[clusterName]
		if !ok {
			desiredServiceEntries = v1alpha3sets.NewServiceEntrySet()
		} else {
			desiredServiceEntries = v1alpha3sets.NewServiceEntrySet(desiredServiceEntriesList...)
		}
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
	ctx context.Context,
	clusterNames []string,
	envoyFilters v1alpha3sets.EnvoyFilterSet,
) error {
	desiredEnvoyFiltersByCluster := map[string][]*istio_client_networking.EnvoyFilter{}
	for _, envoyFilter := range envoyFilters.List() {
		f.addScopeLabels(envoyFilter)
		_, ok := desiredEnvoyFiltersByCluster[envoyFilter.GetClusterName()]
		if !ok {
			desiredEnvoyFiltersByCluster[envoyFilter.GetClusterName()] = []*istio_client_networking.EnvoyFilter{}
		}
		desiredEnvoyFiltersByCluster[envoyFilter.GetClusterName()] = append(desiredEnvoyFiltersByCluster[envoyFilter.GetClusterName()], envoyFilter)
	}
	var multierr *multierror.Error
	// Reconcile per cluster
	for clusterName, desiredEnvoyFilters := range desiredEnvoyFiltersByCluster {
		clusterClient, err := f.dynamicClientGetter.GetClientForCluster(ctx, clusterName)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		envoyFilterClient := f.envoyFilterClientFactory(clusterClient)
		desiredEnvoyFilters := v1alpha3sets.NewEnvoyFilterSet(desiredEnvoyFilters...)
		// Upsert
		for _, desiredEnvoyFilter := range desiredEnvoyFilters.List() {
			err := envoyFilterClient.UpsertEnvoyFilter(f.ctx, desiredEnvoyFilter)
			if err != nil {
				multierr = multierror.Append(multierr, err)
				continue
			}
		}
	}
	// Delete
	for _, clusterName := range clusterNames {
		clusterClient, err := f.dynamicClientGetter.GetClientForCluster(ctx, clusterName)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		envoyFilterClient := f.envoyFilterClientFactory(clusterClient)
		existingEnvoyFilters, err := f.fetchExistingEnvoyFiltersByCluster(clusterName, envoyFilterClient)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		var desiredEnvoyFilters v1alpha3sets.EnvoyFilterSet
		desiredEnvoyFiltersList, ok := desiredEnvoyFiltersByCluster[clusterName]
		if !ok {
			desiredEnvoyFilters = v1alpha3sets.NewEnvoyFilterSet()
		} else {
			desiredEnvoyFilters = v1alpha3sets.NewEnvoyFilterSet(desiredEnvoyFiltersList...)
		}
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

func (f *failoverServiceReconciler) fetchExistingEnvoyFiltersByCluster(
	clusterName string,
	envoyFilterClient v1alpha32.EnvoyFilterClient,
) (v1alpha3sets.EnvoyFilterSet, error) {
	existingEnvoyFilters := v1alpha3sets.NewEnvoyFilterSet()
	existingEnvoyFiltersList, err := envoyFilterClient.ListEnvoyFilter(f.ctx, client.MatchingLabels(FailoverServiceLabels))
	if err != nil {
		return nil, err
	}
	for _, envoyFilter := range existingEnvoyFiltersList.Items {
		envoyFilter := envoyFilter
		envoyFilter.ClusterName = clusterName
		existingEnvoyFilters.Insert(&envoyFilter)
	}
	return existingEnvoyFilters, nil
}

func (f *failoverServiceReconciler) fetchExistingServiceEntriesByCluster(
	clusterName string,
	serviceEntryClient v1alpha32.ServiceEntryClient,
) (v1alpha3sets.ServiceEntrySet, error) {
	existingServiceEntries := v1alpha3sets.NewServiceEntrySet()
	existingEnvoyFiltersList, err := serviceEntryClient.ListServiceEntry(f.ctx, client.MatchingLabels(FailoverServiceLabels))
	if err != nil {
		return nil, err
	}
	for _, serviceEntry := range existingEnvoyFiltersList.Items {
		serviceEntry := serviceEntry
		serviceEntry.ClusterName = clusterName
		existingServiceEntries.Insert(&serviceEntry)
	}
	return existingServiceEntries, nil
}

func (f *failoverServiceReconciler) addScopeLabels(obj v1.Object) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	for k, v := range FailoverServiceLabels {
		labels[k] = v
	}
	obj.SetLabels(labels)
}
