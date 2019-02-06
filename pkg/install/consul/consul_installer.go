package consul

import (
	"context"
	"strconv"
	"strings"

	istiov1 "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

const (
	CrbName          = "consul-crb"
	defaultNamespace = "consul"
	WebhookCfg       = "connect-injector-cfg"
)

type ConsulInstaller struct{}

func (c *ConsulInstaller) GetDefaultNamespace() string {
	return defaultNamespace
}

func (c *ConsulInstaller) GetCrbName() string {
	return CrbName
}

func (c *ConsulInstaller) GetOverridesYaml(install *v1.Install) string {
	return getOverrides(install.Encryption)
}

func (c *ConsulInstaller) DoPreHelmInstall(ctx context.Context, installNamespace string, install *v1.Install, secretList istiov1.IstioCacertsSecretList) error {
	return nil
}

func getOverrides(encryption *v1.Encryption) string {
	strBool := "false"
	if encryption != nil {
		strBool = strconv.FormatBool(encryption.TlsEnabled)
	}
	return strings.Replace(overridesYaml, "@@MTLS_ENABLED@@", strBool, -1)
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

client:
  enabled: true
  grpc: true

connectInject:
  enabled: @@MTLS_ENABLED@@
`
