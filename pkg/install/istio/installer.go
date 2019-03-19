package istio

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/go-utils/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type Installer interface {
	EnsureIstioInstall(ctx context.Context, install *v1.Install, meshes v1.MeshList) (*v1.Mesh, error)
}

type defaultIstioInstaller struct {
	helmInstaller helm.Installer
}

func NewDefaultIstioInstaller(helmInstaller helm.Installer) *defaultIstioInstaller {
	return &defaultIstioInstaller{helmInstaller: helmInstaller}
}

func (i *defaultIstioInstaller) EnsureIstioInstall(ctx context.Context, install *v1.Install, meshes v1.MeshList) (*v1.Mesh, error) {
	installMesh, ok := install.InstallType.(*v1.Install_Mesh)
	if !ok {
		return nil, errors.Errorf("%v: invalid install type, must be a mesh", install.Metadata.Ref())
	}

	istio, ok := installMesh.Mesh.InstallType.(*v1.MeshInstall_IstioMesh)
	if !ok {
		return nil, errors.Errorf("%v: invalid install type, only istio supported currently", install.Metadata.Ref())
	}

	ctx = contextutils.WithLogger(ctx, "istio-installer")
	logger := contextutils.LoggerFrom(ctx)

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
			if err := i.helmInstaller.DeleteFromManifests(ctx, installNamespace, previousInstall); err != nil {
				return nil, errors.Wrapf(err, "uninstalling istio")
			}
			install.InstalledManifest = ""
			installMesh.Mesh.InstalledMesh = nil
		}
		return nil, nil
	}

	var mesh *v1.Mesh
	if installMesh.Mesh.InstalledMesh != nil {
		var err error
		mesh, err = meshes.Find(installMesh.Mesh.InstalledMesh.Strings())
		if err != nil {
			return nil, errors.Wrapf(err, "installed mesh not found")
		}
	}

	// self-signed cert is true if a rootcert is not set on either the install or the mesh
	// mesh takes precedence because it may be updated by the user
	selfSignedCert := istio.IstioMesh.CustomRootCert == nil
	if mesh != nil && mesh.MtlsConfig != nil {
		selfSignedCert = mesh.MtlsConfig.RootCertificate == nil
	}
	mtlsOptions := mtlsInstallOptions{
		Enabled: istio.IstioMesh.EnableMtls,
		// self signed cert is true if using the buildtin istio cert
		SelfSignedCert: selfSignedCert,
	}
	autoInjectOptions := autoInjectInstallOptions{
		Enabled: istio.IstioMesh.EnableAutoInject,
	}
	observabilityOptions := observabilityInstallOptions{
		EnableGrafana:    istio.IstioMesh.InstallGrafana,
		EnablePrometheus: istio.IstioMesh.InstallPrometheus,
		EnableJaeger:     istio.IstioMesh.InstallJaeger,
	}

	opts := NewInstallOptions(
		previousInstall,
		i.helmInstaller,
		istio.IstioMesh.IstioVersion,
		installNamespace,
		autoInjectOptions,
		mtlsOptions,
		observabilityOptions,
		gatewayInstallOptions{},
	)

	logger.Infof("installing istio with options: %#v", opts)

	filterFunc := helm.ReplaceHardcodedNamespace("istio-system", installNamespace)
	manifests, err := helm.InstallOrUpdate(ctx, opts, filterFunc)
	if err != nil {
		return nil, errors.Wrapf(err, "installing istio")
	}

	gzipped, err := manifests.Gzipped()
	if err != nil {
		return nil, errors.Wrapf(err, "converting installed mannifests to gzipped string")
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
				MtlsEnabled:     istio.IstioMesh.EnableMtls,
				RootCertificate: istio.IstioMesh.CustomRootCert,
			},
		}
	}

	// caller should expect the install to have been modified
	install.InstalledManifest = gzipped
	ref := mesh.Metadata.Ref()
	installMesh.Mesh.InstalledMesh = &ref

	return mesh, nil
}
