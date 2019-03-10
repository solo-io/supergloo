package istio

import (
	"bytes"
	"context"
	"strings"
	"text/template"

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

	Version       string
	Namespace     string
	AutoInject    autoInjectInstallOptions
	Mtls          mtlsInstallOptions
	Observability observabilityInstallOptions
	Gateway       gatewayInstallOptions
}

func (o installOptions) validate() error {
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

type autoInjectInstallOptions struct {
	Enabled bool
}

type mtlsInstallOptions struct {
	Enabled        bool
	SelfSignedCert bool
}

type observabilityInstallOptions struct {
	EnableGrafana      bool
	EnablePrometheus   bool
	EnableJaeger       bool
	EnableServiceGraph bool
}

type gatewayInstallOptions struct {
	EnableIngress bool
	EnableEgress  bool
}

func releaseName(namespace, version string) string {
	return "supergloo-" + namespace + version
}

// returns the installed manifests
func (i *defaultIstioInstaller) installOrUpdateIstio(ctx context.Context, opts installOptions) (helm.Manifests, error) {
	if err := opts.validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid install options")
	}
	version := opts.Version
	namespace := opts.Namespace
	installInfo, ok := supportedIstioVersions[version]
	if !ok {
		return nil, errors.Errorf("%v is not a supported istio version. available versions and their chart locations: %v", version, supportedIstioVersions)
	}

	helmValueOverrides, err := template.New("istio-" + version).Parse(installInfo.valuesTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "")
	}

	valuesBuf := &bytes.Buffer{}
	if err := helmValueOverrides.Execute(valuesBuf, opts); err != nil {
		return nil, errors.Wrapf(err, "internal error: rendering helm values")
	}

	manifests, err := helm.RenderManifests(
		ctx,
		installInfo.chartPath,
		valuesBuf.String(),
		releaseName(namespace, version),
		namespace,
		"", // NOTE(ilackarms): use helm default
		true,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "rendering manifests")
	}

	for i, m := range manifests {
		// replace all instances of istio-system with the target namespace
		// based on instructions at https://istio.io/blog/2018/soft-multitenancy/#multiple-istio-control-planes
		m.Content = strings.Replace(m.Content, "istio-system", namespace, -1)
		manifests[i] = m
	}

	// nothing to do if the manifest hasn't changed
	if opts.previousInstall.CombinedString() == manifests.CombinedString() {
		return manifests, nil
	}

	// perform upgrade instead
	if len(opts.previousInstall) > 0 {
		if err := i.helmInstaller.UpdateFromManifests(ctx, namespace, opts.previousInstall, manifests, true); err != nil {
			return nil, errors.Wrapf(err, "creating istio from manifests")
		}
	} else {
		if err := i.helmInstaller.CreateFromManifests(ctx, namespace, manifests); err != nil {
			return nil, errors.Wrapf(err, "creating istio from manifests")
		}
	}

	return manifests, nil
}
