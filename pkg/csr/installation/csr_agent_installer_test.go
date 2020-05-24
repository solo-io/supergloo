package installation_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helminstall"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	mock_types "github.com/solo-io/go-utils/installutils/helminstall/types/mocks"
	"github.com/solo-io/go-utils/testutils"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	mock_version "github.com/solo-io/service-mesh-hub/pkg/container-runtime/version/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/csr/installation"
	"github.com/solo-io/service-mesh-hub/pkg/kube/helm"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = Describe("CSR Agent Installer", func() {
	var (
		ctrl                        *gomock.Controller
		ctx                         context.Context
		mockHelmFileClient          *mock_types.MockHelmClient
		mockHelmFileClientFactory   helm.HelmClientForFileConfigFactory
		mockHelmMemoryClient        *mock_types.MockHelmClient
		mockHelmMemoryClientFactory helm.HelmClientForMemoryConfigFactory
		mockHelmInstaller           *mock_types.MockInstaller
		mockDeployedVersionFinder   *mock_version.MockDeployedVersionFinder
		csrAgentInstaller           installation.CsrAgentInstaller
		testErr                     = eris.New("test-err")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockHelmFileClient = mock_types.NewMockHelmClient(ctrl)
		mockHelmMemoryClient = mock_types.NewMockHelmClient(ctrl)
		mockHelmInstaller = mock_types.NewMockInstaller(ctrl)
		mockDeployedVersionFinder = mock_version.NewMockDeployedVersionFinder(ctrl)
		mockHelmFileClientFactory = func(kubeConfig, kubeContext string) types.HelmClient {
			return mockHelmFileClient
		}
		mockHelmMemoryClientFactory = func(config clientcmd.ClientConfig) types.HelmClient {
			return mockHelmMemoryClient
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can install the correct open source version of csr-agent from file kubeconfig", func() {
		csrAgentInstaller = installation.NewCsrAgentInstaller(
			mockHelmFileClientFactory,
			mockHelmMemoryClientFactory,
			mockDeployedVersionFinder,
			func(helmClient types.HelmClient) types.Installer {
				Expect(helmClient).To(BeIdenticalTo(mockHelmFileClient))
				return mockHelmInstaller
			},
		)
		openSourceVersion := "1.0.0"
		kubeConfig := "kube-config"
		kubeContext := "remote-kube-context"
		remoteWriteNamespace := "remote-write-namespace"
		releaseName := "csr-agent-release-name"
		mockDeployedVersionFinder.EXPECT().
			OpenSourceVersion(ctx, container_runtime.GetWriteNamespace()).
			Return(openSourceVersion, nil)
		mockHelmInstaller.EXPECT().
			Install(&types.InstallerConfig{
				InstallNamespace: remoteWriteNamespace,
				CreateNamespace:  true,
				ReleaseName:      releaseName,
				ReleaseUri:       fmt.Sprintf(installation.CsrAgentChartUriTemplate, openSourceVersion),
			}).
			Return(nil)

		err := csrAgentInstaller.Install(
			ctx,
			&installation.CsrAgentInstallOptions{
				KubeConfig: installation.KubeConfig{
					KubeConfigPath: kubeConfig,
					KubeContext:    kubeContext,
				},
				SmhInstallNamespace:  container_runtime.GetWriteNamespace(),
				ReleaseName:          releaseName,
				RemoteWriteNamespace: remoteWriteNamespace,
			},
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can install the correct open source version of csr-agent from memory kubeconfig", func() {
		csrAgentInstaller = installation.NewCsrAgentInstaller(
			mockHelmFileClientFactory,
			mockHelmMemoryClientFactory,
			mockDeployedVersionFinder,
			func(helmClient types.HelmClient) types.Installer {
				Expect(helmClient).To(BeIdenticalTo(mockHelmMemoryClient))
				return mockHelmInstaller
			},
		)
		openSourceVersion := "1.0.0"
		kubeConfig := &clientcmd.DirectClientConfig{}
		remoteWriteNamespace := "remote-write-namespace"
		releaseName := "csr-agent-release-name"
		mockDeployedVersionFinder.EXPECT().
			OpenSourceVersion(ctx, container_runtime.GetWriteNamespace()).
			Return(openSourceVersion, nil)
		mockHelmInstaller.EXPECT().
			Install(&types.InstallerConfig{
				InstallNamespace: remoteWriteNamespace,
				CreateNamespace:  true,
				ReleaseName:      releaseName,
				ReleaseUri:       fmt.Sprintf(installation.CsrAgentChartUriTemplate, openSourceVersion),
			}).
			Return(nil)

		err := csrAgentInstaller.Install(
			ctx,
			&installation.CsrAgentInstallOptions{
				KubeConfig:           installation.KubeConfig{KubeConfig: kubeConfig},
				SmhInstallNamespace:  container_runtime.GetWriteNamespace(),
				ReleaseName:          releaseName,
				RemoteWriteNamespace: remoteWriteNamespace,
			},
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can install csr-agent from a locally packaged chart", func() {
		csrAgentInstaller = installation.NewCsrAgentInstaller(
			mockHelmFileClientFactory,
			mockHelmMemoryClientFactory,
			mockDeployedVersionFinder,
			func(helmClient types.HelmClient) types.Installer {
				Expect(helmClient).To(BeIdenticalTo(mockHelmFileClient))
				return mockHelmInstaller
			},
		)
		openSourceVersion := "1.0.0"
		kubeConfig := "kube-config"
		kubeContext := "remote-kube-context"
		remoteWriteNamespace := "remote-write-namespace"
		releaseName := "csr-agent-release-name"

		mockDeployedVersionFinder.EXPECT().
			OpenSourceVersion(ctx, container_runtime.GetWriteNamespace()).
			Return(openSourceVersion, nil)
		mockHelmInstaller.EXPECT().
			Install(&types.InstallerConfig{
				InstallNamespace: remoteWriteNamespace,
				CreateNamespace:  true,
				ReleaseName:      releaseName,
				ReleaseUri:       fmt.Sprintf(installation.LocallyPackagedChartTemplate, openSourceVersion),
			}).
			Return(nil)

		err := csrAgentInstaller.Install(
			ctx,
			&installation.CsrAgentInstallOptions{
				KubeConfig: installation.KubeConfig{
					KubeConfigPath: kubeConfig,
					KubeContext:    kubeContext,
				},
				SmhInstallNamespace:  container_runtime.GetWriteNamespace(),
				ReleaseName:          releaseName,
				RemoteWriteNamespace: remoteWriteNamespace,
				UseDevCsrAgentChart:  true,
			},
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("does not complain if csr-agent is already deployed", func() {
		csrAgentInstaller = installation.NewCsrAgentInstaller(
			mockHelmFileClientFactory,
			mockHelmMemoryClientFactory,
			mockDeployedVersionFinder,
			func(helmClient types.HelmClient) types.Installer {
				Expect(helmClient).To(BeIdenticalTo(mockHelmFileClient))
				return mockHelmInstaller
			},
		)
		openSourceVersion := "1.0.0"
		kubeConfig := "kube-config"
		kubeContext := "remote-kube-context"
		remoteWriteNamespace := "remote-write-namespace"
		releaseName := "csr-agent-release-name"

		mockDeployedVersionFinder.EXPECT().
			OpenSourceVersion(ctx, container_runtime.GetWriteNamespace()).
			Return(openSourceVersion, nil)
		mockHelmInstaller.EXPECT().
			Install(&types.InstallerConfig{
				InstallNamespace: remoteWriteNamespace,
				CreateNamespace:  true,
				ReleaseName:      releaseName,
				ReleaseUri:       fmt.Sprintf(installation.LocallyPackagedChartTemplate, openSourceVersion),
			}).
			Return(helminstall.ReleaseAlreadyInstalledErr(releaseName, remoteWriteNamespace))

		err := csrAgentInstaller.Install(
			ctx,
			&installation.CsrAgentInstallOptions{
				KubeConfig: installation.KubeConfig{
					KubeConfigPath: kubeConfig,
					KubeContext:    kubeContext,
				},
				SmhInstallNamespace:  container_runtime.GetWriteNamespace(),
				ReleaseName:          releaseName,
				RemoteWriteNamespace: remoteWriteNamespace,
				UseDevCsrAgentChart:  true,
			},
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("responds with the appropriate error if the helm install fails", func() {
		csrAgentInstaller = installation.NewCsrAgentInstaller(
			mockHelmFileClientFactory,
			mockHelmMemoryClientFactory,
			mockDeployedVersionFinder,
			func(helmClient types.HelmClient) types.Installer {
				Expect(helmClient).To(BeIdenticalTo(mockHelmFileClient))
				return mockHelmInstaller
			},
		)
		openSourceVersion := "1.0.0"
		kubeConfig := "kube-config"
		kubeContext := "remote-kube-context"
		remoteWriteNamespace := "remote-write-namespace"
		releaseName := "csr-agent-release-name"

		mockDeployedVersionFinder.EXPECT().
			OpenSourceVersion(ctx, container_runtime.GetWriteNamespace()).
			Return(openSourceVersion, nil)
		mockHelmInstaller.EXPECT().
			Install(&types.InstallerConfig{
				InstallNamespace: remoteWriteNamespace,
				CreateNamespace:  true,
				ReleaseName:      releaseName,
				ReleaseUri:       fmt.Sprintf(installation.LocallyPackagedChartTemplate, openSourceVersion),
			}).
			Return(testErr)

		err := csrAgentInstaller.Install(
			ctx,
			&installation.CsrAgentInstallOptions{
				KubeConfig: installation.KubeConfig{
					KubeConfigPath: kubeConfig,
					KubeContext:    kubeContext,
				},
				SmhInstallNamespace:  container_runtime.GetWriteNamespace(),
				ReleaseName:          releaseName,
				RemoteWriteNamespace: remoteWriteNamespace,
				UseDevCsrAgentChart:  true,
			},
		)
		Expect(err).To(testutils.HaveInErrorChain(installation.FailedToSetUpCsrAgent(testErr)))
	})

	It("should uninstall", func() {
		csrAgentInstaller = installation.NewCsrAgentInstaller(
			mockHelmFileClientFactory,
			mockHelmMemoryClientFactory,
			mockDeployedVersionFinder,
			func(helmClient types.HelmClient) types.Installer {
				Expect(helmClient).To(BeIdenticalTo(mockHelmMemoryClient))
				return mockHelmInstaller
			},
		)
		kubeConfig := &clientcmd.DirectClientConfig{}
		releaseName := "csr-agent-release-name"
		releaseNamespace := "remote-write-namespace"
		mockHelmUninstaller := mock_types.NewMockHelmUninstaller(ctrl)
		mockHelmMemoryClient.EXPECT().NewUninstall(releaseNamespace).Return(mockHelmUninstaller, nil)
		mockHelmUninstaller.EXPECT().Run(releaseName).Return(nil, nil)
		err := csrAgentInstaller.Uninstall(
			&installation.CsrAgentUninstallOptions{
				KubeConfig:       installation.KubeConfig{KubeConfig: kubeConfig},
				ReleaseName:      releaseName,
				ReleaseNamespace: releaseNamespace,
			},
		)
		Expect(err).NotTo(HaveOccurred())
	})
})
