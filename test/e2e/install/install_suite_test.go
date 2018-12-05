package install

import (
	"context"
	"os"
	"testing"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/constants"

	"github.com/gogo/protobuf/types"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install"
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

var KubeCache *kube.KubeCache
var providedChartPath string
var CreatedSuperglooNamespace bool

var _ = BeforeSuite(func() {
	providedChartPath = os.Getenv("HELM_CHART_PATH")
	CreatedSuperglooNamespace = util.TryCreateNamespace(constants.SuperglooNamespace)
	KubeCache = kube.NewKubeCache()
})

var _ = AfterSuite(func() {
	if CreatedSuperglooNamespace {
		util.TerminateNamespaceBlocking(constants.SuperglooNamespace)
	}
})

var Syncer install.InstallSyncer

// Get set in before each of test files
var MeshName string
var ChartPath string
var InstallNamespace string

var _ = BeforeEach(func() {
	Syncer = install.InstallSyncer{
		Kube:       util.GetKubeClient(),
		MeshClient: util.GetMeshClient(KubeCache),
		ApiExts:    util.GetApiExtsClient(),
	}
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
	return &v1.Install{
		Metadata: core.Metadata{
			Namespace: constants.SuperglooNamespace,
			Name:      MeshName,
		},
		ChartLocator: &v1.HelmChartLocator{
			Kind: &v1.HelmChartLocator_ChartPath{
				ChartPath: &v1.HelmChartPath{
					Path: path,
				},
			},
		},
		Enabled: &types.BoolValue{
			Value: install,
		},
	}
}

func getSnapshot(install *v1.Install) *v1.InstallSnapshot {
	return &v1.InstallSnapshot{
		Installs: v1.InstallsByNamespace{
			constants.SuperglooNamespace: v1.InstallList{
				install,
			},
		},
	}
}

func InstallAndWaitForPods(install *v1.Install, pods int) {
	snap := getSnapshot(install)
	err := Syncer.Sync(context.TODO(), snap)
	Expect(err).NotTo(HaveOccurred())
	Expect(util.WaitForAvailablePods(InstallNamespace)).To(BeEquivalentTo(pods))
}

func UninstallAndWaitForCleanup(install *v1.Install) {
	snap := getSnapshot(install)
	err := Syncer.Sync(context.TODO(), snap)
	Expect(err).NotTo(HaveOccurred())

	// Validate everything cleaned up
	util.WaitForTerminatedNamespace(InstallNamespace)
	Expect(util.HelmReleaseDoesntExist(MeshName)).To(BeTrue())

	mesh, err := util.GetMeshClient(KubeCache).Read(constants.SuperglooNamespace, MeshName, clients.ReadOpts{})
	Expect(mesh).To(BeNil())
	Expect(err).ToNot(BeNil())
}
