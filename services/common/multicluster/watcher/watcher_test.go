package mc_watcher_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/common/cluster/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster"
	mc_watcher "github.com/solo-io/service-mesh-hub/services/common/multicluster/watcher"
	mock_internal_watcher "github.com/solo-io/service-mesh-hub/services/common/multicluster/watcher/internal/mocks"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	Context("handler", func() {
		var (
			mcHandler controller.SecretEventHandler
			csh       *mock_internal_watcher.MockClusterSecretHandler
			secret    *kubev1.Secret
			testErr   = eris.New("this is an error")
		)

		BeforeEach(func() {
			csh = mock_internal_watcher.NewMockClusterSecretHandler(ctrl)
			mcHandler = mc_watcher.NewMultiClusterHandler(ctx, csh)
			secret = &kubev1.Secret{}
		})

		Context("Create", func() {
			It("will return an error if resync is true", func() {
				csh.EXPECT().AddMemberCluster(ctx, secret).Return(true, testErr)
				err := mcHandler.Create(secret)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(testErr))
			})
			It("will not an error if resync is false", func() {
				csh.EXPECT().AddMemberCluster(ctx, secret).Return(false, testErr)
				err := mcHandler.Create(secret)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Delete", func() {
			It("will return an error if resync is true", func() {
				csh.EXPECT().DeleteMemberCluster(ctx, secret).Return(true, testErr)
				err := mcHandler.Delete(secret)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(testErr))
			})
			It("will not an error if resync is false", func() {
				csh.EXPECT().DeleteMemberCluster(ctx, secret).Return(false, testErr)
				err := mcHandler.Delete(secret)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Update", func() {
			var (
				oldSecret *kubev1.Secret
			)
			Context("removed labels", func() {
				BeforeEach(func() {
					oldSecret = &kubev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: env.DefaultWriteNamespace,
							Labels:    map[string]string{multicluster.MultiClusterLabel: "true"},
						},
					}
					secret.Namespace = env.DefaultWriteNamespace
				})
				It("will return an error if resync is true", func() {
					csh.EXPECT().DeleteMemberCluster(ctx, secret).Return(true, testErr)
					err := mcHandler.Update(oldSecret, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(testErr))
				})
				It("will not an error if resync is false", func() {
					csh.EXPECT().DeleteMemberCluster(ctx, secret).Return(false, testErr)
					err := mcHandler.Update(oldSecret, secret)
					Expect(err).NotTo(HaveOccurred())
				})
			})
			Context("added labels", func() {
				BeforeEach(func() {
					oldSecret = &kubev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: env.DefaultWriteNamespace,
						},
					}
					secret.Namespace = env.DefaultWriteNamespace
					secret.Labels = map[string]string{multicluster.MultiClusterLabel: "true"}
				})
				It("will return an error if resync is true", func() {
					csh.EXPECT().AddMemberCluster(ctx, secret).Return(true, testErr)
					err := mcHandler.Update(oldSecret, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(testErr))
				})
				It("will not an error if resync is false", func() {
					csh.EXPECT().AddMemberCluster(ctx, secret).Return(false, testErr)
					err := mcHandler.Update(oldSecret, secret)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			It("will return nil if metadata hasn't changed", func() {
				oldSecret = &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: env.DefaultWriteNamespace,
						Labels:    map[string]string{multicluster.MultiClusterLabel: "true"},
					},
				}
				secret.Namespace = env.DefaultWriteNamespace
				secret.Labels = map[string]string{multicluster.MultiClusterLabel: "true"}
				err := mcHandler.Update(oldSecret, secret)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
