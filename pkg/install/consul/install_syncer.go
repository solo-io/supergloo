package consul

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/solo-io/supergloo/pkg/install"

	"k8s.io/api/admissionregistration/v1beta1"

	"github.com/pkg/errors"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/supergloo/pkg/api/v1"
)

const (
	CrbName          = "consul-crb"
	defaultNamespace = "consul"
	WebhookCfg       = "consul-connect-injector-cfg"
)

func SyncInstall(_ context.Context, install *v1.Install, syncer *install.InstallSyncer) error {
	if install.MeshType != v1.MeshType_CONSUL {
		return errors.Errorf("Expected mesh type consul")
	}

	// 1. Create a namespace
	installNamespace, err := syncer.SetupInstallNamespace(install, defaultNamespace)
	if err != nil {
		return err
	}

	// 2. Set up ClusterRoleBinding for consul in that namespace
	// This is not cleaned up when deleting namespace so it may already exist on the system, don't fail
	err = syncer.CreateCrbIfNotExist(CrbName, installNamespace)
	if err != nil {
		return err
	}

	// 3. Install Consul via helm chart
	releaseName, err := syncer.HelmInstall(install.ChartLocator, installNamespace, getOverrides(install.Encryption))
	if err != nil {
		// TODO: Wrap this one level deeper
		return errors.Wrap(err, "Error installing Consul helm chart")
	}

	// 4. If mtls enabled, fix incorrect configuration name in chart
	if install.Encryption.TlsEnabled {
		err = updateMutatingWebhookAdapter(syncer.Kube, releaseName)
		if err != nil {
			return errors.Wrap(err, "Error setting up webhook")
		}
	}

	return nil
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
func updateMutatingWebhookAdapter(kube *kubernetes.Clientset, releaseName string) error {
	name := fmt.Sprintf("%s-%s", releaseName, WebhookCfg)
	cfg, err := kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Get(name, kubemeta.GetOptions{})
	if err != nil {
		return err
	}
	fixedCfg := getFixedWebhookAdapter(cfg)
	_, err = kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(fixedCfg)
	if err != nil {
		return err
	}
	err = kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(name, &kubemeta.DeleteOptions{})
	return err
}

func getFixedWebhookAdapter(input *v1beta1.MutatingWebhookConfiguration) *v1beta1.MutatingWebhookConfiguration {
	fixed := input.DeepCopy()
	fixed.Name = WebhookCfg
	fixed.ResourceVersion = ""
	return fixed
}
