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

func isLinkerdInstall(mesh *v1.Mesh) *v1.InstallOptions {
	linkerd := mesh.GetLinkerd()
	if linkerd == nil {
		return nil
	}
	return linkerd.Install.Options
}

type linkerdInstaller struct {
	kubeInstaller kubeinstall.Installer
	kubeClient    kubernetes.Interface
}

func (i linkerdInstaller) ensureLinkerdInstall(ctx context.Context, mesh *v1.Mesh) error {
	linkerd := mesh.GetLinkerd()
	if linkerd == nil {
		return errors.Errorf("%v: invalid install type, only linkerd supported currently", mesh.Metadata.Ref().Key())
	}

	logger := contextutils.LoggerFrom(ctx)

	logger.Infof("syncing linkerd install %v with config %v", mesh.Metadata.Ref().Key(), linkerd)

	installLabels := util.LabelsForResource(mesh)

	if linkerd.Install.Options.Disabled {
		logger.Infof("purging resources for disabled install %v", mesh.Metadata.Ref().Key())
		if err := i.kubeInstaller.PurgeResources(ctx, installLabels); err != nil {
			return errors.Wrapf(err, "uninstalling linkerd")
		}
		return nil
	}

	opts := newInstallOpts(linkerd.Install.Options.Version, linkerd.Install.Options.InstallationNamespace,
		linkerd.Config.MtlsConfig.MtlsEnabled, linkerd.Install.EnableAutoInject)

	if err := opts.install(ctx, i.kubeInstaller, installLabels, i.kubeClient); err != nil {
		return errors.Wrapf(err, "executing linkerd install")
	}

	return nil
}
