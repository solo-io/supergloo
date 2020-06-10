package auth_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/kube/auth"
	mock_auth "github.com/solo-io/service-mesh-hub/pkg/kube/auth/mocks"
	"k8s.io/client-go/rest"
)

var _ = Describe("Cluster authorization", func() {
	var (
		ctrl              *gomock.Controller
		ctx               context.Context
		serviceAccountRef = &smh_core_types.ResourceRef{
			Name:      "test-service-account",
			Namespace: "test-ns",
		}
		testKubeConfig = &rest.Config{
			Host: "www.grahams-a-great-programmer.edu",
			TLSClientConfig: rest.TLSClientConfig{
				CertData: []byte("super secure cert data"),
			},
		}
		serviceAccountBearerToken = "test-token"
		serviceAccountKubeConfig  = &rest.Config{
			Host:        "www.grahams-a-great-programmer.edu",
			BearerToken: serviceAccountBearerToken,
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

		outputBearerToken, err := clusterAuthClient.BuildRemoteBearerToken(ctx, testKubeConfig, serviceAccountRef)

		Expect(err).NotTo(HaveOccurred(), "An error should not have occurred")
		Expect(outputBearerToken).To(Equal(serviceAccountBearerToken), "Should have returned the expected kube config")
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

		outputBearerToken, err := clusterAuthClient.BuildRemoteBearerToken(ctx, testKubeConfig, serviceAccountRef)

		Expect(outputBearerToken).To(BeEmpty(), "Should not have created a new config")
		Expect(err).To(Equal(testErr), "Should have reported the expected error")
	})
})
