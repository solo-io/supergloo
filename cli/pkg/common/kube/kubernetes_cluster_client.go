package kube

import (
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/clientset/versioned"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -source ./kubernetes_cluster_client.go -destination ./mocks/kube_mocks.go

// operations on the KubernetesCluster CRD
type KubernetesClusterClient interface {
	Create(cluster *v1alpha1.KubernetesCluster) error
}

func NewKubernetesClusterClient(config *rest.Config) (KubernetesClusterClient, error) {
	clientSet, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &kubernetesClusterClient{clientSet: clientSet}, nil
}

type kubernetesClusterClient struct {
	clientSet *versioned.Clientset
}

func (k *kubernetesClusterClient) Create(cluster *v1alpha1.KubernetesCluster) error {
	_, err := k.clientSet.CoreV1alpha1().KubernetesClusters(cluster.GetNamespace()).Create(cluster)
	if err != nil {
		return err
	}

	return nil
}
