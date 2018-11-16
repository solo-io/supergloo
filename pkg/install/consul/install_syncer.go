package consul

import (
	"context"

	"github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/helm/pkg/helm"
)

type ConsulInstallSyncer struct{}

func (c *ConsulInstallSyncer) Sync(_ context.Context, snap *v1.InstallSnapshot) error {
	for _, install := range snap.Installs.List() {
		if install.Consul != nil {
			// helm install
			GetHelmClient()

		}
	}
	return nil
}

func GetHelmClient() (*helm.Client, error) {
	helm := helm.NewClient()
	if err := helm.PingTiller(); err != nil {
		return nil, err
	}
	return helm, nil
}
