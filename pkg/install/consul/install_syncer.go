package consul

import (
	"context"
	"strconv"
	"strings"

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
	crbName          = "consul-crb"
	defaultNamespace = "consul"
)

type ConsulInstallSyncer struct {
	Kube *kubernetes.Clientset
}

func (c *ConsulInstallSyncer) Sync(ctx context.Context, snap *v1.InstallSnapshot) error {
	for _, install := range snap.Installs.List() {
		c.SyncInstall(ctx, install)
	}
	return nil
}

func (c *ConsulInstallSyncer) SyncInstall(_ context.Context, install *v1.Install) error {
	// See hack/install/consul/install-on... for steps
	// 1. Create a namespace / project for consul (this step should go away and we use provided namespace)
	// 2. Set up ClusterRoleBinding for consul in that namespace (not currently done)
	// 3. Install Consul via helm chart
	// 4. Fix incorrect configuration -> webhook service looking for wrong adapter config name (this should move into an override)
	if install.Consul == nil {
		return nil
	}

	installNamespace := getInstallNamespace(install.Consul)

	err := c.createNamespaceIfNotExist(installNamespace) // extract to CRD
	if err != nil {
		return errors.Wrap(err, "Error setting up namespace")
	}

	err = c.createCrbIfNotExist(installNamespace)
	if err != nil {
		return errors.Wrap(err, "Error setting up CRB")
	}

	err = helmInstall(install.Encryption, install.Consul, installNamespace)
	if err != nil {
		return errors.Wrap(err, "Error installing Consul helm chart")
	}

	// TODO: Fix incorrect webhook configuration
	// TODO: Create Mesh CRD

	return nil
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
	_, err := c.Kube.RbacV1().ClusterRoleBindings().Get(crbName, kubemeta.GetOptions{})
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

func helmInstall(encryption *v1.Encryption, consul *v1.ConsulInstall, installNamespace string) error {
	overrides := []byte(getOverrides(encryption))
	// helm install
	helmClient, err := helm.GetHelmClient()
	if err != nil {
		return err
	}

	_, err = helmClient.InstallRelease(
		consul.Path,
		installNamespace,
		helmlib.ValueOverrides(overrides))
	helm.Teardown()
	return err
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
