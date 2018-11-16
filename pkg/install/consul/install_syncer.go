package consul

import (
	"context"
	"strconv"
	"strings"

	"github.com/solo-io/supergloo/pkg/api/v1"
	helmlib "k8s.io/helm/pkg/helm"

	"github.com/solo-io/supergloo/pkg/install/helm"
)

type ConsulInstallSyncer struct{}

func (c *ConsulInstallSyncer) Sync(_ context.Context, snap *v1.InstallSnapshot) error {

	// See hack/install/consul/install-on... for steps
	// 1. Create a namespace / project for consul (this step should go away and we use provided namespace)
	// 2. Set up ClusterRoleBinding for consul in that namespace (not currently done)
	// 3. Install Consul via helm chart
	// 4. Fix incorrect configuration -> webhook service looking for wrong adapter config name (this should move into an override)

	for _, install := range snap.Installs.List() {
		if install.Consul != nil {

			// TODO: Allow specifying namespace in the proto
			namespace := "consul"
			// TODO: Create namespace

			updatedOverrides := overridesYaml
			if install.Encryption != nil {
				strBool := strconv.FormatBool(install.Encryption.TlsEnabled)
				updatedOverrides = strings.Replace(overridesYaml, "@@MTLS_ENABLED@@", strBool, -1)
			}

			// helm install
			helmClient, err := helm.GetHelmClient()
			if err != nil {
				helm.Teardown() // just in case
				return err
			}

			if err != nil {
				helm.Teardown() // just in case
				return err
			}

			_, err = helmClient.InstallRelease(
				install.Consul.Path,
				namespace,
				helmlib.ValueOverrides([]byte(updatedOverrides)))
			helm.Teardown()
			if err != nil {
				return err
			}

			// TODO: Fix incorrect webhook configuration

			// Create Mesh CRD
		}
	}
	return nil
}

var overridesYaml = `
global:
  # Change this to specify a version of consul.
  # soloio/consul:latest was just published to provide a 1.4 container
  # consul:1.3.0 is the latest container on docker hub from consul
  image: "soloio/consul:latest"
  imageK8S: "hashicorp/consul-k8s:0.2.1"

server:
  replicas: 1
  bootstrapExpect: 1
  connect: @@MTLS_ENABLED@@
  disruptionBudget:
    enabled: false
    maxUnavailable: null

connectInject:
  enabled: @@MTLS_ENABLED@@
`
