package install_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	mock_types "github.com/solo-io/go-utils/installutils/helminstall/types/mocks"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	cli_mocks "github.com/solo-io/mesh-projects/cli/pkg/mocks"
	cli_test "github.com/solo-io/mesh-projects/cli/pkg/test"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/install"
	"k8s.io/client-go/rest"
)

var _ = Describe("Install", func() {
	var (
		ctrl              *gomock.Controller
		ctx               context.Context
		mockKubeLoader    *cli_mocks.MockKubeLoader
		meshctl           *cli_test.MockMeshctl
		mockHelmInstaller *mock_types.MockInstaller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockHelmInstaller = mock_types.NewMockInstaller(ctrl)
		mockKubeLoader = cli_mocks.NewMockKubeLoader(ctrl)
		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Clients:        common.Clients{},
			KubeClients:    common.KubeClients{HelmInstaller: mockHelmInstaller},
			KubeLoader:     mockKubeLoader,
			Ctx:            ctx,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should set default values for flags", func() {
		chartOverride := "chartOverride.tgz"
		installerconfig := &types.InstallerConfig{
			DryRun:             false,
			CreateNamespace:    true,
			Verbose:            false,
			InstallNamespace:   "service-mesh-hub",
			ReleaseName:        cliconstants.ServiceMeshHubReleaseName,
			ReleaseUri:         chartOverride,
			ValuesFiles:        []string{},
			PreInstallMessage:  install.PreInstallMessage,
			PostInstallMessage: install.PostInstallMessage,
		}
		mockKubeLoader.
			EXPECT().
			GetRestConfigForContext("", "").
			Return(&rest.Config{}, nil)
		mockHelmInstaller.
			EXPECT().
			Install(installerconfig).
			Return(nil)

		_, err := meshctl.Invoke(fmt.Sprintf("install -f %s", chartOverride))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should set values for flags", func() {
		chartOverride := "chartOverride.tgz"
		releaseName := "releaseName"
		installNamespace := "service-mesh-hub"
		createNamespace := false
		valuesFile1 := "values-file1"
		valuesFile2 := "values-file2"
		installerconfig := &types.InstallerConfig{
			DryRun:             true,
			CreateNamespace:    true,
			Verbose:            true,
			InstallNamespace:   installNamespace,
			ReleaseName:        releaseName,
			ReleaseUri:         chartOverride,
			ValuesFiles:        []string{valuesFile1, valuesFile2},
			PreInstallMessage:  install.PreInstallMessage,
			PostInstallMessage: install.PostInstallMessage,
		}
		mockKubeLoader.
			EXPECT().
			GetRestConfigForContext("", "").
			Return(&rest.Config{}, nil)
		mockHelmInstaller.
			EXPECT().
			Install(installerconfig).
			Return(nil)

		_, err := meshctl.Invoke(
			fmt.Sprintf(
				"install -f %s --dry-run --create-namespace %t --verbose --release-name %s --namespace %s --values %s,%s",
				chartOverride, createNamespace, releaseName, installNamespace, valuesFile1, valuesFile2))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should fail if invalid version override supplied", func() {
		invalidVersion := "123"
		_, err := meshctl.Invoke(fmt.Sprintf("install --version %s", invalidVersion))
		Expect(err).To(testutils.HaveInErrorChain(install.InvalidVersionErr(invalidVersion)))
	})
})
