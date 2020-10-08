package utils

import (
	"github.com/solo-io/service-mesh-hub/pkg/common/schemes"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildClient(kubeConfigPath, kubeContext string) (client.Client, error) {
	cfg, err := kubeconfig.GetRestConfigWithContext(kubeConfigPath, kubeContext, "")
	if err != nil {
		return nil, err
	}

	scheme := scheme.Scheme
	if err := schemes.AddToScheme(scheme); err != nil {
		return nil, err
	}

	client, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}
