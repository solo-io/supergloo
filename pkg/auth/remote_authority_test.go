package auth_test

import (
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/pkg/auth"
	mock_auth "github.com/solo-io/mesh-projects/pkg/auth/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	kubeapiv1 "k8s.io/api/core/v1"
	rbacapiv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("Remote service account client", func() {
	var (
		ctrl              *gomock.Controller
		serviceAccountRef = &core.ResourceRef{
			Name:      "test-sa",
			Namespace: "test-ns",
		}
		testKubeConfig = &rest.Config{
			Host: "www.grahams-a-great-programmer.edu",
			TLSClientConfig: rest.TLSClientConfig{
				CertData: []byte("super secure cert data"),
			},
		}
		roles = append([]*rbacapiv1.ClusterRole{}, auth.ServiceAccountRoles...)
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

		remoteAuthManager := auth.NewRemoteAuthorityManager(mock_auth.MockClients(saClient, rbacClient, nil))

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

		sa, err := remoteAuthManager.CreateRemoteServiceAccount(testKubeConfig, serviceAccountRef, roles)
		Expect(err).NotTo(HaveOccurred())
		Expect(sa).To(Equal(serviceAccount))
	})

	It("reports an error if service account creation fails", func() {
		saClient := mock_auth.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManager(mock_auth.MockClients(saClient, rbacClient, nil))

		serviceAccount := &kubeapiv1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name: serviceAccountRef.Name,
			},
		}

		testErr := errors.New("test err")

		saClient.
			EXPECT().
			Create(serviceAccount).
			Return(nil, testErr)

		sa, err := remoteAuthManager.CreateRemoteServiceAccount(testKubeConfig, serviceAccountRef, roles)
		Expect(err).To(HaveInErrorChain(testErr))
		Expect(sa).To(BeNil())
	})

	It("reports an error if role binding fails", func() {
		saClient := mock_auth.NewMockServiceAccountClient(ctrl)
		rbacClient := mock_auth.NewMockRbacClient(ctrl)

		remoteAuthManager := auth.NewRemoteAuthorityManager(mock_auth.MockClients(saClient, rbacClient, nil))

		serviceAccount := &kubeapiv1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name: serviceAccountRef.Name,
			},
		}

		testErr := errors.New("test err")

		saClient.
			EXPECT().
			Create(serviceAccount).
			Return(serviceAccount, nil)

		rbacClient.
			EXPECT().
			BindClusterRolesToServiceAccount(serviceAccount, roles).
			Return(testErr)

		sa, err := remoteAuthManager.CreateRemoteServiceAccount(testKubeConfig, serviceAccountRef, roles)
		Expect(err).To(HaveInErrorChain(testErr))
		Expect(sa).To(BeNil())
	})
})
