package meshingress

import (
	"fmt"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"
)

type installOptions struct {
	previousInstall helm.Manifests
	installer       helm.Installer
	namespace       string
	version         string
	uri             string
}

func (o installOptions) NamespaceOverride() string {
	return "gloo-system"
}

func newInstallOptions(previousInstall helm.Manifests, installer helm.Installer, namespace string, version string) *installOptions {
	uri := glooManifestUrl(version)
	return &installOptions{previousInstall: previousInstall, installer: installer, namespace: namespace, version: version, uri: uri}
}

func (o installOptions) Type() string {
	return "gloo-ingress"
}

func (o installOptions) Installer() helm.Installer {
	return o.installer
}

func (o installOptions) HelmValuesTemplate() string {
	return helmValues
}

func (o installOptions) Uri() string {
	return o.uri
}

func (o installOptions) Version() string {
	return o.version
}

func (o installOptions) Namespace() string {
	return o.namespace
}

func (o installOptions) Validate() error {
	return o.validate()
}

func (o installOptions) PreviousInstall() helm.Manifests {
	return o.previousInstall
}

func (o installOptions) validate() error {
	if o.Version() == "" {
		return errors.Errorf("must provide gloo-ingress install version")
	}
	if o.Namespace() == "" {
		return errors.Errorf("must provide gloo-ingress install namespace")
	}
	return nil
}

func glooManifestUrl(version string) string {
	var url = "https://storage.googleapis.com/solo-public-helm/charts/gloo-%s.tgz"
	return fmt.Sprintf(url, version)
}
