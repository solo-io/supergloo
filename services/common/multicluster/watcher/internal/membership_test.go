package internal_watcher_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/kubeutils"
	. "github.com/solo-io/go-utils/testutils"
	mock_mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager/mocks"
	. "github.com/solo-io/mesh-projects/services/common/multicluster/watcher/internal"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var _ = Describe("multicluster-watcher", func() {

	var (
		ctrl *gomock.Controller
		ctx  context.Context
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
			receiver *mock_mc_manager.MockKubeConfigHandler
			cmh      *ClusterMembershipHandler
			cfg      *clientcmdapi.Config

			byteConfig              []byte
			clusterName, secretName = "cluster-name", "secret-name"
		)

		BeforeEach(func() {
			receiver = mock_mc_manager.NewMockKubeConfigHandler(ctrl)
			cmh = NewClusterMembershipHandler(receiver)
			var err error
			cfg, err = kubeutils.GetKubeConfig("", "")
			Expect(err).NotTo(HaveOccurred())

			byteConfig, err = clientcmd.Write(*cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("add cluster", func() {
			It("returns nil if data is nil", func() {
				resync, err := cmh.AddMemberCluster(ctx, &kubev1.Secret{})
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error if there is an invalid kube config string", func() {
				resync, err := cmh.AddMemberCluster(ctx, &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: []byte("failing config"),
					},
				})
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).To(HaveOccurred())
				Expect(err).To(HaveInErrorChain(KubeConfigInvalidFormatError(eris.New("hello"),
					clusterName, secretName, "")))
			})

			It("returns an error if the receiver returns an error", func() {
				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(eris.New("this is an error"))
				resync, err := cmh.AddMemberCluster(ctx, &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				})
				Expect(resync).To(BeTrue(), "resync should be true")
				Expect(err).To(HaveOccurred())
				Expect(err).To(HaveInErrorChain(ClusterAddError(eris.New("hello"), clusterName)))
			})

			It("can successfully add a cluster", func() {
				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(nil)
				resync, err := cmh.AddMemberCluster(ctx, &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				})
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())
			})

			It("cluster exists error", func() {
				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(nil)
				resync, err := cmh.AddMemberCluster(ctx, &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				})
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())
				resync, err = cmh.AddMemberCluster(ctx, &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				})
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).To(HaveOccurred())
				Expect(err).To(HaveInErrorChain(ClusterExistsError(clusterName, secretName, "")))
			})
		})

		Context("delete cluster", func() {
			It("deleting a non-existent cluster will do nothing", func() {
				resync, err := cmh.DeleteMemberCluster(ctx, &kubev1.Secret{})
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())
			})

			It("will return an error if the receiver is called and errors", func() {
				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(nil)
				resync, err := cmh.AddMemberCluster(ctx, &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				})
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
				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(nil)
				resync, err := cmh.AddMemberCluster(ctx, &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				})
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
