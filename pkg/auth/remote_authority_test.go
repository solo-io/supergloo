package auth_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/pkg/auth"
	mock_auth "github.com/solo-io/mesh-projects/pkg/auth/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	kubeapiv1 "k8s.io/api/core/v1"
	rbacapiv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Remote service account client", func() {
	var (
		ctrl              *gomock.Controller
		serviceAccountRef = &core.ResourceRef{
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
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("works", func() {
		saClient := mock_auth.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManagerForTest(saClient, rbacClient)

		serviceAccount := &kubeapiv1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name: serviceAccountRef.Name,
			},
		}

		saClient.
			EXPECT().
			Create(serviceAccount).
			Return(serviceAccount, nil)

		rbacClient.
			EXPECT().
			BindClusterRolesToServiceAccount(serviceAccount, roles).
			Return(nil)

		sa, err := remoteAuthManager.ApplyRemoteServiceAccount(serviceAccountRef, roles)
		Expect(err).NotTo(HaveOccurred())
		Expect(sa).To(Equal(serviceAccount))
	})

	It("will try and update if create fails", func() {
		saClient := mock_auth.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManagerForTest(saClient, rbacClient)

		serviceAccount := &kubeapiv1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name: serviceAccountRef.Name,
			},
		}

		saClient.
			EXPECT().
			Create(serviceAccount).
			Return(nil, notFoundErr)

		saClient.
			EXPECT().
			Update(serviceAccount).
			Return(nil, testErr)

		sa, err := remoteAuthManager.ApplyRemoteServiceAccount(serviceAccountRef, roles)
		Expect(err).To(HaveInErrorChain(testErr))
		Expect(sa).To(BeNil())
	})

	It("reports an error if service account creation fails, on not IsAlreadyExists error", func() {
		saClient := mock_auth.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManagerForTest(saClient, rbacClient)

		serviceAccount := &kubeapiv1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name: serviceAccountRef.Name,
			},
		}

		saClient.
			EXPECT().
			Create(serviceAccount).
			Return(nil, testErr)

		sa, err := remoteAuthManager.ApplyRemoteServiceAccount(serviceAccountRef, roles)
		Expect(err).To(HaveInErrorChain(testErr))
		Expect(sa).To(BeNil())
	})

	It("reports an error if role binding fails", func() {
		saClient := mock_auth.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManagerForTest(saClient, rbacClient)

		serviceAccount := &kubeapiv1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name: serviceAccountRef.Name,
			},
		}

		saClient.
			EXPECT().
			Create(serviceAccount).
			Return(serviceAccount, nil)

		rbacClient.
			EXPECT().
			BindClusterRolesToServiceAccount(serviceAccount, roles).
			Return(testErr)

		sa, err := remoteAuthManager.ApplyRemoteServiceAccount(serviceAccountRef, roles)
		Expect(err).To(HaveInErrorChain(testErr))
		Expect(sa).To(BeNil())
	})
})
