package csr_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helminstall"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/register/csr"
	"github.com/solo-io/mesh-projects/pkg/env"
	mock_version "github.com/solo-io/mesh-projects/pkg/version/mocks"
	mock_go_utils "github.com/solo-io/mesh-projects/test/mocks/go-utils"
)

var _ = Describe("CSR Agent Installer", func() {
	var (
		ctrl    *gomock.Controller
		ctx     context.Context
		testErr = eris.New("test-err")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can install the correct open source version of csr-agent", func() {
		helmInstaller := mock_go_utils.NewMockInstaller(ctrl)
		deployedVersionFinder := mock_version.NewMockDeployedVersionFinder(ctrl)
		csrAgentInstaller := csr.NewCsrAgentInstaller(helmInstaller, deployedVersionFinder)
		openSourceVersion := "1.0.0"
		kubeConfig := "kube-config"
		kubeContext := "remote-kube-context"
		remoteWriteNamespace := "remote-write-namespace"
		releaseName := "csr-agent-release-name"

		deployedVersionFinder.EXPECT().
			OpenSourceVersion(ctx, env.DefaultWriteNamespace).
			Return(openSourceVersion, nil)
		helmInstaller.EXPECT().
			Install(&helminstall.InstallerConfig{
				KubeConfig:       kubeConfig,
				KubeContext:      kubeContext,
				InstallNamespace: remoteWriteNamespace,
				CreateNamespace:  true,
				ReleaseName:      releaseName,
				ReleaseUri:       fmt.Sprintf(csr.CsrAgentChartUriTemplate, openSourceVersion),
			}).
			Return(nil)

		err := csrAgentInstaller.Install(
			ctx,
			&csr.CsrAgentInstallOptions{
				KubeConfig:           kubeConfig,
				KubeContext:          kubeContext,
				ClusterName:          "remote-cluster-name",
				SmhInstallNamespace:  env.DefaultWriteNamespace,
				ReleaseName:          releaseName,
				RemoteWriteNamespace: remoteWriteNamespace,
			},
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can install csr-agent from a locally packaged chart", func() {
		helmInstaller := mock_go_utils.NewMockInstaller(ctrl)
		deployedVersionFinder := mock_version.NewMockDeployedVersionFinder(ctrl)
		csrAgentInstaller := csr.NewCsrAgentInstaller(helmInstaller, deployedVersionFinder)
		openSourceVersion := "1.0.0"
		kubeConfig := "kube-config"
		kubeContext := "remote-kube-context"
		remoteWriteNamespace := "remote-write-namespace"
		releaseName := "csr-agent-release-name"

		deployedVersionFinder.EXPECT().
			OpenSourceVersion(ctx, env.DefaultWriteNamespace).
			Return(openSourceVersion, nil)
		helmInstaller.EXPECT().
			Install(&helminstall.InstallerConfig{
				KubeConfig:       kubeConfig,
				KubeContext:      kubeContext,
				InstallNamespace: remoteWriteNamespace,
				CreateNamespace:  true,
				ReleaseName:      releaseName,
				ReleaseUri:       fmt.Sprintf(csr.LocallyPackagedChartTemplate, openSourceVersion),
			}).
			Return(nil)

		err := csrAgentInstaller.Install(
			ctx,
			&csr.CsrAgentInstallOptions{
				KubeConfig:           kubeConfig,
				KubeContext:          kubeContext,
				ClusterName:          "remote-cluster-name",
				SmhInstallNamespace:  env.DefaultWriteNamespace,
				ReleaseName:          releaseName,
				RemoteWriteNamespace: remoteWriteNamespace,
				UseDevCsrAgentChart:  true,
			},
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("does not complain if csr-agent is already deployed", func() {
		helmInstaller := mock_go_utils.NewMockInstaller(ctrl)
		deployedVersionFinder := mock_version.NewMockDeployedVersionFinder(ctrl)
		csrAgentInstaller := csr.NewCsrAgentInstaller(helmInstaller, deployedVersionFinder)
		openSourceVersion := "1.0.0"
		kubeConfig := "kube-config"
		kubeContext := "remote-kube-context"
		remoteWriteNamespace := "remote-write-namespace"
		releaseName := "csr-agent-release-name"

		deployedVersionFinder.EXPECT().
			OpenSourceVersion(ctx, env.DefaultWriteNamespace).
			Return(openSourceVersion, nil)
		helmInstaller.EXPECT().
			Install(&helminstall.InstallerConfig{
				KubeConfig:       kubeConfig,
				KubeContext:      kubeContext,
				InstallNamespace: remoteWriteNamespace,
				CreateNamespace:  true,
				ReleaseName:      releaseName,
				ReleaseUri:       fmt.Sprintf(csr.LocallyPackagedChartTemplate, openSourceVersion),
			}).
			Return(helminstall.ReleaseAlreadyInstalledErr(releaseName, remoteWriteNamespace))

		err := csrAgentInstaller.Install(
			ctx,
			&csr.CsrAgentInstallOptions{
				KubeConfig:           kubeConfig,
				KubeContext:          kubeContext,
				ClusterName:          "remote-cluster-name",
				SmhInstallNamespace:  env.DefaultWriteNamespace,
				ReleaseName:          releaseName,
				RemoteWriteNamespace: remoteWriteNamespace,
				UseDevCsrAgentChart:  true,
			},
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("responds with the appropriate error if the helm install fails", func() {
		helmInstaller := mock_go_utils.NewMockInstaller(ctrl)
		deployedVersionFinder := mock_version.NewMockDeployedVersionFinder(ctrl)
		csrAgentInstaller := csr.NewCsrAgentInstaller(helmInstaller, deployedVersionFinder)
		openSourceVersion := "1.0.0"
		kubeConfig := "kube-config"
		kubeContext := "remote-kube-context"
		remoteWriteNamespace := "remote-write-namespace"
		releaseName := "csr-agent-release-name"

		deployedVersionFinder.EXPECT().
			OpenSourceVersion(ctx, env.DefaultWriteNamespace).
			Return(openSourceVersion, nil)
		helmInstaller.EXPECT().
			Install(&helminstall.InstallerConfig{
				KubeConfig:       kubeConfig,
				KubeContext:      kubeContext,
				InstallNamespace: remoteWriteNamespace,
				CreateNamespace:  true,
				ReleaseName:      releaseName,
				ReleaseUri:       fmt.Sprintf(csr.LocallyPackagedChartTemplate, openSourceVersion),
			}).
			Return(testErr)

		err := csrAgentInstaller.Install(
			ctx,
			&csr.CsrAgentInstallOptions{
				KubeConfig:           kubeConfig,
				KubeContext:          kubeContext,
				ClusterName:          "remote-cluster-name",
				SmhInstallNamespace:  env.DefaultWriteNamespace,
				ReleaseName:          releaseName,
				RemoteWriteNamespace: remoteWriteNamespace,
				UseDevCsrAgentChart:  true,
			},
		)
		Expect(err).To(testutils.HaveInErrorChain(csr.FailedToSetUpCsrAgent(testErr)))
	})
})
