package istio

import (
	"github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/client-go/kubernetes"

	security "github.com/openshift/client-go/security/clientset/versioned"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CrbName          = "istio-crb"
	defaultNamespace = "istio-system"
)

type IstioInstaller struct {
	SecurityClient *security.Clientset
}

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

func (c *IstioInstaller) DoPreHelmInstall() error {
	if c.SecurityClient == nil {
		return nil
	}
	return c.AddSccToUsers(
		"default",
		"istio-ingress-service-account",
		"prometheus",
		"istio-egressgateway-service-account",
		"istio-citadel-service-account",
		"istio-ingressgateway-service-account",
		"istio-cleanup-old-ca-service-account",
		"istio-mixer-post-install-account",
		"istio-mixer-service-account",
		"istio-pilot-service-account",
		"istio-sidecar-injector-service-account",
		"istio-galley-service-account")
}

// TODO: something like this should enable minishift installs to succeed, but this isn't right. The correct steps are
//       to run "oc adm policy add-scc-to-user anyuid -z %s -n istio-system" for each of the user accounts above
//       maybe the issue is not specifying the namespace?
func (c *IstioInstaller) AddSccToUsers(users ...string) error {
	anyuid, err := c.SecurityClient.SecurityV1().SecurityContextConstraints().Get("anyuid", kubemeta.GetOptions{})
	if err != nil {
		return err
	}
	newUsers := append(anyuid.Users, users...)
	anyuid.Users = newUsers
	_, err = c.SecurityClient.SecurityV1().SecurityContextConstraints().Update(anyuid)
	return err
}
