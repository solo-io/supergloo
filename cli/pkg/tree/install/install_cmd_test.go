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
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mock_kube "github.com/solo-io/service-mesh-hub/cli/pkg/common/kube/mocks"
	cli_mocks "github.com/solo-io/service-mesh-hub/cli/pkg/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/install"
	mock_auth "github.com/solo-io/service-mesh-hub/pkg/auth/mocks"
	mock_clients "github.com/solo-io/service-mesh-hub/pkg/clients/mocks"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var _ = Describe("Install", func() {
	var (
		ctrl                          *gomock.Controller
		ctx                           context.Context
		mockKubeLoader                *cli_mocks.MockKubeLoader
		meshctl                       *cli_test.MockMeshctl
		mockHelmClient                *mock_types.MockHelmClient
		mockHelmInstaller             *mock_types.MockInstaller
		mockClusterRegistrationClient *mock_clients.MockClusterRegistrationClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockHelmClient = mock_types.NewMockHelmClient(ctrl)
		mockHelmInstaller = mock_types.NewMockInstaller(ctrl)
		mockKubeLoader = cli_mocks.NewMockKubeLoader(ctrl)
		mockClusterRegistrationClient = mock_clients.NewMockClusterRegistrationClient(ctrl)
		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Clients:        common.Clients{},
			KubeClients: common.KubeClients{
				HelmInstallerFactory: func(helmClient types.HelmClient) types.Installer {
					return mockHelmInstaller
				},
				HelmClientFileConfigFactory: func(kubeConfig, kubeContext string) types.HelmClient {
					return mockHelmClient
				},
			},
			KubeLoader: mockKubeLoader,
			Ctx:        ctx,
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

	It("should register if flag is set", func() {
		chartOverride := "chartOverride.tgz"
		installNamespace := "service-mesh-hub"
		installerconfig := &types.InstallerConfig{
			CreateNamespace:    true,
			ReleaseUri:         chartOverride,
			InstallNamespace:   installNamespace,
			ReleaseName:        installNamespace,
			ValuesFiles:        []string{},
			PreInstallMessage:  install.PreInstallMessage,
			PostInstallMessage: install.PostInstallMessage,
		}
		mockKubeLoader.
			EXPECT().
			GetRestConfigForContext("", "").
			Return(&rest.Config{}, nil).Times(2)
		mockHelmInstaller.
			EXPECT().
			Install(installerconfig).
			Return(nil)

		clusterName := "test-cluster-name"
		contextABC := "contextABC"
		clusterABC := "clusterABC"
		testServerABC := "test-server-abc"

		contextDEF := "contextDEF"
		clusterDEF := "clusterDEF"
		testServerDEF := "test-server-def"

		secretClient := mock_kubernetes_core.NewMockSecretClient(ctrl)
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		authClient := mock_auth.NewMockClusterAuthorization(ctrl)
		configVerifier := cli_mocks.NewMockMasterKubeConfigVerifier(ctrl)
		clusterClient := mock_zephyr_discovery.NewMockKubernetesClusterClient(ctrl)
		kubeConverter := mock_kube.NewMockConverter(ctrl)

		configVerifier.EXPECT().Verify("", "").Return(nil)

		targetConfig := &clientcmd.DirectClientConfig{}
		cxt := clientcmdapi.Config{
			CurrentContext: contextABC,
			Contexts: map[string]*api.Context{
				contextABC: {Cluster: clusterABC},
				contextDEF: {Cluster: clusterDEF},
			},
			Clusters: map[string]*api.Cluster{
				clusterABC: {Server: testServerABC},
				clusterDEF: {Server: testServerDEF},
			},
		}
		mockKubeLoader.EXPECT().GetConfigWithContext("", "", contextABC).Return(targetConfig, nil)
		mockClusterRegistrationClient.
			EXPECT().
			Register(
				ctx,
				targetConfig,
				clusterName,
				installNamespace,
				false,
				false,
				"",
				contextABC,
				nil,
			)
		mockKubeLoader.EXPECT().GetRawConfigForContext("", "").Return(cxt, nil)

		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Clients: common.Clients{
				MasterClusterVerifier: configVerifier,
				KubeConverter:         kubeConverter,
			},
			KubeClients: common.KubeClients{
				ClusterRegistrationClient: mockClusterRegistrationClient,
				ClusterAuthorization:      authClient,
				HelmInstallerFactory: func(helmClient types.HelmClient) types.Installer {
					return mockHelmInstaller
				},
				HelmClientFileConfigFactory: func(kubeConfig, kubeContext string) types.HelmClient {
					return mockHelmClient
				},
				KubeClusterClient:  clusterClient,
				HealthCheckClients: healthcheck_types.Clients{},
				SecretClient:       secretClient,
				NamespaceClient:    namespaceClient,
				UninstallClients:   common.UninstallClients{},
			},
			KubeLoader: mockKubeLoader,
			Ctx:        ctx,
		}

		_, err := meshctl.Invoke(
			fmt.Sprintf("install --register --cluster-name %s -f %s", clusterName, chartOverride))
		Expect(err).NotTo(HaveOccurred())

	})

	It("should fail if invalid version override supplied", func() {
		invalidVersion := "123"
		_, err := meshctl.Invoke(fmt.Sprintf("install --version %s", invalidVersion))
		Expect(err).To(testutils.HaveInErrorChain(install.InvalidVersionErr(invalidVersion)))
	})
})
