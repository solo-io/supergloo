package linkerd

import (
	"context"

	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/pkg/install/common"
	"github.com/solo-io/supergloo/pkg/util"

	"github.com/solo-io/go-utils/installutils/kubeinstall"

	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

// calling this function with nil is valid and expected outside of tests
func NewInstallSyncer(kubeInstaller kubeinstall.Installer, kubeClient kubernetes.Interface, meshClient v1.MeshClient, reporter reporter.Reporter) v1.InstallSyncer {
	return common.NewMeshInstallSyncer("linkerd", meshClient, reporter, isLinkerdInstall, linkerdInstaller{kubeInstaller: kubeInstaller, kubeClient: kubeClient}.ensureLinkerdInstall)
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
	kubeClient    kubernetes.Interface
}

func (i linkerdInstaller) ensureLinkerdInstall(ctx context.Context, install *v1.Install, meshes v1.MeshList) error {
	installMesh := install.GetMesh()
	if installMesh == nil {
		return errors.Errorf("%v: invalid install type, must be a mesh", install.Metadata.Ref())
	}

	linkerd := installMesh.GetLinkerdMesh()
	if linkerd == nil {
		return errors.Errorf("%v: invalid install type, only linkerd supported currently", install.Metadata.Ref())
	}

	logger := contextutils.LoggerFrom(ctx)

	logger.Infof("syncing linkerd install %v with config %v", install.Metadata.Ref().Key(), linkerd)

	installLabels := util.LabelsForResource(install)

	if install.Disabled {
		logger.Infof("purging resources for disabled install %v", install.Metadata.Ref())
		if err := i.kubeInstaller.PurgeResources(ctx, installLabels); err != nil {
			return errors.Wrapf(err, "uninstalling linkerd")
		}
		return nil
	}

	opts := newInstallOpts(linkerd.LinkerdVersion, install.InstallationNamespace, linkerd.EnableMtls, linkerd.EnableAutoInject)

	if err := opts.install(ctx, i.kubeInstaller, installLabels, i.kubeClient); err != nil {
		return errors.Wrapf(err, "executing linkerd install")
	}

	return nil
}
