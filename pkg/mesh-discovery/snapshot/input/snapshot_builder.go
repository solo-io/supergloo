package input

import (
	"context"

	"github.com/hashicorp/go-multierror"
	appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	corev1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/skv2/pkg/multicluster"
)

// builds the input snapshot from API Clients
type Builder interface {
	BuildSnapshot(ctx context.Context) (Snapshot, error)
}

type builder struct {
	Clusters     multicluster.ClusterSet
	Pods         corev1.MulticlusterPodClient
	Services     corev1.MulticlusterServiceClient
	ConfigMaps   corev1.MulticlusterConfigMapClient
	Deployments  appsv1.MulticlusterDeploymentClient
	ReplicaSets  appsv1.MulticlusterReplicaSetClient
	DaemonSets   appsv1.MulticlusterDaemonSetClient
	StatefulSets appsv1.MulticlusterStatefulSetClient
}

func NewBuilder(
	clusters multicluster.ClusterSet,
	client multicluster.Client,
) *builder {
	return &builder{
		Clusters:     clusters,
		Pods:         corev1.NewMulticlusterPodClient(client),
		Services:     corev1.NewMulticlusterServiceClient(client),
		ConfigMaps:   corev1.NewMulticlusterConfigMapClient(client),
		Deployments:  appsv1.NewMulticlusterDeploymentClient(client),
		ReplicaSets:  appsv1.NewMulticlusterReplicaSetClient(client),
		DaemonSets:   appsv1.NewMulticlusterDaemonSetClient(client),
		StatefulSets: appsv1.NewMulticlusterStatefulSetClient(client),
	}
}

func (b *builder) BuildSnapshot(ctx context.Context) (Snapshot, error) {
	pods := corev1sets.NewPodSet()
	services := corev1sets.NewServiceSet()
	configMaps := corev1sets.NewConfigMapSet()
	deployments := appsv1sets.NewDeploymentSet()
	replicaSets := appsv1sets.NewReplicaSetSet()
	daemonSets := appsv1sets.NewDaemonSetSet()
	statefulSets := appsv1sets.NewStatefulSetSet()

	var errs error

	for _, cluster := range b.Clusters.ListClusters() {

		if err := b.insertPodsFromCluster(ctx, cluster, pods); err != nil {
			errs = multierror.Append(errs, err)
		}

		if err := b.insertServicesFromCluster(ctx, cluster, services); err != nil {
			errs = multierror.Append(errs, err)
		}

		if err := b.insertConfigMapsFromCluster(ctx, cluster, configMaps); err != nil {
			errs = multierror.Append(errs, err)
		}

		if err := b.insertDeploymentsFromCluster(ctx, cluster, deployments); err != nil {
			errs = multierror.Append(errs, err)
		}

		if err := b.insertReplicaSetsFromCluster(ctx, cluster, replicaSets); err != nil {
			errs = multierror.Append(errs, err)
		}

		if err := b.insertDaemonSetsFromCluster(ctx, cluster, daemonSets); err != nil {
			errs = multierror.Append(errs, err)
		}

		if err := b.insertStatefulSetsFromCluster(ctx, cluster, statefulSets); err != nil {
			errs = multierror.Append(errs, err)
		}

	}

	outputSnap := NewSnapshot(
		pods,
		services,
		configMaps,
		deployments,
		replicaSets,
		daemonSets,
		statefulSets,
	)

	return outputSnap, errs
}

func (b *builder) insertPodsFromCluster(ctx context.Context, cluster string, pods corev1sets.PodSet) error {
	podClient, err := b.Pods.Cluster(cluster)
	if err != nil {
		return err
	}

	podList, err := podClient.ListPod(ctx)
	if err != nil {
		return err
	}

	for _, item := range podList.Items {
		item := item               // pike
		item.ClusterName = cluster // set cluster for in-memory processing
		pods.Insert(&item)
	}

	return nil
}

func (b *builder) insertServicesFromCluster(ctx context.Context, cluster string, services corev1sets.ServiceSet) error {
	serviceClient, err := b.Services.Cluster(cluster)
	if err != nil {
		return err
	}

	serviceList, err := serviceClient.ListService(ctx)
	if err != nil {
		return err
	}

	for _, item := range serviceList.Items {
		item := item               // pike
		item.ClusterName = cluster // set cluster for in-memory processing
		services.Insert(&item)
	}

	return nil
}

func (b *builder) insertConfigMapsFromCluster(ctx context.Context, cluster string, configMaps corev1sets.ConfigMapSet) error {
	configMapClient, err := b.ConfigMaps.Cluster(cluster)
	if err != nil {
		return err
	}

	configMapList, err := configMapClient.ListConfigMap(ctx)
	if err != nil {
		return err
	}

	for _, item := range configMapList.Items {
		item := item               // pike
		item.ClusterName = cluster // set cluster for in-memory processing
		configMaps.Insert(&item)
	}

	return nil
}

func (b *builder) insertDeploymentsFromCluster(ctx context.Context, cluster string, deployments appsv1sets.DeploymentSet) error {
	deploymentClient, err := b.Deployments.Cluster(cluster)
	if err != nil {
		return err
	}

	deploymentList, err := deploymentClient.ListDeployment(ctx)
	if err != nil {
		return err
	}

	for _, item := range deploymentList.Items {
		item := item               // pike
		item.ClusterName = cluster // set cluster for in-memory processing
		deployments.Insert(&item)
	}

	return nil
}

func (b *builder) insertReplicaSetsFromCluster(ctx context.Context, cluster string, replicaSets appsv1sets.ReplicaSetSet) error {
	replicaSetClient, err := b.ReplicaSets.Cluster(cluster)
	if err != nil {
		return err
	}

	replicaSetList, err := replicaSetClient.ListReplicaSet(ctx)
	if err != nil {
		return err
	}

	for _, item := range replicaSetList.Items {
		item := item               // pike
		item.ClusterName = cluster // set cluster for in-memory processing
		replicaSets.Insert(&item)
	}

	return nil
}

func (b *builder) insertDaemonSetsFromCluster(ctx context.Context, cluster string, daemonSets appsv1sets.DaemonSetSet) error {
	daemonSetClient, err := b.DaemonSets.Cluster(cluster)
	if err != nil {
		return err
	}

	daemonSetList, err := daemonSetClient.ListDaemonSet(ctx)
	if err != nil {
		return err
	}

	for _, item := range daemonSetList.Items {
		item := item               // pike
		item.ClusterName = cluster // set cluster for in-memory processing
		daemonSets.Insert(&item)
	}

	return nil
}

func (b *builder) insertStatefulSetsFromCluster(ctx context.Context, cluster string, statefulSets appsv1sets.StatefulSetSet) error {
	statefulSetClient, err := b.StatefulSets.Cluster(cluster)
	if err != nil {
		return err
	}

	statefulSetList, err := statefulSetClient.ListStatefulSet(ctx)
	if err != nil {
		return err
	}

	for _, item := range statefulSetList.Items {
		item := item               // pike
		item.ClusterName = cluster // set cluster for in-memory processing
		statefulSets.Insert(&item)
	}

	return nil
}
