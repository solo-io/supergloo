package kubernetes_cluster

import (
	"context"

	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	rpc_v1 "github.com/solo-io/service-mesh-hub/services/apiserver/pkg/api/v1"
)

func NewKubernetesClusterHandler(
	kubeClusterClient zephyr_discovery.KubernetesClusterClient,
) rpc_v1.KubernetesClusterApiServer {
	return &kubernetesClusterHandler{
		kubeClusterClient: kubeClusterClient,
	}
}

type kubernetesClusterHandler struct {
	kubeClusterClient zephyr_discovery.KubernetesClusterClient
}

func (k *kubernetesClusterHandler) ListClusters(
	ctx context.Context,
	_ *rpc_v1.ListKubernetesClustersRequest,
) (*rpc_v1.ListKubernetesClustersResponse, error) {
	clusters, err := k.kubeClusterClient.List(ctx)
	if err != nil {
		return nil, err
	}
	return &rpc_v1.ListKubernetesClustersResponse{
		Clusters: BuildRpcKubernetesClusterList(clusters),
	}, nil
}

func BuildRpcKubernetesClusterList(cluster *discovery_v1alpha1.KubernetesClusterList) []*rpc_v1.KubernetesCluster {
	result := make([]*rpc_v1.KubernetesCluster, 0, len(cluster.Items))
	for _, v := range cluster.Items {
		result = append(result, BuildRpcKubernetesCluster(&v))
	}
	return result
}

func BuildRpcKubernetesCluster(cluster *discovery_v1alpha1.KubernetesCluster) *rpc_v1.KubernetesCluster {
	return &rpc_v1.KubernetesCluster{
		Spec: &cluster.Spec,
		Ref: &core_types.ResourceRef{
			Name:      cluster.GetName(),
			Namespace: cluster.GetNamespace(),
		},
		Labels: cluster.Labels,
	}
}
