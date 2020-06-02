package auth_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/constants"
	"github.com/solo-io/service-mesh-hub/pkg/kube/auth"
	mock_auth "github.com/solo-io/service-mesh-hub/pkg/kube/auth/mocks"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_rbac_types "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Remote service account client", func() {
	var (
		ctx               context.Context
		ctrl              *gomock.Controller
		serviceAccountRef = &zephyr_core_types.ResourceRef{
			Name:      "test-sa",
			Namespace: "test-ns",
		}
		roles       = append([]*k8s_rbac_types.ClusterRole{}, auth.ServiceAccountRoles...)
		testErr     = eris.New("hello")
		notFoundErr = &errors.StatusError{
			ErrStatus: k8s_meta_types.Status{
				Reason: k8s_meta_types.StatusReasonAlreadyExists,
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

		serviceAccount := &k8s_core_types.ServiceAccount{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      serviceAccountRef.Name,
				Namespace: serviceAccountRef.Namespace,
				Labels: map[string]string{
					constants.ManagedByLabel:        constants.ServiceMeshHubApplicationName,
					auth.RegistrationServiceAccount: auth.RegistrationServiceAccountValue,
				},
			},
		}

		saClient.
			EXPECT().
			CreateServiceAccount(ctx, serviceAccount).
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

		serviceAccount := &k8s_core_types.ServiceAccount{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      serviceAccountRef.Name,
				Namespace: serviceAccountRef.Namespace,
				Labels: map[string]string{
					constants.ManagedByLabel:        constants.ServiceMeshHubApplicationName,
					auth.RegistrationServiceAccount: auth.RegistrationServiceAccountValue,
				},
			},
		}

		saClient.
			EXPECT().
			CreateServiceAccount(ctx, serviceAccount).
			Return(notFoundErr)

		saClient.
			EXPECT().
			UpdateServiceAccount(ctx, serviceAccount).
			Return(testErr)

		sa, err := remoteAuthManager.ApplyRemoteServiceAccount(ctx, serviceAccountRef, roles)
		Expect(err).To(HaveInErrorChain(testErr))
		Expect(sa).To(BeNil())
	})

	It("reports an error if service account creation fails, on not IsAlreadyExists error", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManager(saClient, rbacClient)

		serviceAccount := &k8s_core_types.ServiceAccount{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      serviceAccountRef.Name,
				Namespace: serviceAccountRef.Namespace,
				Labels: map[string]string{
					constants.ManagedByLabel:        constants.ServiceMeshHubApplicationName,
					auth.RegistrationServiceAccount: auth.RegistrationServiceAccountValue,
				},
			},
		}

		saClient.
			EXPECT().
			CreateServiceAccount(ctx, serviceAccount).
			Return(testErr)

		sa, err := remoteAuthManager.ApplyRemoteServiceAccount(ctx, serviceAccountRef, roles)
		Expect(err).To(HaveInErrorChain(testErr))
		Expect(sa).To(BeNil())
	})

	It("reports an error if role binding fails", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManager(saClient, rbacClient)

		serviceAccount := &k8s_core_types.ServiceAccount{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      serviceAccountRef.Name,
				Namespace: serviceAccountRef.Namespace,
				Labels: map[string]string{
					constants.ManagedByLabel:        constants.ServiceMeshHubApplicationName,
					auth.RegistrationServiceAccount: auth.RegistrationServiceAccountValue,
				},
			},
		}

		saClient.
			EXPECT().
			CreateServiceAccount(ctx, serviceAccount).
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
