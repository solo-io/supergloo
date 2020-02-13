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
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	mock_kube "github.com/solo-io/mesh-projects/cli/pkg/common/kube/mocks"
	cli_mocks "github.com/solo-io/mesh-projects/cli/pkg/mocks"
	cluster_common "github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/common"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/register"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	mock_auth "github.com/solo-io/mesh-projects/pkg/auth/mocks"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/pkg/kubeconfig"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var _ = Describe("Cluster Operations", func() {
	var (
		ctrl           *gomock.Controller
		ctx            context.Context
		secretWriter   *cli_mocks.MockSecretWriter
		authClient     *mock_auth.MockClusterAuthorization
		kubeLoader     *cli_mocks.MockKubeLoader
		meshctl        *cli_mocks.MockMeshctl
		configVerifier *cli_mocks.MockMasterKubeConfigVerifier
		clusterClient  *mock_kube.MockKubernetesClusterClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()

		secretWriter = cli_mocks.NewMockSecretWriter(ctrl)
		authClient = mock_auth.NewMockClusterAuthorization(ctrl)
		kubeLoader = cli_mocks.NewMockKubeLoader(ctrl)
		configVerifier = cli_mocks.NewMockMasterKubeConfigVerifier(ctrl)
		clusterClient = mock_kube.NewMockKubernetesClusterClient(ctrl)
		meshctl = &cli_mocks.MockMeshctl{
			KubeClients: common.KubeClients{
				ClusterAuthorization: authClient,
				SecretWriter:         secretWriter,
				KubeClusterClient:    clusterClient,
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
			serviceAccountRef = &core.ResourceRef{
				Name:      "test-cluster-name",
				Namespace: "default",
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

		It("works", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteKubeConfig := "~/.kube/target-config"
			clusterName := "test-cluster-name"
			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)

			kubeLoader.EXPECT().GetRestConfigForContext(localKubeConfig, "").Return(targetRestConfig, nil)
			kubeLoader.EXPECT().GetRestConfigForContext(remoteKubeConfig, "").Return(targetRestConfig, nil)
			authClient.EXPECT().CreateAuthConfigForCluster(targetRestConfig, serviceAccountRef).Return(configForServiceAccount, nil)
			kubeLoader.EXPECT().GetRawConfigForContext(remoteKubeConfig, "").Return(cxt, nil)

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
					Name:      serviceAccountRef.Name,
					Namespace: env.DefaultWriteNamespace,
				},
				Data: map[string][]byte{
					clusterName: []byte(expectedKubeConfig(testServerABC)),
				},
				Type: v1.SecretTypeOpaque,
			}

			secretWriter.
				EXPECT().
				Apply(secret).
				Return(nil)

			clusterClient.EXPECT().Create(&v1alpha1.KubernetesCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterName,
					Namespace: env.DefaultWriteNamespace,
				},
				Spec: types.KubernetesClusterSpec{
					SecretRef: &types.ResourceRef{
						Name:      secret.GetName(),
						Namespace: secret.GetNamespace(),
					},
				},
			}).Return(nil)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name %s", remoteKubeConfig, localKubeConfig, clusterName))

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal(`Successfully wrote service account to remote cluster...
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
				CreateAuthConfigForCluster(targetRestConfig, serviceAccountRef).
				Return(configForServiceAccount, nil)

			kubeLoader.
				EXPECT().
				GetRawConfigForContext(remoteKubeConfig, "").
				Return(cxt, nil)

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
					Name:      serviceAccountRef.Name,
					Namespace: env.DefaultWriteNamespace,
				},
				Data: map[string][]byte{
					"test-cluster-name": []byte(expectedKubeConfig(testServerABC)),
				},
				Type: v1.SecretTypeOpaque,
			}

			secretWriter.
				EXPECT().
				Apply(secret).
				Return(nil)

			clusterClient.
				EXPECT().
				Create(&v1alpha1.KubernetesCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-name",
						Namespace: env.DefaultWriteNamespace,
					},
					Spec: types.KubernetesClusterSpec{
						SecretRef: &types.ResourceRef{
							Name:      secret.GetName(),
							Namespace: secret.GetNamespace(),
						},
					},
				}).
				Return(nil)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig "+
				"%s --remote-cluster-name test-cluster-name", remoteKubeConfig))

			Expect(stdout).To(Equal(`Successfully wrote service account to remote cluster...
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
				CreateAuthConfigForCluster(targetRestConfig, serviceAccountRef).
				Return(configForServiceAccount, nil)

			kubeLoader.
				EXPECT().
				GetRawConfigForContext(remoteKubeConfig, contextDEF).
				Return(cxt, nil)

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
					Name:      serviceAccountRef.Name,
					Namespace: env.DefaultWriteNamespace,
				},
				Data: map[string][]byte{
					"test-cluster-name": []byte(expectedKubeConfig(testServerDEF)),
				},
				Type: v1.SecretTypeOpaque,
			}

			secretWriter.
				EXPECT().
				Apply(secret).
				Return(nil)

			clusterClient.
				EXPECT().
				Create(&v1alpha1.KubernetesCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-name",
						Namespace: env.DefaultWriteNamespace,
					},
					Spec: types.KubernetesClusterSpec{
						SecretRef: &types.ResourceRef{
							Name:      secret.GetName(),
							Namespace: secret.GetNamespace(),
						},
					},
				}).
				Return(nil)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s "+
				"--remote-context %s --remote-cluster-name test-cluster-name", remoteKubeConfig, contextDEF))

			Expect(stdout).To(Equal(`Successfully wrote service account to remote cluster...
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

			Expect(err).To(HaveInErrorChain(register.FailedLoadingMasterConfig(testErr)))

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
				CreateAuthConfigForCluster(targetRestConfig, serviceAccountRef).
				Return(nil, testErr)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name test-cluster-name", remoteKubeConfig, localKubeConfig))

			Expect(err).To(Equal(testErr))
			Expect(stdout).To(ContainSubstring(register.FailedToCreateAuthToken(serviceAccountRef, remoteKubeConfig, "")))
		})

		It("will fail if unable to write secret", func() {
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
				CreateAuthConfigForCluster(targetRestConfig, serviceAccountRef).
				Return(configForServiceAccount, nil)

			kubeLoader.
				EXPECT().
				GetRawConfigForContext(remoteKubeConfig, "").
				Return(cxt, nil)

			secretWriter.
				EXPECT().
				Apply(&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
						Name:      serviceAccountRef.Name,
						Namespace: env.DefaultWriteNamespace,
					},
					Data: map[string][]byte{
						"test-cluster-name": []byte(expectedKubeConfig(testServerABC)),
					},
					Type: v1.SecretTypeOpaque,
				}).
				Return(testErr)

			output, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name test-cluster-name", remoteKubeConfig, localKubeConfig))

			Expect(output).To(ContainSubstring("Successfully wrote service account to remote cluster..."))
			Expect(err).To(HaveInErrorChain(register.FailedToWriteSecret(testErr)))
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
			Expect(err).To(HaveInErrorChain(cluster_common.NoRemoteConfigSpecifiedError))

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
			authClient.EXPECT().CreateAuthConfigForCluster(targetRestConfig, serviceAccountRef).Return(configForServiceAccount, nil)
			kubeLoader.EXPECT().GetRawConfigForContext(remoteKubeConfig, "").Return(cxt, nil)

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
					Name:      serviceAccountRef.Name,
					Namespace: env.DefaultWriteNamespace,
				},
				Data: map[string][]byte{
					clusterName: []byte(expectedKubeConfig(testServerABC)),
				},
				Type: v1.SecretTypeOpaque,
			}

			secretWriter.
				EXPECT().
				Apply(secret).
				Return(nil)

			testErr := eris.New("test")

			clusterClient.EXPECT().Create(&v1alpha1.KubernetesCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterName,
					Namespace: env.DefaultWriteNamespace,
				},
				Spec: types.KubernetesClusterSpec{
					SecretRef: &types.ResourceRef{
						Name:      secret.GetName(),
						Namespace: secret.GetNamespace(),
					},
				},
			}).Return(testErr)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --remote-kubeconfig %s"+
				" --kubeconfig %s --remote-cluster-name %s", remoteKubeConfig, localKubeConfig, clusterName))

			Expect(err).To(HaveInErrorChain(register.FailedToWriteKubeCluster(testErr)))
			Expect(stdout).To(Equal("Successfully wrote service account to remote cluster...\n"))
		})

		It("can use the same kube config with different contexts", func() {
			localKubeConfig := "~/.kube/master-config"
			remoteContext := contextDEF
			clusterName := "test-cluster-name"
			configVerifier.EXPECT().Verify(localKubeConfig, "").Return(nil)

			kubeLoader.EXPECT().GetRestConfigForContext(localKubeConfig, "").Return(targetRestConfig, nil)
			kubeLoader.EXPECT().GetRestConfigForContext(localKubeConfig, remoteContext).Return(targetRestConfig, nil)
			authClient.EXPECT().CreateAuthConfigForCluster(targetRestConfig, serviceAccountRef).Return(configForServiceAccount, nil)
			kubeLoader.EXPECT().GetRawConfigForContext(localKubeConfig, remoteContext).Return(cxt, nil)

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
					Name:      serviceAccountRef.Name,
					Namespace: env.DefaultWriteNamespace,
				},
				Data: map[string][]byte{
					clusterName: []byte(expectedKubeConfig(testServerDEF)),
				},
				Type: v1.SecretTypeOpaque,
			}

			secretWriter.
				EXPECT().
				Apply(secret).
				Return(nil)

			clusterClient.EXPECT().Create(&v1alpha1.KubernetesCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterName,
					Namespace: env.DefaultWriteNamespace,
				},
				Spec: types.KubernetesClusterSpec{
					SecretRef: &types.ResourceRef{
						Name:      secret.GetName(),
						Namespace: secret.GetNamespace(),
					},
				},
			}).Return(nil)

			stdout, err := meshctl.Invoke(fmt.Sprintf("cluster register --kubeconfig %s --remote-cluster-name %s --remote-context %s", localKubeConfig, clusterName, remoteContext))

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal(`Successfully wrote service account to remote cluster...
Successfully wrote kube config secret to master cluster...

Cluster test-cluster-name is now registered in your Service Mesh Hub installation
`))
		})
	})
})
