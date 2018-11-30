package install

import (
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/pkg/constants"

	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/istio"
	"github.com/solo-io/supergloo/test/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

/*
Smoke test for installing and uninstalling istio.
*/
var _ = Describe("Istio Installer", func() {

	getInstall := func(install bool) *v1.Install {
		installCrd := GetInstallWithoutMeshType(install)
		installCrd.MeshType = &v1.Install_Istio{
			Istio: &v1.Istio{
				InstallationNamespace: InstallNamespace,
			},
		}
		return installCrd
	}

	BeforeEach(func() {
		ChartPath = constants.IstioInstallPath
		randStr := helpers.RandString(8)
		InstallNamespace = "istio-install-test-" + randStr
		MeshName = "istio-mesh-test-" + randStr
	})

	AfterEach(func() {
		// just in case
		util.DeleteCrb(istio.CrbName)
	})

	It("Can install and uninstall istio", func() {
		InstallAndWaitForPods(getInstall(true), 9)
		UninstallAndWaitForCleanup(getInstall(false))

		// Check for non-namespaced resources
		Expect(util.CrbDoesntExist(istio.CrbName)).To(BeTrue())
	})
})
