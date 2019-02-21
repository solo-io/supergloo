package istio

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/go-utils/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type istioInstaller struct{}

// installs istio, returns a mesh object created for the install,
// and updates the install itself with the inline manifest
func (i *istioInstaller) InstallIstio(ctx context.Context, install *v1.Install) (*v1.Mesh, error) {
	istio, ok := install.InstallType.(*v1.Install_Istio_)
	if !ok {
		return nil, errors.Errorf("%v: invalid mesh type, istioInstaller only supports istio", install.Metadata.Ref())
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

	installNamespace := istio.Istio.InstallationNamespace

	if install.Disabled {
		if len(previousInstall) > 0 {
			logger.Infof("deleting previous istio install")
			if err := helm.DeleteFromManifests(ctx, installNamespace, previousInstall); err != nil {
				return nil, errors.Wrapf(err, "uninstalling istio")
			}
			install.InstalledManifest = ""
		}
		return nil, nil
	}

	opts := installOptions{
		previousInstall: previousInstall,
		Version:         istio.Istio.IstioVersion,
		Namespace:       installNamespace,
		AutoInject: autoInjectInstallOptions{
			Enabled: istio.Istio.EnableAutoInject,
		},
		Mtls: mtlsInstallOptions{
			Enabled: istio.Istio.EnableMtls,
		},
		Observability: observabilityInstallOptions{
			EnableGrafana:    istio.Istio.InstallGrafana,
			EnablePrometheus: istio.Istio.InstallPrometheus,
			EnableJaeger:     istio.Istio.InstallJaeger,
		},
	}
	logger.Infof("installing istio with options: %#v", opts)

	manifests, err := installIstio(ctx, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "installing istio")
	}

	gzipped, err := manifests.Gzipped()
	if err != nil {
		return nil, errors.Wrapf(err, "converting installed mannifests to gzipped string")
	}

	// caller should expect the install to have been modified
	install.InstalledManifest = gzipped

	mesh := &v1.Mesh{
		Metadata: core.Metadata{
			Namespace: install.Metadata.Namespace,
			Name:      install.Metadata.Name,
		},
		MeshType: &v1.Mesh_Istio_{
			Istio: &v1.Mesh_Istio{
				// TODO
			},
		},
	}

	return mesh, nil
}
