package linkerd2

import (
	"context"

	istiov1 "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

const (
	defaultNamespace = "linkerd"
)

type Linkerd2Installer struct{}

func (c *Linkerd2Installer) GetDefaultNamespace() string {
	return defaultNamespace
}

func (c *Linkerd2Installer) GetCrbName() string {
	return ""
}

func (c *Linkerd2Installer) GetOverridesYaml(install *v1.Install) string {
	return ""
}

func (c *Linkerd2Installer) DoPreHelmInstall(ctx context.Context, installNamespace string, install *v1.Install, secretList istiov1.IstioCacertsSecretList) error {
	return nil
}
