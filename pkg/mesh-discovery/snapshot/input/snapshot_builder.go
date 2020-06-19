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
	Clusters    multicluster.ClusterSet
	Pods        corev1.MulticlusterPodClient
	Services    corev1.MulticlusterServiceClient
	ConfigMaps  corev1.MulticlusterConfigMapClient
	Deployments appsv1.MulticlusterDeploymentClient
}

func (b *builder) BuildSnapshot(ctx context.Context) (Snapshot, error) {
	pods := corev1sets.NewPodSet()
	services := corev1sets.NewServiceSet()
	configMaps := corev1sets.NewConfigMapSet()
	deployments := appsv1sets.NewDeploymentSet()

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

	}

	outputSnap := NewSnapshot(pods, services, configMaps, deployments)

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
