package consul

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/pkg/install/helm"
)

// weird note with consul
// the version of the consul chart we use is 0.6.0
// the version of consul that chart deploys is 1.4.2

// TODO (ilackarms): move these hard-coded values to crds
const (
	ConsulVersion060      = "0.6.0"
	ConsulVersion060Chart = "https://storage.googleapis.com/supergloo-charts/consul-0.6.0.tar.gz"
)

var supportedConsulVersions = map[string]versionedInstall{
	ConsulVersion060: {
		chartPath:      ConsulVersion060Chart,
		valuesTemplate: helmValues,
	},
}

type versionedInstall struct {
	chartPath      string
	valuesTemplate string
}

type InstallOptions struct {
	Version    string
	Namespace  string
	NodeCount  int
	AutoInject AutoInjectInstallOptions
}

type AutoInjectInstallOptions struct {
	Enabled bool
}

func (o InstallOptions) Validate() error {
	if o.Version == "" {
		return errors.Errorf("must provide istio install version")
	}
	if o.Namespace == "" {
		return errors.Errorf("must provide istio install namespace")
	}
	return nil
}

func releaseName(namespace, version string) string {
	version = strings.Replace(version, ".", "", -1)
	return "supergloo-" + namespace + version
}

func InstallConsul(ctx context.Context, opts InstallOptions) error {
	if err := opts.Validate(); err != nil {
		return errors.Wrapf(err, "invalid install options")
	}
	version := opts.Version
	namespace := opts.Namespace
	installInfo, ok := supportedConsulVersions[version]
	if !ok {
		return errors.Errorf("%v is not a supported istio version. available versions and their chart locations: %v", version, supportedConsulVersions)
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
