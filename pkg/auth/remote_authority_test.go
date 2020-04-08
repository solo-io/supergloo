package auth_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/auth"
	mock_auth "github.com/solo-io/service-mesh-hub/pkg/auth/mocks"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core/mocks"
	kubeapiv1 "k8s.io/api/core/v1"
	rbacapiv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Remote service account client", func() {
	var (
		ctx               context.Context
		ctrl              *gomock.Controller
		serviceAccountRef = &types.ResourceRef{
			Name:      "test-sa",
			Namespace: "test-ns",
		}
		roles       = append([]*rbacapiv1.ClusterRole{}, auth.ServiceAccountRoles...)
		testErr     = eris.New("hello")
		notFoundErr = &errors.StatusError{
			ErrStatus: v1.Status{
				Reason: v1.StatusReasonAlreadyExists,
			},
		}
	)

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("works", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManager(saClient, rbacClient)

		serviceAccount := &kubeapiv1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name:      serviceAccountRef.Name,
				Namespace: serviceAccountRef.Namespace,
			},
		}

		saClient.
			EXPECT().
			Create(ctx, serviceAccount).
			Return(nil)

		rbacClient.
			EXPECT().
			BindClusterRolesToServiceAccount(serviceAccount, roles).
			Return(nil)

		sa, err := remoteAuthManager.ApplyRemoteServiceAccount(ctx, serviceAccountRef, roles)
		Expect(err).NotTo(HaveOccurred())
		Expect(sa).To(Equal(serviceAccount))
	})

	It("will try and update if create fails", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManager(saClient, rbacClient)

		serviceAccount := &kubeapiv1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name:      serviceAccountRef.Name,
				Namespace: serviceAccountRef.Namespace,
			},
		}

		saClient.
			EXPECT().
			Create(ctx, serviceAccount).
			Return(notFoundErr)

		saClient.
			EXPECT().
			Update(ctx, serviceAccount).
			Return(testErr)

		sa, err := remoteAuthManager.ApplyRemoteServiceAccount(ctx, serviceAccountRef, roles)
		Expect(err).To(HaveInErrorChain(testErr))
		Expect(sa).To(BeNil())
	})

	It("reports an error if service account creation fails, on not IsAlreadyExists error", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManager(saClient, rbacClient)

		serviceAccount := &kubeapiv1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name:      serviceAccountRef.Name,
				Namespace: serviceAccountRef.Namespace,
			},
		}

		saClient.
			EXPECT().
			Create(ctx, serviceAccount).
			Return(testErr)

		sa, err := remoteAuthManager.ApplyRemoteServiceAccount(ctx, serviceAccountRef, roles)
		Expect(err).To(HaveInErrorChain(testErr))
		Expect(sa).To(BeNil())
	})

	It("reports an error if role binding fails", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManager(saClient, rbacClient)

		serviceAccount := &kubeapiv1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name:      serviceAccountRef.Name,
				Namespace: serviceAccountRef.Namespace,
			},
		}

		saClient.
			EXPECT().
			Create(ctx, serviceAccount).
			Return(nil)

		rbacClient.
			EXPECT().
			BindClusterRolesToServiceAccount(serviceAccount, roles).
			Return(testErr)

		sa, err := remoteAuthManager.ApplyRemoteServiceAccount(ctx, serviceAccountRef, roles)
		Expect(err).To(HaveInErrorChain(testErr))
		Expect(sa).To(BeNil())
	})
})
