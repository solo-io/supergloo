package register_test

import (
	"context"
	"fmt"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	cli_mocks "github.com/solo-io/service-mesh-hub/cli/pkg/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	cluster_internal "github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/internal"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register"
	cluster_registration "github.com/solo-io/service-mesh-hub/pkg/clients/cluster-registration"
	mock_clients "github.com/solo-io/service-mesh-hub/pkg/clients/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = Describe("Cluster Operations", func() {
	var (
		ctrl                          *gomock.Controller
		ctx                           context.Context
		secretClient                  *mock_kubernetes_core.MockSecretClient
		kubeLoader                    *cli_mocks.MockKubeLoader
		meshctl                       *cli_test.MockMeshctl
		configVerifier                *cli_mocks.MockMasterKubeConfigVerifier
		mockClusterRegistrationClient *mock_clients.MockClusterRegistrationClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()

		secretClient = mock_kubernetes_core.NewMockSecretClient(ctrl)
		kubeLoader = cli_mocks.NewMockKubeLoader(ctrl)
		configVerifier = cli_mocks.NewMockMasterKubeConfigVerifier(ctrl)
		mockClusterRegistrationClient = mock_clients.NewMockClusterRegistrationClient(ctrl)
		meshctl = &cli_test.MockMeshctl{
			KubeClients: common.KubeClients{
				SecretClient:              secretClient,
				ClusterRegistrationClient: mockClusterRegistrationClient,
			},
			Clients: common.Clients{
				MasterClusterVerifier: configVerifier,
			},
			MockController: ctrl,
			KubeLoader:     kubeLoader,
			Ctx:            ctx,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Cluster Registration", func() {
		var (
			contextDEF       = "contextDEF"
			targetRestConfig = &rest.Config{Host: "www.test.com", TLSClientConfig: rest.TLSClientConfig{CertData: []byte("secret!!!")}}
		)

		It("works", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfigPath := "~/.kube/target-config"
			remoteKubeConfig := &clientcmd.DirectClientConfig{}
			clusterName := "test-cluster-name"
			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)

			kubeLoader.EXPECT().GetConfigWithContext("", remoteKubeConfigPath, "").Return(remoteKubeConfig, nil)
			mockClusterRegistrationClient.
				EXPECT().
				Register(
					ctx,
					remoteKubeConfig,
					clusterName,
					env.GetWriteNamespace(),
					"",
					register.MeshctlDiscoverySource,
					cluster_registration.ClusterRegisterOpts{},
				).
				Return(nil)
			kubeLoader.EXPECT().GetRestConfigForContext(localKubeConfig, "").Return(targetRestConfig, nil)

			_, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name %s", remoteKubeConfigPath, localKubeConfig, clusterName))

			Expect(err).NotTo(HaveOccurred())
		})

		It("works if you implicitly set master through KUBECONFIG", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfigPath := "~/.kube/target-config"
			clusterName := "test-cluster-name"
			os.Setenv("KUBECONFIG", localKubeConfig)
			defer os.Setenv("KUBECONFIG", "")

			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			remoteKubeConfig := &clientcmd.DirectClientConfig{}
			kubeLoader.EXPECT().GetConfigWithContext("", remoteKubeConfigPath, "").Return(remoteKubeConfig, nil)
			mockClusterRegistrationClient.
				EXPECT().
				Register(
					ctx,
					remoteKubeConfig,
					clusterName,
					env.GetWriteNamespace(),
					"",
					register.MeshctlDiscoverySource,
					cluster_registration.ClusterRegisterOpts{},
				).
				Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(targetRestConfig, nil)

			_, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig "+
				"%s --remote-cluster-name test-cluster-name", remoteKubeConfigPath))
			Expect(err).NotTo(HaveOccurred())
		})

		It("works if you use a different context for the remote and local config", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfigPath := "~/.kube/target-config"
			clusterName := "test-cluster-name"
			os.Setenv("KUBECONFIG", localKubeConfig)
			defer os.Setenv("KUBECONFIG", "")

			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(targetRestConfig, nil)
			remoteKubeConfig := &clientcmd.DirectClientConfig{}
			kubeLoader.EXPECT().GetConfigWithContext("", remoteKubeConfigPath, contextDEF).Return(remoteKubeConfig, nil)
			mockClusterRegistrationClient.
				EXPECT().
				Register(
					ctx,
					remoteKubeConfig,
					clusterName,
					env.GetWriteNamespace(),
					contextDEF,
					register.MeshctlDiscoverySource,
					cluster_registration.ClusterRegisterOpts{},
				).
				Return(nil)

			_, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s "+
				"--remote-context %s --remote-cluster-name test-cluster-name", remoteKubeConfigPath, contextDEF))
			Expect(err).NotTo(HaveOccurred())
		})

		It("will fail if local or remote cluster config fails to initialize", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfigPath := "~/.kube/target-config"
			os.Setenv("KUBECONFIG", localKubeConfig)
			defer os.Setenv("KUBECONFIG", "")
			testErr := eris.New("hello")
			remoteKubeConfig := &clientcmd.DirectClientConfig{}
			kubeLoader.EXPECT().GetConfigWithContext("", remoteKubeConfigPath, "").Return(remoteKubeConfig, nil)
			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(nil, testErr)
			_, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name test-cluster-name", remoteKubeConfigPath, localKubeConfig))
			Expect(err).To(HaveInErrorChain(testErr))
		})

		It("errors if a master or target cluster are not set", func() {
			os.Setenv("KUBECONFIG", "")

			stdout, err := meshctl.Invoke("cluster register")
			Expect(stdout).To(BeEmpty())
			Expect(err.Error()).To(ContainSubstring("\"remote-cluster-name\" not set"))

			kubeConfigPath := ""
			testErr := eris.New("hello")

			stdout, err = meshctl.Invoke("cluster register --remote-cluster-name hello")
			Expect(stdout).To(BeEmpty())
			Expect(err).To(HaveInErrorChain(cluster_internal.NoRemoteConfigSpecifiedError))

			configVerifier.EXPECT().Verify(kubeConfigPath, "").Return(testErr)

			stdout, err = meshctl.Invoke("cluster register --remote-cluster-name hello --remote-context hello")
			Expect(stdout).To(BeEmpty())
			Expect(err).To(HaveInErrorChain(testErr))

			configVerifier.EXPECT().Verify(kubeConfigPath, "").Return(testErr)

			stdout, err = meshctl.Invoke("cluster register --remote-cluster-name hello --remote-kubeconfig hello")
			Expect(stdout).To(BeEmpty())
			Expect(err).To(HaveInErrorChain(testErr))
		})

		It("can use the same kube config with different contexts", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteContext := contextDEF
			clusterName := "test-cluster-name"
			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			remoteKubeConfig := &clientcmd.DirectClientConfig{}
			kubeLoader.EXPECT().GetConfigWithContext("", localKubeConfig, remoteContext).Return(remoteKubeConfig, nil)
			mockClusterRegistrationClient.
				EXPECT().
				Register(
					ctx,
					remoteKubeConfig,
					clusterName,
					env.GetWriteNamespace(),
					contextDEF,
					register.MeshctlDiscoverySource,
					cluster_registration.ClusterRegisterOpts{},
				).
				Return(nil)
			kubeLoader.EXPECT().GetRestConfigForContext(localKubeConfig, "").Return(targetRestConfig, nil)
			_, err := meshctl.Invoke(fmt.Sprintf("cluster register --kubeconfig %s --remote-cluster-name %s --remote-context %s", localKubeConfig, clusterName, remoteContext))

			Expect(err).NotTo(HaveOccurred())
		})

		It("can register with the CSR agent being installed from a dev chart", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfigPath := "~/.kube/target-config"
			clusterName := "test-cluster-name"
			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			remoteKubeConfig := &clientcmd.DirectClientConfig{}
			kubeLoader.EXPECT().GetConfigWithContext("", remoteKubeConfigPath, "").Return(remoteKubeConfig, nil)
			mockClusterRegistrationClient.
				EXPECT().
				Register(
					ctx,
					remoteKubeConfig,
					clusterName,
					env.GetWriteNamespace(),
					"",
					register.MeshctlDiscoverySource,
					cluster_registration.ClusterRegisterOpts{
						UseDevCsrAgentChart: true,
					},
				).
				Return(nil)

			kubeLoader.EXPECT().GetRestConfigForContext(localKubeConfig, "").Return(targetRestConfig, nil)

			_, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name %s --dev-csr-agent-chart", remoteKubeConfigPath, localKubeConfig, clusterName))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
