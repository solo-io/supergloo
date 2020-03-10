package auth_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/auth"
	mock_auth "github.com/solo-io/mesh-projects/pkg/auth/mocks"
	"k8s.io/client-go/rest"
)

var _ = Describe("Cluster authorization", func() {
	var (
		ctrl              *gomock.Controller
		ctx               context.Context
		serviceAccountRef = &types.ResourceRef{
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
		ctx = context.TODO()
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
			ApplyRemoteServiceAccount(ctx, serviceAccountRef, auth.ServiceAccountRoles).
			Return(nil, nil)

		mockConfigCreator.
			EXPECT().
			ConfigFromRemoteServiceAccount(ctx, testKubeConfig, serviceAccountRef).
			Return(serviceAccountKubeConfig, nil)

		outputConfig, err := clusterAuthClient.CreateAuthConfigForCluster(ctx, testKubeConfig, serviceAccountRef)

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
			ApplyRemoteServiceAccount(ctx, serviceAccountRef, auth.ServiceAccountRoles).
			Return(nil, testErr)

		outputConfig, err := clusterAuthClient.CreateAuthConfigForCluster(ctx, testKubeConfig, serviceAccountRef)

		Expect(outputConfig).To(BeNil(), "Should not have created a new config")
		Expect(err).To(Equal(testErr), "Should have reported the expected error")
	})
})
