package install

import (
	"context"
	"os"
	"testing"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/supergloo/pkg2/api/v1"
	"github.com/solo-io/supergloo/pkg2/constants"

	"github.com/solo-io/supergloo/pkg2/install"
	"github.com/solo-io/supergloo/test/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

/*
Set environment variable HELM_CHART_PATH to override the default helm chart. This applies to
all tests that run, so focus a test if you are testing a specific new chart.
*/

func TestInstallers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Installers e2e Suite")
}

var providedChartPath string
var CreatedSuperglooNamespace bool

var _ = BeforeSuite(func() {
	providedChartPath = os.Getenv("HELM_CHART_PATH")
	CreatedSuperglooNamespace = util.TryCreateNamespace(constants.SuperglooNamespace)
})

var _ = AfterSuite(func() {
	if CreatedSuperglooNamespace {
		util.TerminateNamespaceBlocking(constants.SuperglooNamespace)
	}
})

var Syncer *install.InstallSyncer

// Get set in before each of test files
var MeshName string
var ChartPath string
var InstallNamespace string

var _ = BeforeEach(func() {
	var err error
	Syncer, err = install.NewKubeInstallSyncer(util.GetMeshClient(), util.GetSecretClient(), util.GetKubeClient(), util.GetApiExtsClient())
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	util.UninstallHelmRelease(MeshName)
	util.TerminateNamespaceBlocking(InstallNamespace)
})

func GetInstallWithoutMeshType(install bool) *v1.Install {
	path := providedChartPath
	if path == "" {
		path = ChartPath
	}
	return util.GetInstallWithoutMeshType(path, MeshName, install)
}

func InstallAndWaitForPods(install *v1.Install, pods int) {
	snap := util.GetSnapshot(install)
	err := Syncer.Sync(context.TODO(), snap)
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() int {
		return util.WaitForAvailablePods(InstallNamespace)
	}).Should(BeEquivalentTo(pods))
}

func UninstallAndWaitForCleanup(install *v1.Install) {
	snap := util.GetSnapshot(install)
	err := Syncer.Sync(context.TODO(), snap)
	Expect(err).NotTo(HaveOccurred())

	// Validate everything cleaned up
	util.WaitForTerminatedNamespace(InstallNamespace)
	Expect(util.HelmReleaseDoesntExist(MeshName)).To(BeTrue())

	mesh, err := util.GetMeshClient().Read(constants.SuperglooNamespace, MeshName, clients.ReadOpts{})
	Expect(mesh).To(BeNil())
	Expect(err).ToNot(BeNil())
}
