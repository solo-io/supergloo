package linkerd2

import (
	"github.com/solo-io/supergloo/pkg/api/v1"
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

func (c *Linkerd2Installer) DoPreHelmInstall(installNamespace string, install *v1.Install) error {
	return nil
}
