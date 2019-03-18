package istio

import (
	"github.com/pkg/errors"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"
)

var supportedIstioVersions = map[string]versionedInstall{
	IstioVersion103: {
		chartPath:      IstioVersion103Chart,
		valuesTemplate: helmValues,
	},
	IstioVersion105: {
		chartPath:      IstioVersion105Chart,
		valuesTemplate: helmValues,
	},
	IstioVersion106: {
		chartPath:      IstioVersion106Chart,
		valuesTemplate: helmValues,
	},
}

type versionedInstall struct {
	chartPath      string
	valuesTemplate string
}

type installOptions struct {
	// if set, this is an upgrade from a previous install
	previousInstall helm.Manifests
	installer       helm.Installer

	version       string
	namespace     string
	AutoInject    AutoInjectInstallOptions
	Mtls          MtlsInstallOptions
	Observability ObservabilityInstallOptions
	Gateway       GatewayInstallOptions
}

func NewInstallOptions(previousInstall helm.Manifests, installer helm.Installer, version string, namespace string,
	autoInject AutoInjectInstallOptions, mtls MtlsInstallOptions, observability ObservabilityInstallOptions,
	gateway GatewayInstallOptions) *installOptions {
	return &installOptions{previousInstall: previousInstall, installer: installer, version: version,
		namespace: namespace, AutoInject: autoInject, Mtls: mtls, Observability: observability, Gateway: gateway}
}

func (o installOptions) Type() string {
	return "istio"
}

func (o installOptions) Uri() string {
	return supportedIstioVersions[o.Version()].chartPath
}

func (o installOptions) Version() string {
	return o.version
}

func (o installOptions) Namespace() string {
	return o.namespace
}

func (o installOptions) NamespaceOverride() string {
	return "istio-system"
}

func (o installOptions) Validate() error {
	return o.validate()
}

func (o installOptions) PreviousInstall() helm.Manifests {
	return o.previousInstall
}

func (o installOptions) Installer() helm.Installer {
	return o.installer
}

func (o installOptions) HelmValuesTemplate() string {
	return supportedIstioVersions[o.Version()].valuesTemplate
}

func (o installOptions) validate() error {
	if o.Version() == "" {
		return errors.Errorf("must provide istio install version")
	} else {
		_, ok := supportedIstioVersions[o.Version()]
		if !ok {
			return errors.Errorf("%v is not a supported istio version. available versions and their chart locations: %v", o.Version(), supportedIstioVersions)
		}
	}
	if o.Namespace() == "" {
		return errors.Errorf("must provide istio install namespace")
	}
	if o.Observability.EnableServiceGraph && !o.Observability.EnablePrometheus {
		return errors.Errorf("servicegraph can only be enabled with prometheus")
	}
	return nil
}

type AutoInjectInstallOptions struct {
	Enabled bool
}

type MtlsInstallOptions struct {
	Enabled        bool
	SelfSignedCert bool
}

type ObservabilityInstallOptions struct {
	EnableGrafana      bool
	EnablePrometheus   bool
	EnableJaeger       bool
	EnableServiceGraph bool
}

type GatewayInstallOptions struct {
	EnableIngress bool
	EnableEgress  bool
}
