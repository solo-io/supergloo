package consul

import (
	"context"

	"github.com/solo-io/supergloo/pkg/api/v1"

	"github.com/solo-io/supergloo/pkg/install/helm"
)

type ConsulInstallSyncer struct{}

func (c *ConsulInstallSyncer) Sync(_ context.Context, snap *v1.InstallSnapshot) error {
	for _, install := range snap.Installs.List() {
		if install.Consul != nil {
			// helm install
			helm.GetHelmClient()
			helm.Teardown()
		}
	}
	return nil
}
