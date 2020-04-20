package register_test

import (
	"context"
	"fmt"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	cli_mocks "github.com/solo-io/service-mesh-hub/cli/pkg/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	cluster_internal "github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/internal"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register/csr"
	mock_csr "github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register/csr/mocks"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	mock_auth "github.com/solo-io/service-mesh-hub/pkg/auth/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/k8s_secrets/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/version"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Cluster Operations", func() {
	var (
		ctrl              *gomock.Controller
		ctx               context.Context
		secretClient      *mock_kubernetes_core.MockSecretClient
		namespaceClient   *mock_kubernetes_core.MockNamespaceClient
		authClient        *mock_auth.MockClusterAuthorization
		kubeLoader        *cli_mocks.MockKubeLoader
		meshctl           *cli_test.MockMeshctl
		configVerifier    *cli_mocks.MockMasterKubeConfigVerifier
		clusterClient     *mock_core.MockKubernetesClusterClient
		csrAgentInstaller *mock_csr.MockCsrAgentInstaller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()

		secretClient = mock_kubernetes_core.NewMockSecretClient(ctrl)
		namespaceClient = mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		authClient = mock_auth.NewMockClusterAuthorization(ctrl)
		kubeLoader = cli_mocks.NewMockKubeLoader(ctrl)
		configVerifier = cli_mocks.NewMockMasterKubeConfigVerifier(ctrl)
		clusterClient = mock_core.NewMockKubernetesClusterClient(ctrl)
		csrAgentInstaller = mock_csr.NewMockCsrAgentInstaller(ctrl)
		meshctl = &cli_test.MockMeshctl{
			KubeClients: common.KubeClients{
				ClusterAuthorization: authClient,
				SecretClient:         secretClient,
				NamespaceClient:      namespaceClient,
				KubeClusterClient:    clusterClient,
			},
			Clients: common.Clients{
				MasterClusterVerifier: configVerifier,
				ClusterRegistrationClients: common.ClusterRegistrationClients{
					CsrAgentInstallerFactory: func(_ types.Installer, _ version.DeployedVersionFinder) csr.CsrAgentInstaller {
						return csrAgentInstaller
					},
				},
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
			expectedKubeConfig = func(server string) string {
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
			serviceAccountRef = &zephyr_core_types.ResourceRef{
				Name:      "test-cluster-name",
				Namespace: env.GetWriteNamespace(),
			}

			contextABC    = "contextABC"
			clusterABC    = "clusterABC"
			testServerABC = "test-server-abc"

			contextDEF    = "contextDEF"
			clusterDEF    = "clusterDEF"
			testServerDEF = "test-server-def"

			targetRestConfig        = &rest.Config{Host: "www.test.com", TLSClientConfig: rest.TLSClientConfig{CertData: []byte("secret!!!")}}
			configForServiceAccount = &rest.Config{Host: "www.test.com", BearerToken: "alphanumericgarbage"}
			cxt                     = clientcmdapi.Config{
				CurrentContext: "contextABC",
				Contexts: map[string]*api.Context{
					contextABC: {Cluster: clusterABC},
					contextDEF: {Cluster: clusterDEF},
				},
				Clusters: map[string]*api.Cluster{
					clusterABC: {Server: testServerABC},
					clusterDEF: {Server: testServerDEF},
				},
			}
		)

		var expectUpsertSecretData = func(ctx context.Context, secret *k8s_core_types.Secret) {
			existing := &k8s_core_types.Secret{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "existing"},
			}
			secretClient.
				EXPECT().
				GetSecret(ctx, client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace}).
				Return(existing, nil)
			existing.Data = secret.Data
			secretClient.EXPECT().UpdateSecret(ctx, existing).Return(nil)
		}

		It("works", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			clusterName := "test-cluster-name"
			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)

			kubeLoader.EXPECT().GetRestConfigForContext(localKubeConfig, "").Return(targetRestConfig, nil)
			kubeLoader.EXPECT().GetRestConfigForContext(remoteKubeConfig, "").Return(targetRestConfig, nil)
			authClient.EXPECT().CreateAuthConfigForCluster(ctx, targetRestConfig, serviceAccountRef).Return(configForServiceAccount, nil)
			kubeLoader.EXPECT().GetRawConfigForContext(remoteKubeConfig, "").Return(cxt, nil)
			clusterClient.EXPECT().GetKubernetesCluster(ctx,
				client.ObjectKey{
					Name:      clusterName,
					Namespace: env.GetWriteNamespace(),
				}).Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

			secret := &k8s_core_types.Secret{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
					Name:      serviceAccountRef.Name,
					Namespace: env.GetWriteNamespace(),
				},
				Data: map[string][]byte{
					clusterName: []byte(expectedKubeConfig(testServerABC)),
				},
				Type: k8s_core_types.SecretTypeOpaque,
			}

			expectUpsertSecretData(ctx, secret)

			namespaceClient.
				EXPECT().
				GetNamespace(ctx, client.ObjectKey{Name: env.GetWriteNamespace()}).
				Return(nil, nil)

			csrAgentInstaller.EXPECT().
				Install(ctx, &csr.CsrAgentInstallOptions{
					KubeConfig:           remoteKubeConfig,
					KubeContext:          "",
					ClusterName:          clusterName,
					SmhInstallNamespace:  env.GetWriteNamespace(),
					RemoteWriteNamespace: env.GetWriteNamespace(),
					ReleaseName:          cliconstants.CsrAgentReleaseName,
				}).
				Return(nil)

			clusterClient.EXPECT().UpsertKubernetesClusterSpec(ctx, &zephyr_discovery.KubernetesCluster{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      clusterName,
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.KubernetesClusterSpec{
					SecretRef: &zephyr_core_types.ResourceRef{
						Name:      secret.GetName(),
						Namespace: secret.GetNamespace(),
					},
					WriteNamespace: env.GetWriteNamespace(),
				},
			}).Return(nil)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name %s", remoteKubeConfig, localKubeConfig, clusterName))

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal(`Successfully wrote service account to remote cluster...
Successfully set up CSR agent...
Successfully wrote kube config secret to master cluster...

Cluster test-cluster-name is now registered in your Service Mesh Hub installation
`))
		})

		It("works if you implicitly set master through KUBECONFIG", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			os.Setenv("KUBECONFIG", localKubeConfig)
			defer os.Setenv("KUBECONFIG", "")

			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(targetRestConfig, nil)

			kubeLoader.
				EXPECT().
				GetRestConfigForContext(remoteKubeConfig, "").
				Return(targetRestConfig, nil)

			authClient.
				EXPECT().
				CreateAuthConfigForCluster(ctx, targetRestConfig, serviceAccountRef).
				Return(configForServiceAccount, nil)

			kubeLoader.
				EXPECT().
				GetRawConfigForContext(remoteKubeConfig, "").
				Return(cxt, nil)

			clusterClient.EXPECT().GetKubernetesCluster(ctx,
				client.ObjectKey{
					Name:      "test-cluster-name",
					Namespace: env.GetWriteNamespace(),
				}).Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

			secret := &k8s_core_types.Secret{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
					Name:      serviceAccountRef.Name,
					Namespace: env.GetWriteNamespace(),
				},
				Data: map[string][]byte{
					"test-cluster-name": []byte(expectedKubeConfig(testServerABC)),
				},
				Type: k8s_core_types.SecretTypeOpaque,
			}

			expectUpsertSecretData(ctx, secret)

			namespaceClient.
				EXPECT().
				GetNamespace(ctx, client.ObjectKey{Name: env.GetWriteNamespace()}).
				Return(nil, nil)

			csrAgentInstaller.
				EXPECT().
				Install(ctx, &csr.CsrAgentInstallOptions{
					KubeConfig:           remoteKubeConfig,
					ClusterName:          "test-cluster-name",
					SmhInstallNamespace:  env.GetWriteNamespace(),
					RemoteWriteNamespace: env.GetWriteNamespace(),
					ReleaseName:          cliconstants.CsrAgentReleaseName,
				})

			clusterClient.
				EXPECT().
				UpsertKubernetesClusterSpec(ctx, &zephyr_discovery.KubernetesCluster{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "test-cluster-name",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: zephyr_discovery_types.KubernetesClusterSpec{
						SecretRef: &zephyr_core_types.ResourceRef{
							Name:      secret.GetName(),
							Namespace: secret.GetNamespace(),
						},
						WriteNamespace: env.GetWriteNamespace(),
					},
				}).
				Return(nil)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig "+
				"%s --remote-cluster-name test-cluster-name", remoteKubeConfig))

			Expect(stdout).To(Equal(`Successfully wrote service account to remote cluster...
Successfully set up CSR agent...
Successfully wrote kube config secret to master cluster...

Cluster test-cluster-name is now registered in your Service Mesh Hub installation
`))
			Expect(err).NotTo(HaveOccurred())
		})

		It("works if you use a different context for the remote and local config", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			os.Setenv("KUBECONFIG", localKubeConfig)
			defer os.Setenv("KUBECONFIG", "")

			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(targetRestConfig, nil)

			kubeLoader.
				EXPECT().
				GetRestConfigForContext(remoteKubeConfig, contextDEF).
				Return(targetRestConfig, nil)

			authClient.
				EXPECT().
				CreateAuthConfigForCluster(ctx, targetRestConfig, serviceAccountRef).
				Return(configForServiceAccount, nil)

			kubeLoader.
				EXPECT().
				GetRawConfigForContext(remoteKubeConfig, contextDEF).
				Return(cxt, nil)

			clusterClient.EXPECT().GetKubernetesCluster(ctx,
				client.ObjectKey{
					Name:      "test-cluster-name",
					Namespace: env.GetWriteNamespace(),
				}).Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

			secret := &k8s_core_types.Secret{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
					Name:      serviceAccountRef.Name,
					Namespace: env.GetWriteNamespace(),
				},
				Data: map[string][]byte{
					"test-cluster-name": []byte(expectedKubeConfig(testServerDEF)),
				},
				Type: k8s_core_types.SecretTypeOpaque,
			}

			expectUpsertSecretData(ctx, secret)

			namespaceClient.
				EXPECT().
				GetNamespace(ctx, client.ObjectKey{Name: env.GetWriteNamespace()}).
				Return(nil, nil)

			csrAgentInstaller.
				EXPECT().
				Install(ctx, &csr.CsrAgentInstallOptions{
					KubeConfig:           remoteKubeConfig,
					KubeContext:          contextDEF,
					ClusterName:          "test-cluster-name",
					SmhInstallNamespace:  env.GetWriteNamespace(),
					RemoteWriteNamespace: env.GetWriteNamespace(),
					ReleaseName:          cliconstants.CsrAgentReleaseName,
				})

			clusterClient.
				EXPECT().
				UpsertKubernetesClusterSpec(ctx, &zephyr_discovery.KubernetesCluster{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "test-cluster-name",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: zephyr_discovery_types.KubernetesClusterSpec{
						SecretRef: &zephyr_core_types.ResourceRef{
							Name:      secret.GetName(),
							Namespace: secret.GetNamespace(),
						},
						WriteNamespace: env.GetWriteNamespace(),
					},
				}).
				Return(nil)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s "+
				"--remote-context %s --remote-cluster-name test-cluster-name", remoteKubeConfig, contextDEF))

			Expect(stdout).To(Equal(`Successfully wrote service account to remote cluster...
Successfully set up CSR agent...
Successfully wrote kube config secret to master cluster...

Cluster test-cluster-name is now registered in your Service Mesh Hub installation
`))
			Expect(err).NotTo(HaveOccurred())
		})

		It("will fail if local or remote cluster config fails to initialize", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			os.Setenv("KUBECONFIG", localKubeConfig)
			defer os.Setenv("KUBECONFIG", "")
			testErr := eris.New("hello")

			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(nil, testErr)

			_, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name test-cluster-name", remoteKubeConfig, localKubeConfig))

			Expect(err).To(HaveInErrorChain(common.FailedLoadingMasterConfig(testErr)))

			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(targetRestConfig, nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(remoteKubeConfig, "").
				Return(nil, testErr)

			_, err = meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name test-cluster-name", remoteKubeConfig, localKubeConfig))

			Expect(err).To(HaveInErrorChain(register.FailedLoadingRemoteConfig(testErr)))
		})

		It("will fail if unable to create auth config", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			os.Setenv("KUBECONFIG", localKubeConfig)
			defer os.Setenv("KUBECONFIG", "")
			testErr := eris.New("hello")

			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(targetRestConfig, nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(remoteKubeConfig, "").
				Return(targetRestConfig, nil)
			authClient.
				EXPECT().
				CreateAuthConfigForCluster(ctx, targetRestConfig, serviceAccountRef).
				Return(nil, testErr)

			namespaceClient.
				EXPECT().
				GetNamespace(ctx, client.ObjectKey{Name: env.GetWriteNamespace()}).
				Return(nil, nil)

			clusterClient.EXPECT().GetKubernetesCluster(ctx,
				client.ObjectKey{
					Name:      "test-cluster-name",
					Namespace: env.GetWriteNamespace(),
				}).Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name test-cluster-name", remoteKubeConfig, localKubeConfig))

			Expect(err).To(Equal(testErr))
			Expect(stdout).To(ContainSubstring(register.FailedToCreateAuthToken(serviceAccountRef, remoteKubeConfig, "")))
		})

		It("will create namespace if it does not exist", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			os.Setenv("KUBECONFIG", localKubeConfig)
			defer os.Setenv("KUBECONFIG", "")
			testErr := eris.New("hello")

			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(targetRestConfig, nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(remoteKubeConfig, "").
				Return(targetRestConfig, nil)
			authClient.
				EXPECT().
				CreateAuthConfigForCluster(ctx, targetRestConfig, serviceAccountRef).
				Return(nil, testErr)

			namespaceClient.
				EXPECT().
				GetNamespace(ctx, client.ObjectKey{Name: env.GetWriteNamespace()}).
				Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))
			namespaceClient.
				EXPECT().
				CreateNamespace(ctx, &k8s_core_types.Namespace{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: env.GetWriteNamespace()},
				}).
				Return(nil)

			clusterClient.EXPECT().GetKubernetesCluster(ctx,
				client.ObjectKey{
					Name:      "test-cluster-name",
					Namespace: env.GetWriteNamespace(),
				}).Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name test-cluster-name", remoteKubeConfig, localKubeConfig))

			Expect(err).To(Equal(testErr))
			Expect(stdout).To(ContainSubstring(register.FailedToCreateAuthToken(serviceAccountRef, remoteKubeConfig, "")))
		})

		It("will fail if unable to verify kube cluster already exists", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			os.Setenv("KUBECONFIG", localKubeConfig)
			defer os.Setenv("KUBECONFIG", "")
			testErr := eris.New("hello")

			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(targetRestConfig, nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(remoteKubeConfig, "").
				Return(targetRestConfig, nil)

			clusterClient.EXPECT().GetKubernetesCluster(ctx,
				client.ObjectKey{
					Name:      "test-cluster-name",
					Namespace: env.GetWriteNamespace(),
				}).Return(nil, testErr)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name test-cluster-name", remoteKubeConfig, localKubeConfig))

			Expect(err).To(HaveInErrorChain(testErr))
			Expect(stdout).To(ContainSubstring(register.FailedToCheckForPreviousKubeCluster))
		})

		It("will print previous command to run with --overwrite if kube cluster already exists", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			os.Setenv("KUBECONFIG", localKubeConfig)
			defer os.Setenv("KUBECONFIG", "")

			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(targetRestConfig, nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(remoteKubeConfig, "").
				Return(targetRestConfig, nil)

			clusterClient.EXPECT().GetKubernetesCluster(ctx,
				client.ObjectKey{
					Name:      "test-cluster-name",
					Namespace: env.GetWriteNamespace(),
				}).Return(nil, nil)

			command := fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name test-cluster-name", remoteKubeConfig, localKubeConfig)
			stdout, err := meshctl.Invoke(command)

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal(`Cluster already registered; if you would like to update this cluster please run the previous command with the --overwrite flag: 

$ meshctl --kubeconfig ~/.kube/master-config --remote-cluster-name test-cluster-name --remote-kubeconfig ~/.kube/target-config --overwrite
`))
		})

		It("will fail if unable install CSR agent", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			os.Setenv("KUBECONFIG", localKubeConfig)
			defer os.Setenv("KUBECONFIG", "")
			testErr := eris.New("hello")

			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(localKubeConfig, "").
				Return(targetRestConfig, nil)
			kubeLoader.
				EXPECT().
				GetRestConfigForContext(remoteKubeConfig, "").
				Return(targetRestConfig, nil)
			authClient.
				EXPECT().
				CreateAuthConfigForCluster(ctx, targetRestConfig, serviceAccountRef).
				Return(configForServiceAccount, nil)

			clusterClient.EXPECT().GetKubernetesCluster(ctx,
				client.ObjectKey{
					Name:      "test-cluster-name",
					Namespace: env.GetWriteNamespace(),
				}).Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

			csrAgentInstaller.
				EXPECT().
				Install(ctx, &csr.CsrAgentInstallOptions{
					KubeConfig:           remoteKubeConfig,
					KubeContext:          "",
					ClusterName:          "test-cluster-name",
					SmhInstallNamespace:  env.GetWriteNamespace(),
					ReleaseName:          cliconstants.CsrAgentReleaseName,
					RemoteWriteNamespace: env.GetWriteNamespace(),
				}).
				Return(testErr)

			namespaceClient.
				EXPECT().
				GetNamespace(ctx, client.ObjectKey{Name: env.GetWriteNamespace()}).
				Return(nil, nil)

			_, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name test-cluster-name", remoteKubeConfig, localKubeConfig))

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

		It("fails if KubernetesCluster resource writing fails", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			clusterName := "test-cluster-name"
			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)

			kubeLoader.EXPECT().GetRestConfigForContext(localKubeConfig, "").Return(targetRestConfig, nil)
			kubeLoader.EXPECT().GetRestConfigForContext(remoteKubeConfig, "").Return(targetRestConfig, nil)
			authClient.EXPECT().CreateAuthConfigForCluster(ctx, targetRestConfig, serviceAccountRef).Return(configForServiceAccount, nil)
			kubeLoader.EXPECT().GetRawConfigForContext(remoteKubeConfig, "").Return(cxt, nil)

			clusterClient.EXPECT().GetKubernetesCluster(ctx,
				client.ObjectKey{
					Name:      clusterName,
					Namespace: env.GetWriteNamespace(),
				}).Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

			secret := &k8s_core_types.Secret{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
					Name:      serviceAccountRef.Name,
					Namespace: env.GetWriteNamespace(),
				},
				Data: map[string][]byte{
					clusterName: []byte(expectedKubeConfig(testServerABC)),
				},
				Type: k8s_core_types.SecretTypeOpaque,
			}

			expectUpsertSecretData(ctx, secret)

			namespaceClient.
				EXPECT().
				GetNamespace(ctx, client.ObjectKey{Name: env.GetWriteNamespace()}).
				Return(nil, nil)

			testErr := eris.New("test")

			csrAgentInstaller.
				EXPECT().
				Install(ctx, &csr.CsrAgentInstallOptions{
					KubeConfig:           remoteKubeConfig,
					KubeContext:          "",
					ClusterName:          "test-cluster-name",
					SmhInstallNamespace:  env.GetWriteNamespace(),
					ReleaseName:          cliconstants.CsrAgentReleaseName,
					RemoteWriteNamespace: env.GetWriteNamespace(),
				}).
				Return(nil)

			clusterClient.EXPECT().UpsertKubernetesClusterSpec(ctx, &zephyr_discovery.KubernetesCluster{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      clusterName,
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.KubernetesClusterSpec{
					SecretRef: &zephyr_core_types.ResourceRef{
						Name:      secret.GetName(),
						Namespace: secret.GetNamespace(),
					},
					WriteNamespace: env.GetWriteNamespace(),
				},
			}).Return(testErr)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name %s", remoteKubeConfig, localKubeConfig, clusterName))

			Expect(err).To(HaveInErrorChain(register.FailedToWriteKubeCluster(testErr)))
			Expect(stdout).To(Equal(`Successfully wrote service account to remote cluster...
Successfully set up CSR agent...
Successfully wrote kube config secret to master cluster...
`))
		})

		It("can use the same kube config with different contexts", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteContext := contextDEF
			clusterName := "test-cluster-name"
			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)

			kubeLoader.EXPECT().GetRestConfigForContext(localKubeConfig, "").Return(targetRestConfig, nil)
			kubeLoader.EXPECT().GetRestConfigForContext(localKubeConfig, remoteContext).Return(targetRestConfig, nil)
			authClient.EXPECT().CreateAuthConfigForCluster(ctx, targetRestConfig, serviceAccountRef).Return(configForServiceAccount, nil)
			kubeLoader.EXPECT().GetRawConfigForContext(localKubeConfig, remoteContext).Return(cxt, nil)

			clusterClient.EXPECT().GetKubernetesCluster(ctx,
				client.ObjectKey{
					Name:      clusterName,
					Namespace: env.GetWriteNamespace(),
				}).Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

			secret := &k8s_core_types.Secret{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
					Name:      serviceAccountRef.Name,
					Namespace: env.GetWriteNamespace(),
				},
				Data: map[string][]byte{
					clusterName: []byte(expectedKubeConfig(testServerDEF)),
				},
				Type: k8s_core_types.SecretTypeOpaque,
			}

			expectUpsertSecretData(ctx, secret)

			namespaceClient.
				EXPECT().
				GetNamespace(ctx, client.ObjectKey{Name: env.GetWriteNamespace()}).
				Return(nil, nil)

			csrAgentInstaller.
				EXPECT().
				Install(ctx, &csr.CsrAgentInstallOptions{
					KubeConfig:           localKubeConfig,
					KubeContext:          remoteContext,
					ClusterName:          clusterName,
					SmhInstallNamespace:  env.GetWriteNamespace(),
					ReleaseName:          cliconstants.CsrAgentReleaseName,
					RemoteWriteNamespace: env.GetWriteNamespace(),
				})

			clusterClient.EXPECT().UpsertKubernetesClusterSpec(ctx, &zephyr_discovery.KubernetesCluster{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      clusterName,
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.KubernetesClusterSpec{
					SecretRef: &zephyr_core_types.ResourceRef{
						Name:      secret.GetName(),
						Namespace: secret.GetNamespace(),
					},
					WriteNamespace: env.GetWriteNamespace(),
				},
			}).Return(nil)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --kubeconfig %s --remote-cluster-name %s --remote-context %s", localKubeConfig, clusterName, remoteContext))

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal(`Successfully wrote service account to remote cluster...
Successfully set up CSR agent...
Successfully wrote kube config secret to master cluster...

Cluster test-cluster-name is now registered in your Service Mesh Hub installation
`))
		})

		It("can register with the CSR agent being installed from a dev chart", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			clusterName := "test-cluster-name"
			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)

			kubeLoader.EXPECT().GetRestConfigForContext(localKubeConfig, "").Return(targetRestConfig, nil)
			kubeLoader.EXPECT().GetRestConfigForContext(remoteKubeConfig, "").Return(targetRestConfig, nil)
			authClient.EXPECT().CreateAuthConfigForCluster(ctx, targetRestConfig, serviceAccountRef).Return(configForServiceAccount, nil)
			kubeLoader.EXPECT().GetRawConfigForContext(remoteKubeConfig, "").Return(cxt, nil)
			clusterClient.EXPECT().GetKubernetesCluster(ctx,
				client.ObjectKey{
					Name:      clusterName,
					Namespace: env.GetWriteNamespace(),
				}).Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

			secret := &k8s_core_types.Secret{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
					Name:      serviceAccountRef.Name,
					Namespace: env.GetWriteNamespace(),
				},
				Data: map[string][]byte{
					clusterName: []byte(expectedKubeConfig(testServerABC)),
				},
				Type: k8s_core_types.SecretTypeOpaque,
			}

			expectUpsertSecretData(ctx, secret)

			namespaceClient.
				EXPECT().
				GetNamespace(ctx, client.ObjectKey{Name: env.GetWriteNamespace()}).
				Return(nil, nil)

			csrAgentInstaller.EXPECT().
				Install(ctx, &csr.CsrAgentInstallOptions{
					KubeConfig:           remoteKubeConfig,
					KubeContext:          "",
					ClusterName:          clusterName,
					SmhInstallNamespace:  env.GetWriteNamespace(),
					RemoteWriteNamespace: env.GetWriteNamespace(),
					ReleaseName:          cliconstants.CsrAgentReleaseName,
					UseDevCsrAgentChart:  true,
				}).
				Return(nil)

			clusterClient.EXPECT().UpsertKubernetesClusterSpec(ctx, &zephyr_discovery.KubernetesCluster{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      clusterName,
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.KubernetesClusterSpec{
					SecretRef: &zephyr_core_types.ResourceRef{
						Name:      secret.GetName(),
						Namespace: secret.GetNamespace(),
					},
					WriteNamespace: env.GetWriteNamespace(),
				},
			}).Return(nil)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name %s --dev-csr-agent-chart", remoteKubeConfig, localKubeConfig, clusterName))

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal(`Successfully wrote service account to remote cluster...
Successfully set up CSR agent...
Successfully wrote kube config secret to master cluster...

Cluster test-cluster-name is now registered in your Service Mesh Hub installation
`))
		})
	})
})
