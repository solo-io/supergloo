package consul

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"k8s.io/api/admissionregistration/v1beta1"

	"github.com/pkg/errors"
	kubecore "k8s.io/api/core/v1"
	kuberbac "k8s.io/api/rbac/v1"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/supergloo/pkg/api/v1"
	helmlib "k8s.io/helm/pkg/helm"

	"github.com/solo-io/supergloo/pkg/install/helm"
)

const (
	CrbName          = "consul-crb"
	defaultNamespace = "consul"
	WebhookCfg       = "consul-connect-injector-cfg"
)

type ConsulInstallSyncer struct {
	Kube       *kubernetes.Clientset
	MeshClient v1.MeshClient
}

func (c *ConsulInstallSyncer) Sync(ctx context.Context, snap *v1.InstallSnapshot) error {
	for _, install := range snap.Installs.List() {
		err := c.SyncInstall(ctx, install)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ConsulInstallSyncer) SyncInstall(_ context.Context, install *v1.Install) error {
	if install.Consul == nil {
		return nil
	}

	// 1. Create a namespace
	installNamespace := getInstallNamespace(install.Consul)
	err := c.createNamespaceIfNotExist(installNamespace) // extract to CRD
	if err != nil {
		return errors.Wrap(err, "Error setting up namespace")
	}

	// 2. Set up ClusterRoleBinding for consul in that namespace
	// This is not cleaned up when deleting namespace so it may already exist on the system, don't fail
	err = c.createCrbIfNotExist(installNamespace)
	if err != nil {
		return errors.Wrap(err, "Error setting up CRB")
	}

	// 3. Install Consul via helm chart
	releaseName, err := helmInstall(install.Encryption, install.Consul, installNamespace)
	if err != nil {
		return errors.Wrap(err, "Error installing Consul helm chart")
	}

	// 4. If mtls enabled, fix incorrect configuration name in chart
	if install.Encryption.TlsEnabled {
		err = c.updateMutatingWebhookAdapter(releaseName)
		if err != nil {
			return errors.Wrap(err, "Error setting up webhook")
		}
	}

	return c.createMesh(install)
}

func getInstallNamespace(consul *v1.ConsulInstall) string {
	installNamespace := defaultNamespace
	if consul.Namespace != "" {
		installNamespace = consul.Namespace
	}
	return installNamespace
}

func (c *ConsulInstallSyncer) createNamespaceIfNotExist(namespaceName string) error {
	_, err := c.Kube.CoreV1().Namespaces().Get(namespaceName, kubemeta.GetOptions{})
	if err == nil {
		// Namespace already exists
		return nil
	}
	_, err = c.Kube.CoreV1().Namespaces().Create(getNamespace(namespaceName))
	return err
}

func getNamespace(namespaceName string) *kubecore.Namespace {
	return &kubecore.Namespace{
		ObjectMeta: kubemeta.ObjectMeta{
			Name: namespaceName,
		},
	}
}

func (c *ConsulInstallSyncer) createCrbIfNotExist(namespaceName string) error {
	_, err := c.Kube.RbacV1().ClusterRoleBindings().Get(CrbName, kubemeta.GetOptions{})
	if err == nil {
		// crb already exists
		return nil
	}
	_, err = c.Kube.RbacV1().ClusterRoleBindings().Create(getCrb(namespaceName))
	return err
}

func getCrb(namespaceName string) *kuberbac.ClusterRoleBinding {
	meta := kubemeta.ObjectMeta{
		Name: "consul-crb",
	}
	subject := kuberbac.Subject{
		Kind:      "ServiceAccount",
		Namespace: namespaceName,
		Name:      "default",
	}
	roleRef := kuberbac.RoleRef{
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
		APIGroup: "rbac.authorization.k8s.io",
	}
	return &kuberbac.ClusterRoleBinding{
		ObjectMeta: meta,
		Subjects:   []kuberbac.Subject{subject},
		RoleRef:    roleRef,
	}
}

func helmInstall(encryption *v1.Encryption, consul *v1.ConsulInstall, installNamespace string) (string, error) {
	overrides := []byte(getOverrides(encryption))
	// helm install
	helmClient, err := helm.GetHelmClient()
	if err != nil {
		return "", err
	}

	installPath, err := helm.LocateChartPathDefault(consul.Path)
	if err != nil {
		return "", err
	}
	response, err := helmClient.InstallRelease(
		installPath,
		installNamespace,
		helmlib.ValueOverrides(overrides))
	helm.Teardown()
	if err != nil {
		return "", err
	} else {
		return response.Release.Name, nil
	}
}

func getOverrides(encryption *v1.Encryption) string {
	updatedOverrides := overridesYaml
	if encryption != nil {
		strBool := strconv.FormatBool(encryption.TlsEnabled)
		updatedOverrides = strings.Replace(overridesYaml, "@@MTLS_ENABLED@@", strBool, -1)
	}
	return updatedOverrides
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

// The webhook config is created with the wrong name by the chart
// Grab it, recreate with correct name, and delete the old one
func (c *ConsulInstallSyncer) updateMutatingWebhookAdapter(releaseName string) error {
	name := fmt.Sprintf("%s-%s", releaseName, WebhookCfg)
	cfg, err := c.Kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Get(name, kubemeta.GetOptions{})
	if err != nil {
		return err
	}
	fixedCfg := getFixedWebhookAdapter(cfg)
	_, err = c.Kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(fixedCfg)
	if err != nil {
		return err
	}
	err = c.Kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(name, &kubemeta.DeleteOptions{})
	return err
}

func getFixedWebhookAdapter(input *v1beta1.MutatingWebhookConfiguration) *v1beta1.MutatingWebhookConfiguration {
	fixed := input.DeepCopy()
	fixed.Name = WebhookCfg
	fixed.ResourceVersion = ""
	return fixed
}

func (c *ConsulInstallSyncer) createMesh(install *v1.Install) error {
	mesh := getMeshObject(install)
	_, err := c.MeshClient.Write(mesh, clients.WriteOpts{})
	return err
}

func getMeshObject(install *v1.Install) *v1.Mesh {
	return &v1.Mesh{
		Metadata: core.Metadata{
			Name:      install.Metadata.Name,
			Namespace: install.Metadata.Namespace,
		},
		TargetMesh: &v1.TargetMesh{
			MeshType: v1.MeshType_CONSUL,
		},
		Encryption: install.Encryption,
	}
}
