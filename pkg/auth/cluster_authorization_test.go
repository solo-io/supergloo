package auth_test

import (
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/pkg/auth"
	mock_auth "github.com/solo-io/mesh-projects/pkg/auth/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"k8s.io/client-go/rest"
)

var _ = Describe("Cluster authorization", func() {
	var (
		ctrl              *gomock.Controller
		serviceAccountRef = &core.ResourceRef{
			Name:      "test-service-account",
			Namespace: "test-ns",
		}
		testKubeConfig = &rest.Config{
			Host: "www.grahams-a-great-programmer.edu",
			TLSClientConfig: rest.TLSClientConfig{
				CertData: []byte("super secure cert data"),
			},
		}
		serviceAccountKubeConfig = &rest.Config{
			Host:        "www.grahams-a-great-programmer.edu",
			BearerToken: "test-token",
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("works when its clients work", func() {
		mockConfigCreator := mock_auth.NewMockRemoteAuthorityConfigCreator(ctrl)
		mockRemoteAuthorityManager := mock_auth.NewMockRemoteAuthorityManager(ctrl)

		clusterAuthClient := auth.NewClusterAuthorization(mockConfigCreator, mockRemoteAuthorityManager)

		mockRemoteAuthorityManager.
			EXPECT().
			CreateRemoteServiceAccount(testKubeConfig, serviceAccountRef, auth.ServiceAccountRoles).
			Return(nil, nil)

		mockConfigCreator.
			EXPECT().
			ConfigFromRemoteServiceAccount(testKubeConfig, serviceAccountRef).
			Return(serviceAccountKubeConfig, nil)

		outputConfig, err := clusterAuthClient.CreateAuthConfigForCluster(testKubeConfig, serviceAccountRef)

		Expect(err).NotTo(HaveOccurred(), "An error should not have occurred")
		Expect(outputConfig).To(Equal(serviceAccountKubeConfig), "Should have returned the expected kube config")
	})

	It("reports an error when the service account can't be created", func() {
		mockConfigCreator := mock_auth.NewMockRemoteAuthorityConfigCreator(ctrl)
		mockRemoteAuthorityManager := mock_auth.NewMockRemoteAuthorityManager(ctrl)

		clusterAuthClient := auth.NewClusterAuthorization(mockConfigCreator, mockRemoteAuthorityManager)

		testErr := errors.New("test-err")

		mockRemoteAuthorityManager.
			EXPECT().
			CreateRemoteServiceAccount(testKubeConfig, serviceAccountRef, auth.ServiceAccountRoles).
			Return(nil, testErr)

		outputConfig, err := clusterAuthClient.CreateAuthConfigForCluster(testKubeConfig, serviceAccountRef)

		Expect(outputConfig).To(BeNil(), "Should not have created a new config")
		Expect(err).To(Equal(testErr), "Should have reported the expected error")
	})
})
