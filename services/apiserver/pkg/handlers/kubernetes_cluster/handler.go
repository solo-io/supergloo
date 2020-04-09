package kubernetes_cluster

import (
	"context"

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
	context.Context,
	*rpc_v1.ListKubernetesClustersRequest,
) (*rpc_v1.ListKubernetesClustersResponse, error) {
	panic("implement me")
}
