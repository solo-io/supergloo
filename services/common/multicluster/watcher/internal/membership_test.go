package internal_watcher_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	mock_kube "github.com/solo-io/service-mesh-hub/cli/pkg/common/kube/mocks"
	mock_mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/mocks"
	. "github.com/solo-io/service-mesh-hub/services/common/multicluster/watcher/internal"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

var _ = Describe("multicluster-watcher", func() {

	var (
		ctrl       *gomock.Controller
		ctx        context.Context
		restConfig = &rest.Config{}

		byteConfig = []byte(`
apiVersion: v1
clusters:
- cluster:
    server: https://localhost:9090
  name: k3s-default
contexts:
- context:
    cluster: k3s-default
    user: k3s-default
  name: k3s-default
current-context: k3s-default
kind: Config
preferences: {}
users:
- name: k3s-default
  user:
    password: admin
    username: admin

`)
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("cluster membership", func() {

		var (
			receiver      *mock_mc_manager.MockKubeConfigHandler
			kubeConverter *mock_kube.MockConverter
			cmh           *MeshAPIMembershipHandler

			clusterName, secretName = "cluster-name", "secret-name"
		)

		BeforeEach(func() {
			receiver = mock_mc_manager.NewMockKubeConfigHandler(ctrl)
			kubeConverter = mock_kube.NewMockConverter(ctrl)
			cmh = NewClusterMembershipHandler(receiver, kubeConverter)
		})

		Context("add cluster", func() {
			It("returns an error if the secret is malformed", func() {
				secret := &kubev1.Secret{}
				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return("", nil, kube.NoDataInKubeConfigSecret(&kubev1.Secret{}))

				resync, err := cmh.AddMemberMeshAPI(ctx, secret)
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).To(HaveInErrorChain(kube.NoDataInKubeConfigSecret(&kubev1.Secret{})))
			})

			It("returns an error if there is an invalid kube config string", func() {
				secret := &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: []byte("failing config"),
					},
				}

				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return("", nil, KubeConfigInvalidFormatError(eris.New("hello"), clusterName, secretName, ""))

				resync, err := cmh.AddMemberMeshAPI(ctx, secret)
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).To(HaveOccurred())
				Expect(err).To(HaveInErrorChain(KubeConfigInvalidFormatError(eris.New("hello"),
					clusterName, secretName, "")))
			})

			It("returns an error if the receiver returns an error", func() {
				secret := &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}

				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return(clusterName, &kube.ConvertedConfigs{
						RestConfig: restConfig,
					}, nil)

				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(eris.New("this is an error"))
				resync, err := cmh.AddMemberMeshAPI(ctx, secret)
				Expect(resync).To(BeTrue(), "resync should be true")
				Expect(err).To(HaveOccurred())
				Expect(err).To(HaveInErrorChain(ClusterAddError(eris.New("hello"), clusterName)))
			})

			It("can successfully add a cluster", func() {
				secret := &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}

				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return(clusterName, &kube.ConvertedConfigs{
						RestConfig: restConfig,
						ApiConfig: &api.Config{
							CurrentContext: "current-context",
						},
					}, nil)

				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(nil)
				resync, err := cmh.AddMemberMeshAPI(ctx, secret)
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("delete cluster", func() {
			It("deleting a non-existent cluster will do nothing", func() {
				resync, err := cmh.DeleteMemberCluster(ctx, &kubev1.Secret{})
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())
			})

			It("will return an error if the receiver is called and errors", func() {
				secret := &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}
				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return(clusterName, &kube.ConvertedConfigs{
						RestConfig: restConfig,
						ApiConfig: &api.Config{
							CurrentContext: "current-context",
						},
					}, nil)

				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(nil)
				resync, err := cmh.AddMemberMeshAPI(ctx, secret)
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())

				receiver.EXPECT().ClusterRemoved(clusterName).Return(eris.New("this is an error"))
				resync, err = cmh.DeleteMemberCluster(ctx, &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
				})
				Expect(resync).To(BeTrue(), "resync should be true")
				Expect(err).To(HaveOccurred())
				Expect(err).To(HaveInErrorChain(ClusterDeletionError(eris.New("hello"), clusterName)))
			})

			It("will return nil and delete cluster if return is nil", func() {
				secret := &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}
				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return(clusterName, &kube.ConvertedConfigs{
						RestConfig: restConfig,
						ApiConfig: &api.Config{
							CurrentContext: "current-context",
						},
					}, nil)

				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(nil)
				resync, err := cmh.AddMemberMeshAPI(ctx, secret)
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())

				receiver.EXPECT().ClusterRemoved(clusterName).Return(nil)
				resync, err = cmh.DeleteMemberCluster(ctx, &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
				})
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
