package install

import (
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/pkg/constants"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

/*
Smoke test for installing and uninstalling linkerd2
*/
var _ = Describe("Linkerd2 Installer", func() {

	getInstall := func(install bool) *v1.Install {
		installCrd := GetInstallWithoutMeshType(install)
		installCrd.MeshType = &v1.Install_Linkerd2{
			Linkerd2: &v1.Linkerd2{
				InstallationNamespace: InstallNamespace,
			},
		}
		return installCrd
	}

	BeforeEach(func() {
		ChartPath = constants.LinkerdInstallPath
		randStr := helpers.RandString(8)
		InstallNamespace = "linkerd-install-test-" + randStr
		MeshName = "linkerd-mesh-test-" + randStr
	})

	It("Can install and uninstall linkerd2", func() {
		InstallAndWaitForPods(getInstall(true), 4)
		UninstallAndWaitForCleanup(getInstall(false))
	})
})
