package istio

import (
	"context"

	"github.com/pkg/errors"
	"github.com/solo-io/supergloo/pkg/install/helm"
)

const (
	IstioVersion103      = "1.0.3"
	IstioVersion103Chart = "https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.3.tgz"

	IstioVersion105      = "1.0.5"
	IstioVersion105Chart = "https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.5.tgz"
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
}

type versionedInstall struct {
	chartPath      string
	valuesTemplate string
}

type InstallOptions struct {
	Version       string
	Namespace     string
	AutoInject    AutoInjectInstallOptions
	Mtls          MtlsInstallOptions
	Observability ObservabilityInstallOptions
	Gateway       GatewayInstallOptions
}

func (o InstallOptions) Validate() error {
	if o.Version == "" {
		return errors.Errorf("must provide istio install version")
	}
	if o.Namespace == "" {
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

func releaseName(namespace, version string) string {
	return "supergloo-" + namespace + version
}

func InstallIstio(ctx context.Context, opts InstallOptions) error {
	if err := opts.Validate(); err != nil {
		return errors.Wrapf(err, "invalid install options")
	}
	version := opts.Version
	namespace := opts.Namespace
	installInfo, ok := supportedIstioVersions[version]
	if !ok {
		return errors.Errorf("%v is not a supported istio version. available versions and their chart locations: %v", version, supportedIstioVersions)
	}

	return helm.Install(
		ctx,
		releaseName(namespace, version),
		namespace,
		installInfo.chartPath,
		installInfo.valuesTemplate,
		opts,
	)
}
