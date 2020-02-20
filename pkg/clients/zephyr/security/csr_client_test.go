package zephyr_security_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	mock_controller_runtime "github.com/solo-io/mesh-projects/test/mocks/controller-runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("csr client", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("client impl", func() {
		var (
			mockClient       *mock_controller_runtime.MockClient
			mockStatusWriter *mock_controller_runtime.MockStatusWriter
			csrClient        zephyr_security.CertificateSigningRequestClient
		)

		BeforeEach(func() {
			mockClient = mock_controller_runtime.NewMockClient(ctrl)
			mockStatusWriter = mock_controller_runtime.NewMockStatusWriter(ctrl)
			csrClient = zephyr_security.NewCertificateSigningRequestClient(mockClient)
		})

		Context("update", func() {
			It("can call update with the proper args", func() {
				csr := &v1alpha1.MeshGroupCertificateSigningRequest{}
				mockClient.EXPECT().Update(ctx, csr).Return(testErr)
				Expect(csrClient.Update(ctx, csr)).To(Equal(testErr))
			})

			It("can call update status with the proper args", func() {
				csr := &v1alpha1.MeshGroupCertificateSigningRequest{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "namespace",
						Name:      "name",
					},
					Status: security_types.MeshGroupCertificateSigningRequestStatus{
						Response: &security_types.MeshGroupCertificateSigningResponse{},
					},
				}
				mockClient.EXPECT().Status().Return(mockStatusWriter)
				mockStatusWriter.EXPECT().Update(ctx, csr).Return(testErr)
				Expect(csrClient.UpdateStatus(ctx, &csr.Status, &csr.ObjectMeta)).To(Equal(testErr))
			})

			It("will return a nil args error if either the status or object meta are nil", func() {
				Expect(csrClient.UpdateStatus(ctx, nil, &metav1.ObjectMeta{})).
					To(Equal(zephyr_security.NilArgsError))
				Expect(csrClient.UpdateStatus(ctx, &security_types.MeshGroupCertificateSigningRequestStatus{}, nil)).
					To(Equal(zephyr_security.NilArgsError))
			})
		})

		Context("get", func() {

			It("can call get with the proper args, and return err", func() {
				csr := &v1alpha1.MeshGroupCertificateSigningRequest{}
				resourceName := client.ObjectKey{
					Name:      "name",
					Namespace: "namespace",
				}
				mockClient.EXPECT().Get(ctx, resourceName, csr).Return(testErr)
				_, err := csrClient.Get(ctx, resourceName.Name, resourceName.Namespace)
				Expect(err).To(Equal(testErr))
			})

			It("can call get with the proper args, and return non-err", func() {
				csr := &v1alpha1.MeshGroupCertificateSigningRequest{}
				resourceName := client.ObjectKey{
					Name:      "name",
					Namespace: "namespace",
				}
				mockClient.EXPECT().Get(ctx, resourceName, csr).Return(nil)
				response, err := csrClient.Get(ctx, resourceName.Name, resourceName.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
			})

		})

		Context("list", func() {

			It("can call list with the proper args, and return err", func() {
				csr := &v1alpha1.MeshGroupCertificateSigningRequestList{}
				listOptions := metav1.ListOptions{}
				mockClient.EXPECT().List(ctx, csr, &client.ListOptions{Raw: &listOptions}).Return(testErr)
				_, err := csrClient.List(ctx, listOptions)
				Expect(err).To(Equal(testErr))
			})

			It("can call get with the proper args, and return non-err", func() {
				csr := &v1alpha1.MeshGroupCertificateSigningRequestList{}
				listOptions := metav1.ListOptions{}
				mockClient.EXPECT().List(ctx, csr, &client.ListOptions{Raw: &listOptions}).Return(nil)
				response, err := csrClient.List(ctx, listOptions)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
			})
		})

	})
})
