package istio

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kubeinstall"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/util"
)

type Installer interface {
	EnsureIstioInstall(ctx context.Context, mesh *v1.Mesh) error
}

type defaultIstioInstaller struct {
	kubeInstaller kubeinstall.Installer
}

func newIstioInstaller(kubeInstaller kubeinstall.Installer) *defaultIstioInstaller {
	return &defaultIstioInstaller{kubeInstaller: kubeInstaller}
}

func (i *defaultIstioInstaller) EnsureIstioInstall(ctx context.Context, mesh *v1.Mesh) error {
	ctx = contextutils.WithLogger(ctx, "istio-installer")
	logger := contextutils.LoggerFrom(ctx)

	istio := mesh.GetIstio()
	if istio == nil {
		return errors.Errorf("%v: invalid install type, only istio supported currently", mesh.GetMetadata().Ref().Key())
	}

	logger.Infof("syncing istio install %v with config %v", mesh.GetMetadata().Ref().Key(), istio)

	if istio.Install.Options.Disabled {
		logger.Infof("purging resources for disabled install %v", mesh.GetMetadata().Ref().Key())
		if err := i.kubeInstaller.PurgeResources(ctx, util.LabelsForResource(mesh)); err != nil {
			return errors.Wrapf(err, "uninstalling istio")
		}
		return nil
	}

	manifests, err := makeManifestsForInstall(ctx, mesh, istio)
	if err != nil {
		return err
	}

	rawResources, err := manifests.ResourceList()
	if err != nil {
		return err
	}

	installNamespace := istio.Install.Options.InstallationNamespace

	logger.Infof("installing istio with options: %#v", istio)
	if err := i.kubeInstaller.ReconcileResources(ctx, installNamespace, rawResources, util.LabelsForResource(mesh)); err != nil {
		return errors.Wrapf(err, "reconciling install resources failed")
	}

	return nil
}

func makeManifestsForInstall(ctx context.Context, mesh *v1.Mesh, istio *v1.IstioMesh) (helmchart.Manifests, error) {
	if istio.Install.Options.InstallationNamespace == "" {
		return nil, errors.Errorf("must provide installation namespace")
	}

	// self-signed cert is true if a rootcert is not set on either the install or the mesh
	// mesh takes precedence because it has been created from install, the expected place for config
	// to be updated after install
	selfSignedCert := istio.Config.MtlsConfig.RootCertificate == nil
	if mesh != nil {
		selfSignedCert = istio.Config.MtlsConfig != nil && istio.Config.MtlsConfig.RootCertificate == nil
	}
	mtlsOptions := mtlsInstallOptions{
		Enabled: istio.Config.MtlsConfig.MtlsEnabled,
		// self signed cert is true if using the buildtin istio cert
		SelfSignedCert: selfSignedCert,
	}
	autoInjectOptions := autoInjectInstallOptions{
		Enabled: istio.Install.EnableAutoInject,
	}
	observabilityOptions := observabilityInstallOptions{
		EnableGrafana:    istio.Install.Grafana,
		EnablePrometheus: istio.Install.Prometheus,
		EnableJaeger:     istio.Install.Jaeger,
	}

	installVersion, ok := supportedIstioVersions[istio.Install.Options.Version]
	if !ok {
		return nil, errors.Errorf("%v is not a suppported istio version. available: %v", istio.Install.Options.Version, supportedIstioVersions)
	}

	chartParams := helmChartParams{
		valuesTemplate: installVersion.valuesTemplate,
		Mtls:           mtlsOptions,
		AutoInject:     autoInjectOptions,
		Observability:  observabilityOptions,
	}

	helmValues, err := chartParams.helmValues()
	if err != nil {
		return nil, errors.Wrapf(err, "rendering helm values")
	}
	installNamespace := istio.Install.Options.InstallationNamespace

	manifests, err := helmchart.RenderManifests(ctx,
		installVersion.chartPath,
		helmValues,
		"istio", // release name used in some manifests for rendering
		installNamespace,
		"", // use default kube version
	)
	if err != nil {
		return nil, errors.Wrapf(err, "rendering install manifests")
	}

	manifests = append(manifests, installVersion.extraManifests...)

	// based on https://istio.io/blog/2018/soft-multitenancy/#multiple-istio-control-planes
	for i, man := range manifests {
		man.Content = strings.Replace(man.Content, "istio-system", installNamespace, -1)
		manifests[i] = man
	}

	return manifests, nil
}
