package discovery_core

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/clientset/versioned"
	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/clientset/versioned/typed/discovery.zephyr.solo.io/v1alpha1"
	"k8s.io/client-go/rest"
)

func NewGeneratedKubernetesClusterClient(config *rest.Config) (KubernetesClusterClient, error) {
	clientSet, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &kubernetesClusterClient{clientSet: clientSet.DiscoveryV1alpha1()}, nil
}

type kubernetesClusterClient struct {
	clientSet discoveryv1alpha1.DiscoveryV1alpha1Interface
}

func (k *kubernetesClusterClient) Create(ctx context.Context, cluster *v1alpha1.KubernetesCluster) error {
	_, err := k.clientSet.KubernetesClusters(cluster.GetNamespace()).Create(cluster)
	if err != nil {
		return err
	}

	return nil
}
