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
	cli_mocks "github.com/solo-io/service-mesh-hub/cli/pkg/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register/csr"
	mock_csr "github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register/csr/mocks"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/install"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	mock_auth "github.com/solo-io/service-mesh-hub/pkg/auth/mocks"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/version"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	It("should rergister is flag is sett", func() {
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
		clusterClient := mock_core.NewMockKubernetesClusterClient(ctrl)
		csrAgentInstaller := mock_csr.NewMockCsrAgentInstaller(ctrl)

		configVerifier.EXPECT().Verify("", "").Return(nil)

		serviceAccountRef := &core_types.ResourceRef{
			Name:      "test-cluster-name",
			Namespace: env.GetWriteNamespace(),
		}

		expectedKubeConfig := func(server string) string {
			return fmt.Sprintf(`apiVersion: v1
clusters:
- cluster:
    server: %s
  name: test-cluster-name
contexts:
- context:
    cluster: test-cluster-name
    user: test-cluster-name
  name: test-cluster-name
current-context: test-cluster-name
kind: Config
preferences: {}
users:
- name: test-cluster-name
  user:
    token: alphanumericgarbage
`, server)
		}
		targetRestConfig := &rest.Config{Host: "www.test.com", TLSClientConfig: rest.TLSClientConfig{CertData: []byte("secret!!!")}}
		configForServiceAccount := &rest.Config{Host: "www.test.com", BearerToken: "alphanumericgarbage"}
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
		mockKubeLoader.EXPECT().GetRestConfigForContext("", contextABC).Return(targetRestConfig, nil)
		authClient.EXPECT().CreateAuthConfigForCluster(ctx, targetRestConfig, serviceAccountRef).Return(configForServiceAccount, nil)
		mockKubeLoader.EXPECT().GetRawConfigForContext("", "").Return(cxt, nil)
		mockKubeLoader.EXPECT().GetRawConfigForContext("", contextABC).Return(cxt, nil)
		clusterClient.EXPECT().Get(ctx,
			client.ObjectKey{
				Name:      clusterName,
				Namespace: env.GetWriteNamespace(),
			}).Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
				Name:      serviceAccountRef.Name,
				Namespace: env.GetWriteNamespace(),
			},
			Data: map[string][]byte{
				clusterName: []byte(expectedKubeConfig(testServerABC)),
			},
			Type: v1.SecretTypeOpaque,
		}

		secretClient.
			EXPECT().
			UpsertData(ctx, secret).
			Return(nil)

		namespaceClient.
			EXPECT().
			Get(ctx, env.GetWriteNamespace()).
			Return(nil, nil)

		csrAgentInstaller.EXPECT().
			Install(ctx, &csr.CsrAgentInstallOptions{
				KubeConfig:           "",
				KubeContext:          contextABC,
				ClusterName:          clusterName,
				SmhInstallNamespace:  env.GetWriteNamespace(),
				RemoteWriteNamespace: env.GetWriteNamespace(),
				ReleaseName:          cliconstants.CsrAgentReleaseName,
			}).
			Return(nil)

		clusterClient.EXPECT().Upsert(ctx, &v1alpha1.KubernetesCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: env.GetWriteNamespace(),
			},
			Spec: discovery_types.KubernetesClusterSpec{
				SecretRef: &core_types.ResourceRef{
					Name:      secret.GetName(),
					Namespace: secret.GetNamespace(),
				},
				WriteNamespace: env.GetWriteNamespace(),
			},
		}).Return(nil)

		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Clients: common.Clients{
				MasterClusterVerifier: configVerifier,
				ClusterRegistrationClients: common.ClusterRegistrationClients{
					CsrAgentInstallerFactory: func(helmInstaller types.Installer, deployedVersionFinder version.DeployedVersionFinder) csr.CsrAgentInstaller {
						return csrAgentInstaller
					},
				},
			},
			KubeClients: common.KubeClients{
				ClusterAuthorization: authClient,
				HelmInstaller:        mockHelmInstaller,
				KubeClusterClient:    clusterClient,
				HealthCheckClients:   healthcheck_types.Clients{},
				SecretsClient:        secretClient,
				NamespaceClient:      namespaceClient,
				UninstallClients:     common.UninstallClients{},
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
