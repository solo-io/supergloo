package istio

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/utils/helmchart"
	"github.com/solo-io/supergloo/pkg/install/utils/kubeinstall"
	"github.com/solo-io/supergloo/pkg/util"
)

type Installer interface {
	EnsureIstioInstall(ctx context.Context, install *v1.Install, meshes v1.MeshList) (*v1.Mesh, error)
}

type defaultIstioInstaller struct {
	kubeInstaller kubeinstall.Installer
}

func newIstioInstaller(kubeInstaller kubeinstall.Installer) *defaultIstioInstaller {
	return &defaultIstioInstaller{kubeInstaller: kubeInstaller}
}

func (i *defaultIstioInstaller) EnsureIstioInstall(ctx context.Context, install *v1.Install, meshes v1.MeshList) (*v1.Mesh, error) {
	ctx = contextutils.WithLogger(ctx, "istio-installer")
	logger := contextutils.LoggerFrom(ctx)
	installMesh := install.GetMesh()
	if installMesh == nil {
		return nil, errors.Errorf("%v: invalid install type, must be a mesh", install.Metadata.Ref())
	}

	istio := installMesh.GetIstioMesh()
	if istio == nil {
		return nil, errors.Errorf("%v: invalid install type, only istio supported currently", install.Metadata.Ref())
	}

	logger.Infof("syncing istio install %v with config %v", install.Metadata.Ref().Key(), istio)

	if install.Disabled {
		logger.Infof("purging resources for disabled install %v", install.Metadata.Ref())
		if err := i.kubeInstaller.PurgeResources(ctx, util.LabelsForResource(install)); err != nil {
			return nil, errors.Wrapf(err, "uninstalling istio")
		}
		installMesh.InstalledMesh = nil
		return nil, nil
	}

	var mesh *v1.Mesh
	if installMesh.InstalledMesh != nil {
		var err error
		mesh, err = meshes.Find(installMesh.InstalledMesh.Strings())
		if err != nil {
			return nil, errors.Wrapf(err, "installed mesh not found")
		}
	}

	manifests, err := makeManifestsForInstall(ctx, install, mesh, istio)
	if err != nil {
		return nil, err
	}

	rawResources, err := manifests.ResourceList()
	if err != nil {
		return nil, err
	}

	installNamespace := install.InstallationNamespace

	logger.Infof("installing istio with options: %#v", istio)
	if err := i.kubeInstaller.ReconcilleResources(ctx, installNamespace, rawResources, util.LabelsForResource(install)); err != nil {
		return nil, errors.Wrapf(err, "reconciling install resources failed")
	}

	if mesh != nil {
		mesh.MeshType = &v1.Mesh_Istio{
			Istio: &v1.IstioMesh{
				InstallationNamespace: installNamespace,
			},
		}
	} else {
		mesh = &v1.Mesh{
			Metadata: core.Metadata{
				Namespace: install.Metadata.Namespace,
				Name:      install.Metadata.Name,
			},
			MeshType: &v1.Mesh_Istio{
				Istio: &v1.IstioMesh{
					InstallationNamespace: installNamespace,
				},
			},
			MtlsConfig: &v1.MtlsConfig{
				MtlsEnabled:     istio.EnableMtls,
				RootCertificate: istio.CustomRootCert,
			},
		}
	}

	// caller should expect the install to have been modified
	ref := mesh.Metadata.Ref()
	installMesh.InstalledMesh = &ref

	return mesh, nil
}

func makeManifestsForInstall(ctx context.Context, install *v1.Install, mesh *v1.Mesh, istio *v1.IstioInstall) (helmchart.Manifests, error) {
	if install.InstallationNamespace == "" {
		return nil, errors.Errorf("must provide installation namespace")
	}

	// self-signed cert is true if a rootcert is not set on either the install or the mesh
	// mesh takes precedence because it has been created from install, the expected place for config
	// to be updated after install
	selfSignedCert := istio.CustomRootCert == nil
	if mesh != nil {
		selfSignedCert = mesh.MtlsConfig != nil && mesh.MtlsConfig.RootCertificate == nil
	}
	mtlsOptions := mtlsInstallOptions{
		Enabled: istio.EnableMtls,
		// self signed cert is true if using the buildtin istio cert
		SelfSignedCert: selfSignedCert,
	}
	autoInjectOptions := autoInjectInstallOptions{
		Enabled: istio.EnableAutoInject,
	}
	observabilityOptions := observabilityInstallOptions{
		EnableGrafana:    istio.InstallGrafana,
		EnablePrometheus: istio.InstallPrometheus,
		EnableJaeger:     istio.InstallJaeger,
	}

	installVersion, ok := supportedIstioVersions[istio.IstioVersion]
	if !ok {
		return nil, errors.Errorf("%v is not a suppported istio version. available: %v", istio.IstioVersion, supportedIstioVersions)
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
	installNamespace := install.InstallationNamespace

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

	// based on https://istio.io/blog/2018/soft-multitenancy/#multiple-istio-control-planes
	for i, man := range manifests {
		man.Content = strings.Replace(man.Content, "istio-system", installNamespace, -1)
		manifests[i] = man
	}

	return manifests, nil
}
