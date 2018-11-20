package linkerd2

import (
	"k8s.io/client-go/kubernetes"

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

func (c *Linkerd2Installer) DoPreHelmInstall() error {
	return nil
}

func (c *Linkerd2Installer) DoPostHelmInstall(install *v1.Install, kube *kubernetes.Clientset, releaseName string) error {
	return nil
}
