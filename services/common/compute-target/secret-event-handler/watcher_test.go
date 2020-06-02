package mc_watcher_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	mc_watcher "github.com/solo-io/service-mesh-hub/services/common/compute-target/secret-event-handler"
	mock_internal_watcher "github.com/solo-io/service-mesh-hub/services/common/compute-target/secret-event-handler/internal/mocks"
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
			csh       *mock_internal_watcher.MockComputeTargetSecretHandler
			secret    *kubev1.Secret
			testErr   = eris.New("this is an error")
		)

		BeforeEach(func() {
			csh = mock_internal_watcher.NewMockComputeTargetSecretHandler(ctrl)
			mcHandler = mc_watcher.NewComputeTargetHandler(ctx, csh)
			secret = &kubev1.Secret{}
		})

		Context("Create", func() {
			It("will return an error if resync is true", func() {
				csh.EXPECT().ComputeTargetSecretAdded(ctx, secret).Return(true, testErr)
				err := mcHandler.CreateSecret(secret)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(testErr))
			})
			It("will not an error if resync is false", func() {
				csh.EXPECT().ComputeTargetSecretAdded(ctx, secret).Return(false, testErr)
				err := mcHandler.CreateSecret(secret)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Delete", func() {
			It("will return an error if resync is true", func() {
				csh.EXPECT().ComputeTargetSecretRemoved(ctx, secret).Return(true, testErr)
				err := mcHandler.DeleteSecret(secret)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(testErr))
			})
			It("will not an error if resync is false", func() {
				csh.EXPECT().ComputeTargetSecretRemoved(ctx, secret).Return(false, testErr)
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
							Namespace: container_runtime.GetWriteNamespace(),
							Labels:    map[string]string{mc_manager.MultiClusterLabel: "true"},
						},
					}
					secret.Namespace = container_runtime.GetWriteNamespace()
				})
				It("will return an error if resync is true", func() {
					csh.EXPECT().ComputeTargetSecretRemoved(ctx, secret).Return(true, testErr)
					err := mcHandler.UpdateSecret(oldSecret, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(testErr))
				})
				It("will not an error if resync is false", func() {
					csh.EXPECT().ComputeTargetSecretRemoved(ctx, secret).Return(false, testErr)
					err := mcHandler.UpdateSecret(oldSecret, secret)
					Expect(err).NotTo(HaveOccurred())
				})
			})
			Context("added labels", func() {
				BeforeEach(func() {
					oldSecret = &kubev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: container_runtime.GetWriteNamespace(),
						},
					}
					secret.Namespace = container_runtime.GetWriteNamespace()
					secret.Labels = map[string]string{mc_manager.MultiClusterLabel: "true"}
				})
				It("will return an error if resync is true", func() {
					csh.EXPECT().ComputeTargetSecretAdded(ctx, secret).Return(true, testErr)
					err := mcHandler.UpdateSecret(oldSecret, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(testErr))
				})
				It("will not an error if resync is false", func() {
					csh.EXPECT().ComputeTargetSecretAdded(ctx, secret).Return(false, testErr)
					err := mcHandler.UpdateSecret(oldSecret, secret)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			It("will return nil if metadata hasn't changed", func() {
				oldSecret = &kubev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: container_runtime.GetWriteNamespace(),
						Labels:    map[string]string{mc_manager.MultiClusterLabel: "true"},
					},
				}
				secret.Namespace = container_runtime.GetWriteNamespace()
				secret.Labels = map[string]string{mc_manager.MultiClusterLabel: "true"}
				err := mcHandler.UpdateSecret(oldSecret, secret)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
