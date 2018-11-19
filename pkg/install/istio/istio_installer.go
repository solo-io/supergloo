package istio

import (
	"github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	CrbName          = "istio-crb"
	defaultNamespace = "istio-system"
)

type IstioInstaller struct{}

func (c *IstioInstaller) GetDefaultNamespace() string {
	return defaultNamespace
}

func (c *IstioInstaller) GetCrbName() string {
	return CrbName
}

func (c *IstioInstaller) GetOverridesYaml(install *v1.Install) string {
	return ""
}

func (c *IstioInstaller) DoPostHelmInstall(install *v1.Install, kube *kubernetes.Clientset, releaseName string) error {
	return nil
}
