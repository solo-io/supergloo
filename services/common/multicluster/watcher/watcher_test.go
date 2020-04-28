package mc_watcher_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/env"
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
			csh       *mock_internal_watcher.MockMeshPlatformSecretHandler
			secret    *kubev1.Secret
			testErr   = eris.New("this is an error")
		)

		BeforeEach(func() {
			csh = mock_internal_watcher.NewMockMeshPlatformSecretHandler(ctrl)
			mcHandler = mc_watcher.NewMeshPlatformHandler(ctx, csh)
			secret = &kubev1.Secret{}
		})

		Context("Create", func() {
			It("will return an error if resync is true", func() {
				csh.EXPECT().MeshPlatformSecretAdded(ctx, secret).Return(true, testErr)
				err := mcHandler.CreateSecret(secret)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(testErr))
			})
			It("will not an error if resync is false", func() {
				csh.EXPECT().MeshPlatformSecretAdded(ctx, secret).Return(false, testErr)
				err := mcHandler.CreateSecret(secret)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Delete", func() {
			It("will return an error if resync is true", func() {
				csh.EXPECT().MeshPlatformSecretRemoved(ctx, secret).Return(true, testErr)
				err := mcHandler.DeleteSecret(secret)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(testErr))
			})
			It("will not an error if resync is false", func() {
				csh.EXPECT().MeshPlatformSecretRemoved(ctx, secret).Return(false, testErr)
				err := mcHandler.DeleteSecret(secret)
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
							Namespace: env.GetWriteNamespace(),
							Labels:    map[string]string{multicluster.MultiClusterLabel: "true"},
						},
					}
					secret.Namespace = env.GetWriteNamespace()
				})
				It("will return an error if resync is true", func() {
					csh.EXPECT().MeshPlatformSecretRemoved(ctx, secret).Return(true, testErr)
					err := mcHandler.UpdateSecret(oldSecret, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(testErr))
				})
				It("will not an error if resync is false", func() {
					csh.EXPECT().MeshPlatformSecretRemoved(ctx, secret).Return(false, testErr)
					err := mcHandler.UpdateSecret(oldSecret, secret)
					Expect(err).NotTo(HaveOccurred())
				})
			})
			Context("added labels", func() {
				BeforeEach(func() {
					oldSecret = &kubev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: env.GetWriteNamespace(),
						},
					}
					secret.Namespace = env.GetWriteNamespace()
					secret.Labels = map[string]string{multicluster.MultiClusterLabel: "true"}
				})
				It("will return an error if resync is true", func() {
					csh.EXPECT().MeshPlatformSecretAdded(ctx, secret).Return(true, testErr)
					err := mcHandler.UpdateSecret(oldSecret, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(testErr))
				})
				It("will not an error if resync is false", func() {
					csh.EXPECT().MeshPlatformSecretAdded(ctx, secret).Return(false, testErr)
					err := mcHandler.UpdateSecret(oldSecret, secret)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			It("will return nil if metadata hasn't changed", func() {
				oldSecret = &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: env.GetWriteNamespace(),
						Labels:    map[string]string{multicluster.MultiClusterLabel: "true"},
					},
				}
				secret.Namespace = env.GetWriteNamespace()
				secret.Labels = map[string]string{multicluster.MultiClusterLabel: "true"}
				err := mcHandler.UpdateSecret(oldSecret, secret)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
