package istio

import (
	"github.com/solo-io/supergloo/pkg/install/common"

	"github.com/solo-io/go-utils/installutils/kubeinstall"

	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

// calling this function with nil is valid and expected outside of tests
func NewInstallSyncer(kubeInstaller kubeinstall.Installer, meshClient v1.MeshClient, reporter reporter.Reporter) v1.InstallSyncer {
	istioInstaller := newIstioInstaller(kubeInstaller)
	return common.NewMeshInstallSyncer("istio", meshClient, reporter, isIstioInstall, istioInstaller.EnsureIstioInstall)
}

func isIstioInstall(mesh *v1.Mesh) *v1.InstallOptions {
	istio := mesh.GetIstio()
	if istio == nil {
		return nil
	}
	return istio.Install.Options
}
