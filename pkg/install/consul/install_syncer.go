package consul

import (
	"context"

	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/solo-io/supergloo/pkg/api/v1"

	"github.com/solo-io/supergloo/pkg/install/helm"
)

type ConsulInstallSyncer struct{}

func (c *ConsulInstallSyncer) Sync(_ context.Context, snap *v1.InstallSnapshot) error {
	for _, install := range snap.Installs.List() {
		if install.Consul != nil {
			// helm install
			helmClient, err := helm.GetHelmClient()
			if err != nil {
				helm.Teardown() // just in case
				return err
			}

			consulChart := chart.Chart{}
			namespace := "consul"

			helmClient.InstallReleaseFromChart(&consulChart, namespace)

			helm.Teardown()
		}
	}
	return nil
}
