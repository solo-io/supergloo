package mesh

import (
	"context"

	"github.com/solo-io/supergloo/pkg/install/mesh/istio"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/go-utils/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type Installer interface {
	EnsureMeshInstall(ctx context.Context, install *v1.Install) (*v1.Mesh, error)
}

type defaultInstaller struct {
	helmInstaller helm.Installer
}

func NewDefaultInstaller(helmInstaller helm.Installer) *defaultInstaller {
	return &defaultInstaller{helmInstaller: helmInstaller}
}

func (installer *defaultInstaller) EnsureMeshInstall(ctx context.Context, install *v1.Install) (*v1.Mesh, error) {
	ctx = contextutils.WithLogger(ctx, "istio-installer")
	logger := contextutils.LoggerFrom(ctx)
	installMesh, ok := install.InstallType.(*v1.Install_Mesh)
	if !ok {
		return nil, errors.Errorf("non mesh install detected in mesh install, %v", install.Metadata.Ref())
	}

	istioMesh, ok := installMesh.Mesh.InstallType.(*v1.MeshInstall_IstioMesh)
	if !ok {
		return nil, errors.Errorf("%v: invalid install type, only istio supported currently", install.Metadata.Ref())
	}

	var previousInstall helm.Manifests
	if install.InstalledManifest != "" {
		logger.Infof("detected previous install of istio")
		manifests, err := helm.NewManifestsFromGzippedString(install.InstalledManifest)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing previously installed manifest")
		}
		previousInstall = manifests
	}

	installNamespace := install.InstallationNamespace

	if install.Disabled {
		if len(previousInstall) > 0 {
			logger.Infof("deleting previous istio install")
			if err := installer.helmInstaller.DeleteFromManifests(ctx, installNamespace, previousInstall); err != nil {
				return nil, errors.Wrapf(err, "uninstalling istio")
			}
			install.InstalledManifest = ""
			installMesh.Mesh.InstalledMesh = nil
		}
		return nil, nil
	}

	version := istioMesh.IstioMesh.IstioVersion
	autoInjectOptions := istio.AutoInjectInstallOptions{
		Enabled: istioMesh.IstioMesh.EnableAutoInject,
	}
	mtlsOptions := istio.MtlsInstallOptions{
		Enabled: istioMesh.IstioMesh.EnableMtls,
		// self signed cert is true if using the buildtin istio cert
		SelfSignedCert: istioMesh.IstioMesh.CustomRootCert == nil,
	}
	observabilityOptions := istio.ObservabilityInstallOptions{
		EnableGrafana:    istioMesh.IstioMesh.InstallGrafana,
		EnablePrometheus: istioMesh.IstioMesh.InstallPrometheus,
		EnableJaeger:     istioMesh.IstioMesh.InstallJaeger,
	}

	opts := istio.NewInstallOptions(previousInstall,
		installer.helmInstaller,
		version,
		installNamespace,
		autoInjectOptions,
		mtlsOptions,
		observabilityOptions,
		istio.GatewayInstallOptions{},
	)

	logger.Infof("installing istio with options: %#v", opts)

	manifests, err := helm.InstallOrUpdate(ctx, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "installing istio")
	}

	gzipped, err := manifests.Gzipped()
	if err != nil {
		return nil, errors.Wrapf(err, "converting installed mannifests to gzipped string")
	}

	mesh := &v1.Mesh{
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
			MtlsEnabled: istioMesh.IstioMesh.EnableMtls,
		},
	}

	// caller should expect the install to have been modified
	install.InstalledManifest = gzipped
	ref := mesh.Metadata.Ref()
	installMesh.Mesh.InstalledMesh = &ref

	return mesh, nil
}
