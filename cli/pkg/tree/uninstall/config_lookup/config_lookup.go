package config_lookup

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	FailedToFindKubeConfigSecret = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to find kube config secret for cluster %s", clusterName)
	}
	FailedToConvertSecretToKubeConfig = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to convert kube config secret for cluster %s to REST config", clusterName)
	}
	ClusterNotFound = func(clusterName string) error {
		return eris.Errorf("Cluster '%s' was not found", clusterName)
	}
)

func NewKubeConfigLookup(
	kubeClusterClient zephyr_discovery.KubernetesClusterClient,
	secrestClient k8s_core.SecretClient,
	kubeConverter kube.Converter,
) KubeConfigLookup {
	return &kubeConfigLookup{
		kubeClusterClient: kubeClusterClient,
		secretsClient:     secrestClient,
		kubeConverter:     kubeConverter,
	}
}

type kubeConfigLookup struct {
	secretsClient     k8s_core.SecretClient
	kubeClusterClient zephyr_discovery.KubernetesClusterClient
	kubeConverter     kube.Converter
}

func (k *kubeConfigLookup) FromCluster(ctx context.Context, clusterName string) (config *kube.ConvertedConfigs, err error) {
	var kubeCluster *zephyr_discovery.KubernetesCluster
	allClusters, err := k.kubeClusterClient.ListKubernetesCluster(ctx)
	if err != nil {
		return nil, err
	}
	for _, foundCluster := range allClusters.Items {
		if foundCluster.GetName() == clusterName {
			kubeCluster = &foundCluster
			break
		}
	}

	if kubeCluster == nil {
		return nil, ClusterNotFound(clusterName)
	}

	cfgSecretRef := kubeCluster.Spec.GetSecretRef()
	secret, err := k.secretsClient.GetSecret(ctx, client.ObjectKey{Name: cfgSecretRef.GetName(), Namespace: cfgSecretRef.GetNamespace()})
	if err != nil {
		return nil, FailedToFindKubeConfigSecret(err, kubeCluster.GetName())
	}

	clusterName, config, err = k.kubeConverter.SecretToConfig(secret)
	if err != nil {
		return nil, FailedToConvertSecretToKubeConfig(err, kubeCluster.GetName())
	}

	return config, nil
}
