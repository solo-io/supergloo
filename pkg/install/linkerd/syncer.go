package linkerd

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/pkg/install/common"
	"github.com/solo-io/supergloo/pkg/util"

	"github.com/solo-io/go-utils/installutils/kubeinstall"

	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

// calling this function with nil is valid and expected outside of tests
func NewInstallSyncer(kubeInstaller kubeinstall.Installer, meshClient v1.MeshClient, reporter reporter.Reporter) v1.InstallSyncer {
	return common.NewMeshInstallSyncer("linkerd", meshClient, reporter, isLinkerdInstall, linkerdInstaller{kubeInstaller}.ensureLinkerdInstall)
}

func isLinkerdInstall(install *v1.Install) bool {
	mesh := install.GetMesh()
	if mesh == nil {
		return false
	}
	return mesh.GetLinkerdMesh() != nil
}

type linkerdInstaller struct {
	kubeInstaller kubeinstall.Installer
}

func (i linkerdInstaller) ensureLinkerdInstall(ctx context.Context, install *v1.Install, meshes v1.MeshList) (*v1.Mesh, error) {
	installMesh := install.GetMesh()
	if installMesh == nil {
		return nil, errors.Errorf("%v: invalid install type, must be a mesh", install.Metadata.Ref())
	}

	linkerd := installMesh.GetLinkerdMesh()
	if linkerd == nil {
		return nil, errors.Errorf("%v: invalid install type, only linkerd supported currently", install.Metadata.Ref())
	}

	logger := contextutils.LoggerFrom(ctx)

	logger.Infof("syncing linkerd install %v with config %v", install.Metadata.Ref().Key(), linkerd)

	installLabels := util.LabelsForResource(install)

	if install.Disabled {
		logger.Infof("purging resources for disabled install %v", install.Metadata.Ref())
		if err := i.kubeInstaller.PurgeResources(ctx, installLabels); err != nil {
			return nil, errors.Wrapf(err, "uninstalling linkerd")
		}
		installMesh.InstalledMesh = nil
		return nil, nil
	}

	opts := newInstallOpts(linkerd.LinkerdVersion, install.InstallationNamespace, linkerd.EnableMtls, linkerd.EnableAutoInject)

	if err := opts.install(ctx, i.kubeInstaller, installLabels); err != nil {
		return nil, errors.Wrapf(err, "executing linkerd install")
	}

	return createOrUpdateMesh(install, installMesh, linkerd, meshes)
}

func createOrUpdateMesh(install *v1.Install, installMesh *v1.MeshInstall, linkerd *v1.LinkerdInstall, meshes v1.MeshList) (*v1.Mesh, error) {
	var mesh *v1.Mesh
	if installMesh.InstalledMesh != nil {
		var err error
		mesh, err = meshes.Find(installMesh.InstalledMesh.Strings())
		if err != nil {
			return nil, errors.Wrapf(err, "installed mesh not found")
		}
	}

	if mesh != nil {
		mesh.MeshType = &v1.Mesh_LinkerdMesh{
			LinkerdMesh: &v1.LinkerdMesh{
				InstallationNamespace: install.InstallationNamespace,
			},
		}
		return mesh, nil
	}

	return &v1.Mesh{
		Metadata: core.Metadata{
			Namespace: install.Metadata.Namespace,
			Name:      install.Metadata.Name,
		},
		MeshType: &v1.Mesh_LinkerdMesh{
			LinkerdMesh: &v1.LinkerdMesh{
				InstallationNamespace: install.InstallationNamespace,
			},
		},
		MtlsConfig: &v1.MtlsConfig{
			MtlsEnabled: linkerd.EnableMtls,
		},
	}, nil
}
