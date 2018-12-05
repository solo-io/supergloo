package install

import (
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/pkg/constants"
	"github.com/solo-io/supergloo/test/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/consul"
)

/*
Smoke test for installing and uninstalling consul
*/
var _ = Describe("Consul Installer", func() {

	getInstall := func(install bool) *v1.Install {
		installCrd := GetInstallWithoutMeshType(install)
		installCrd.MeshType = &v1.Install_Consul{
			Consul: &v1.Consul{
				InstallationNamespace: InstallNamespace,
			},
		}
		return installCrd
	}

	BeforeEach(func() {
		ChartPath = constants.ConsulInstallPath
		randStr := helpers.RandString(8)
		InstallNamespace = "consul-install-test-" + randStr
		MeshName = "consul-mesh-test-" + randStr
	})

	AfterEach(func() {
		// Just in case
		util.DeleteWebhookConfigIfExists(consul.WebhookCfg)
		util.DeleteCrb(consul.CrbName)
	})

	It("Can install and uninstall consul", func() {
		InstallAndWaitForPods(getInstall(true), 2)
		UninstallAndWaitForCleanup(getInstall(false))

		// Check for non-namespaced resources
		Expect(util.CrbDoesntExist(consul.CrbName)).To(BeTrue())
		Expect(util.WebhookConfigNotFound(consul.WebhookCfg)).To(BeTrue())
	})
})
