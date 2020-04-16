package auth_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/auth"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core/mocks"
	kubeapiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("Config creator", func() {
	var (
		ctx               context.Context
		ctrl              *gomock.Controller
		serviceAccountRef = &types.ResourceRef{
			Name:      "test-sa",
			Namespace: "test-ns",
		}
		tokenSecretRef = kubeapiv1.ObjectReference{
			Name: "test-secret",
		}
		secret = &kubeapiv1.Secret{
			Data: map[string][]byte{
				auth.SecretTokenKey: []byte("my-test-token"),
			},
		}
		testKubeConfig = &rest.Config{
			Host: "www.grahams-a-great-programmer.edu",
			TLSClientConfig: rest.TLSClientConfig{
				CertData: []byte("super secure cert data"),
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

	It("works when the service account is immediately ready", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		secretClient := mock_kubernetes_core.NewMockSecretClient(ctrl)

		remoteAuthConfigCreator := auth.NewRemoteAuthorityConfigCreator(secretClient, saClient)

		saClient.
			EXPECT().
			Get(ctx, serviceAccountRef.Name, serviceAccountRef.Namespace).
			Return(&kubeapiv1.ServiceAccount{
				Secrets: []kubeapiv1.ObjectReference{tokenSecretRef},
			}, nil)

		secretClient.
			EXPECT().
			Get(ctx, tokenSecretRef.Name, serviceAccountRef.Namespace).
			Return(secret, nil)

		newCfg, err := remoteAuthConfigCreator.ConfigFromRemoteServiceAccount(ctx, testKubeConfig, serviceAccountRef)

		Expect(err).NotTo(HaveOccurred())
		Expect(newCfg.TLSClientConfig.CertData).To(BeEmpty())
		Expect([]byte(newCfg.BearerToken)).To(Equal(secret.Data[auth.SecretTokenKey]))
	})

	It("works when the service account eventually has a secret attached to it", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		secretClient := mock_kubernetes_core.NewMockSecretClient(ctrl)

		remoteAuthConfigCreator := auth.NewRemoteAuthorityConfigCreator(secretClient, saClient)

		attemptsRemaining := 3
		saClient.
			EXPECT().
			Get(ctx, serviceAccountRef.Name, serviceAccountRef.Namespace).
			DoAndReturn(func(ctx context.Context, serviceAccountName, namespace string) (*kubeapiv1.ServiceAccount, error) {
				attemptsRemaining -= 1
				if attemptsRemaining > 0 {
					return nil, errors.New("whoops not ready yet")
				}

				return &kubeapiv1.ServiceAccount{
					Secrets: []kubeapiv1.ObjectReference{tokenSecretRef},
				}, nil
			}).
			AnyTimes()

		secretClient.
			EXPECT().
			Get(ctx, tokenSecretRef.Name, serviceAccountRef.Namespace).
			Return(secret, nil)

		newCfg, err := remoteAuthConfigCreator.ConfigFromRemoteServiceAccount(ctx, testKubeConfig, serviceAccountRef)

		Expect(err).NotTo(HaveOccurred())
		Expect(newCfg.TLSClientConfig.CertData).To(BeEmpty())
		Expect([]byte(newCfg.BearerToken)).To(Equal(secret.Data[auth.SecretTokenKey]))
	})

	It("returns an error when the secret is malformed", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		secretClient := mock_kubernetes_core.NewMockSecretClient(ctrl)

		remoteAuthConfigCreator := auth.NewRemoteAuthorityConfigCreator(secretClient, saClient)

		saClient.
			EXPECT().
			Get(ctx, serviceAccountRef.Name, serviceAccountRef.Namespace).
			Return(&kubeapiv1.ServiceAccount{
				Secrets: []kubeapiv1.ObjectReference{tokenSecretRef},
			}, nil)

		secretClient.
			EXPECT().
			Get(ctx, tokenSecretRef.Name, serviceAccountRef.Namespace).
			Return(&kubeapiv1.Secret{Data: map[string][]byte{"whoops wrong key": []byte("yikes")}}, nil)

		newCfg, err := remoteAuthConfigCreator.ConfigFromRemoteServiceAccount(ctx, testKubeConfig, serviceAccountRef)

		Expect(err).To(Equal(auth.MalformedSecret))
		Expect(err).To(HaveInErrorChain(auth.MalformedSecret))
		Expect(newCfg).To(BeNil())
	})

	It("returns an error if the secret never appears", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		secretClient := mock_kubernetes_core.NewMockSecretClient(ctrl)

		remoteAuthConfigCreator := auth.NewRemoteAuthorityConfigCreator(secretClient, saClient)

		saClient.
			EXPECT().
			Get(ctx, serviceAccountRef.Name, serviceAccountRef.Namespace).
			Return(&kubeapiv1.ServiceAccount{
				Secrets: []kubeapiv1.ObjectReference{tokenSecretRef},
			}, nil).
			AnyTimes()

		testErr := errors.New("not ready yet")

		secretClient.
			EXPECT().
			Get(ctx, tokenSecretRef.Name, serviceAccountRef.Namespace).
			Return(nil, testErr).
			AnyTimes()

		newCfg, err := remoteAuthConfigCreator.ConfigFromRemoteServiceAccount(ctx, testKubeConfig, serviceAccountRef)

		Expect(err).To(HaveInErrorChain(auth.SecretNotReady(errors.New("test-err"))))
		Expect(newCfg).To(BeNil())
	})
})
