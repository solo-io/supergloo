package register_test

import (
	"os"

	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/pkg/kubeconfig"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	cli_mocks "github.com/solo-io/mesh-projects/cli/pkg/mocks"
	mock_auth "github.com/solo-io/mesh-projects/pkg/auth/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

var _ = Describe("Cluster Operations", func() {
	var (
		ctrl         *gomock.Controller
		secretWriter *cli_mocks.MockSecretWriter
		authClient   *mock_auth.MockClusterAuthorization
		kubeLoader   *cli_mocks.MockKubeLoader
		meshctl      *cli_mocks.MockMeshctl
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		secretWriter = cli_mocks.NewMockSecretWriter(ctrl)
		authClient = mock_auth.NewMockClusterAuthorization(ctrl)
		kubeLoader = cli_mocks.NewMockKubeLoader(ctrl)

		meshctl = &cli_mocks.MockMeshctl{
			Clients: &common.Clients{
				ClusterAuthorization: authClient,
				SecretWriter:         secretWriter,
				KubeLoader:           kubeLoader,
			},
			KubeLoader:     kubeLoader,
			MockController: ctrl,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Cluster Registration", func() {
		var (
			expectedKubeConfig = `apiVersion: v1
clusters:
- cluster:
    server: test-server
  name: test-name
contexts:
- context:
    cluster: test-name
    user: test-name
  name: test-name
current-context: test-name
kind: Config
preferences: {}
users:
- name: test-name
  user:
    token: alphanumericgarbage
`
			serviceAccountRef = &core.ResourceRef{
				Name:      "test-name",
				Namespace: "default",
			}

			targetRestConfig        = &rest.Config{Host: "www.test.com", TLSClientConfig: rest.TLSClientConfig{CertData: []byte("secret!!!")}}
			configForServiceAccount = &rest.Config{Host: "www.test.com", BearerToken: "alphanumericgarbage"}
			cxt                     = &common.KubeContext{
				CurrentContext: "contextABC",
				Contexts: map[string]*api.Context{
					"contextABC": {Cluster: "clusterABC"},
				},
				Clusters: map[string]*api.Cluster{
					"clusterABC": {Server: "test-server"},
				},
			}
		)
		It("works", func() {
			kubeLoader.
				EXPECT().
				GetRestConfig("~/.kube/target-config").
				Return(targetRestConfig, nil)

			authClient.
				EXPECT().
				CreateAuthConfigForCluster(targetRestConfig, serviceAccountRef).
				Return(configForServiceAccount, nil)

			kubeLoader.
				EXPECT().
				ParseContext("~/.kube/target-config").
				Return(cxt, nil)

			secretWriter.
				EXPECT().
				Write(&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
						Name:      serviceAccountRef.Name,
						Namespace: env.DefaultWriteNamespace,
					},
					Data: map[string][]byte{
						"test-name": []byte(expectedKubeConfig),
					},
					Type: v1.SecretTypeOpaque,
				}).
				Return(nil)

			stdout, err := meshctl.Invoke("cluster register --target-cluster-config ~/.kube/target-config --master-cluster-config ~/.kube/master-config --target-cluster-name test-name")

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal(`Successfully wrote service account to target cluster...
Successfully wrote kube config secret to master cluster...

Cluster test-name is now registered in your Service Mesh Hub installation
`))
		})

		It("works if you implicitly set master through KUBECONFIG", func() {
			os.Setenv("KUBECONFIG", "~/.kube/master-config")
			defer os.Setenv("KUBECONFIG", "")

			kubeLoader.
				EXPECT().
				GetRestConfig("~/.kube/target-config").
				Return(targetRestConfig, nil)

			authClient.
				EXPECT().
				CreateAuthConfigForCluster(targetRestConfig, serviceAccountRef).
				Return(configForServiceAccount, nil)

			kubeLoader.
				EXPECT().
				ParseContext("~/.kube/target-config").
				Return(cxt, nil)

			secretWriter.
				EXPECT().
				Write(&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Labels:    map[string]string{kubeconfig.KubeConfigSecretLabel: "true"},
						Name:      serviceAccountRef.Name,
						Namespace: env.DefaultWriteNamespace,
					},
					Data: map[string][]byte{
						"test-name": []byte(expectedKubeConfig),
					},
					Type: v1.SecretTypeOpaque,
				}).
				Return(nil)

			stdout, err := meshctl.Invoke("cluster register --target-cluster-config ~/.kube/target-config --target-cluster-name test-name")

			Expect(stdout).To(Equal(`Successfully wrote service account to target cluster...
Successfully wrote kube config secret to master cluster...

Cluster test-name is now registered in your Service Mesh Hub installation
`))
			Expect(err).NotTo(HaveOccurred())
		})

		It("errors if a master or target cluster are not set", func() {
			os.Setenv("KUBECONFIG", "")
			stdout, err := meshctl.Invoke("cluster register --target-cluster-config ~/.kube/target-config --target-cluster-name test-name")
			Expect(stdout).To(BeEmpty())
			Expect(err.Error()).To(ContainSubstring("required flag(s) \"master-cluster-config\" not set"))

			stdout, err = meshctl.Invoke("cluster register --master-cluster-config ~/.kube/master-config --target-cluster-name test-name")
			Expect(stdout).To(BeEmpty())
			Expect(err.Error()).To(ContainSubstring("\"target-cluster-config\" not set"))
		})
	})
})
