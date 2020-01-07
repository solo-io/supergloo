package setup

import (
	"context"
	"os"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/authorization/v1alpha1"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/customresourcedefinition"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/deployment"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/clientfactory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	multicluster2 "github.com/solo-io/solo-kit/pkg/multicluster"

	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
)

// TODO revisit whether CRD creation should be skipped

func InitializeUpstreamClient(ctx context.Context, sharedCacheGetter clustercache.CacheGetter, watchHandler multicluster.ClientForClusterHandler) (gloov1.UpstreamClient, handler.ClusterHandler) {
	upstreamClientFactory := clientfactory.NewKubeResourceClientFactory(sharedCacheGetter,
		gloov1.UpstreamCrd,
		false,
		nil,
		0,
		factory.NewResourceClientParams{ResourceType: &gloov1.Upstream{}})
	upstreamClientGetter := multicluster.NewClusterClientManager(ctx, upstreamClientFactory, watchHandler)
	upstreamBaseClient := multicluster.NewMultiClusterResourceClient(&gloov1.Upstream{}, upstreamClientGetter)
	return gloov1.NewUpstreamClientWithBase(upstreamBaseClient), upstreamClientGetter
}

func InitializeDeploymentClient(ctx context.Context, deploymentCacheGetter clustercache.CacheGetter, watchHandler multicluster.ClientForClusterHandler) (kubernetes.DeploymentClient, handler.ClusterHandler) {
	deploymentClientFactory := deployment.NewDeploymentResourceClientFactory(deploymentCacheGetter)
	deploymentClientGetter := multicluster.NewClusterClientManager(ctx, deploymentClientFactory, watchHandler)
	deploymentBaseClient := multicluster.NewMultiClusterResourceClient(&kubernetes.Deployment{}, deploymentClientGetter)
	return kubernetes.NewDeploymentClientWithBase(deploymentBaseClient), deploymentClientGetter
}

func InitializePodClient(ctx context.Context, podCacheGetter clustercache.CacheGetter, watchHandler multicluster.ClientForClusterHandler) (kubernetes.PodClient, handler.ClusterHandler) {
	podClientFactory := pod.NewPodResourceClientFactory(podCacheGetter)
	podClientGetter := multicluster.NewClusterClientManager(ctx, podClientFactory, watchHandler)
	podBaseClient := multicluster.NewMultiClusterResourceClient(&kubernetes.Pod{}, podClientGetter)
	return kubernetes.NewPodClientWithBase(podBaseClient), podClientGetter
}

func InitializeMeshPolicyClient(ctx context.Context, sharedCacheGetter clustercache.CacheGetter) (v1alpha1.MeshPolicyClient, handler.ClusterHandler) {
	meshPolicyClientFactory := clientfactory.NewKubeResourceClientFactory(sharedCacheGetter,
		v1alpha1.MeshPolicyCrd,
		false,
		nil,
		0,
		factory.NewResourceClientParams{ResourceType: &v1alpha1.MeshPolicy{}})
	meshPolicyClientGetter := multicluster.NewClusterClientManager(ctx, meshPolicyClientFactory)
	meshPolicyBaseClient := multicluster.NewMultiClusterResourceClient(&v1alpha1.MeshPolicy{}, meshPolicyClientGetter)
	return v1alpha1.NewMeshPolicyClientWithBase(meshPolicyBaseClient), meshPolicyClientGetter
}

func InitializeCustomResourceDefinitionClient(ctx context.Context, coreCacheGetter clustercache.CacheGetter) (kubernetes.CustomResourceDefinitionClient, handler.ClusterHandler) {
	crdClientFactory := customresourcedefinition.NewCrdResourceClientFactory(coreCacheGetter)
	crdClientGetter := multicluster.NewClusterClientManager(ctx, crdClientFactory)
	crdBaseClient := multicluster.NewMultiClusterResourceClient(&kubernetes.CustomResourceDefinition{}, crdClientGetter)
	return kubernetes.NewCustomResourceDefinitionClientWithBase(crdBaseClient), crdClientGetter
}

func InitializeMeshClient(ctx context.Context, sharedCacheGetter clustercache.CacheGetter) (v1.MeshClient, error) {
	cfg, err := kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic("KUBECONFIG is not defined")
	}
	meshClientFactory := factory.KubeResourceClientFactory{
		Crd: v1.MeshCrd,
		Cfg: cfg,
		// TODO panic if the cache is wrong type
		SharedCache:     sharedCacheGetter.GetCache(multicluster2.LocalCluster, cfg).(kube.SharedCache),
		SkipCrdCreation: false,
	}

	client, err := v1.NewMeshClient(&meshClientFactory)
	if err != nil {
		return nil, err
	}
	if err := client.Register(); err != nil {
		return nil, err
	}
	return client, nil
}

func InitializeMeshIngressClient(ctx context.Context, sharedCacheGetter clustercache.CacheGetter) (v1.MeshIngressClient, error) {
	cfg, err := kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic("KUBECONFIG is not defined")
	}
	meshIngressClientFactory := factory.KubeResourceClientFactory{
		Crd: v1.MeshIngressCrd,
		Cfg: cfg,
		// TODO panic if the cache is wrong type
		SharedCache:     sharedCacheGetter.GetCache(multicluster2.LocalCluster, cfg).(kube.SharedCache),
		SkipCrdCreation: false,
	}

	client, err := v1.NewMeshIngressClient(&meshIngressClientFactory)
	if err != nil {
		return nil, err
	}
	if err := client.Register(); err != nil {
		return nil, err
	}
	return client, nil
}
