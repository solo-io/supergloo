package auth_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/auth"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Config creator", func() {
	var (
		ctx               context.Context
		ctrl              *gomock.Controller
		serviceAccountRef = &zephyr_core_types.ResourceRef{
			Name:      "test-sa",
			Namespace: "test-ns",
		}
		tokenSecretRef = k8s_core_types.ObjectReference{
			Name: "test-secret",
		}
		secret = &k8s_core_types.Secret{
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
			GetServiceAccount(ctx, client.ObjectKey{Name: serviceAccountRef.Name, Namespace: serviceAccountRef.Namespace}).
			Return(&k8s_core_types.ServiceAccount{
				Secrets: []k8s_core_types.ObjectReference{tokenSecretRef},
			}, nil)

		secretClient.
			EXPECT().
			GetSecret(ctx, client.ObjectKey{Name: tokenSecretRef.Name, Namespace: serviceAccountRef.Namespace}).
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
			GetServiceAccount(ctx, client.ObjectKey{Name: serviceAccountRef.Name, Namespace: serviceAccountRef.Namespace}).
			DoAndReturn(func(ctx context.Context, key client.ObjectKey) (*k8s_core_types.ServiceAccount, error) {
				attemptsRemaining -= 1
				if attemptsRemaining > 0 {
					return nil, errors.New("whoops not ready yet")
				}

				return &k8s_core_types.ServiceAccount{
					Secrets: []k8s_core_types.ObjectReference{tokenSecretRef},
				}, nil
			}).
			AnyTimes()

		secretClient.
			EXPECT().
			GetSecret(ctx, client.ObjectKey{Name: tokenSecretRef.Name, Namespace: serviceAccountRef.Namespace}).
			Return(secret, nil)

		newCfg, err := remoteAuthConfigCreator.ConfigFromRemoteServiceAccount(ctx, testKubeConfig, serviceAccountRef)

		Expect(err).NotTo(HaveOccurred())
		Expect(newCfg.TLSClientConfig.CertData).To(BeEmpty())
		Expect([]byte(newCfg.BearerToken)).To(Equal(secret.Data[auth.SecretTokenKey]))
	})

	It("works when the service account is immediately ready, and the CA data is in a file", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		secretClient := mock_kubernetes_core.NewMockSecretClient(ctrl)

		fileTestKubeConfig := &rest.Config{
			Host: "www.grahams-a-great-programmer.edu",
			TLSClientConfig: rest.TLSClientConfig{
				CAFile:   "path-to-ca-file",
				CertData: []byte("super secure cert data"),
			},
		}

		remoteAuthConfigCreator := auth.NewRemoteAuthorityConfigCreator(secretClient, saClient)

		saClient.
			EXPECT().
			GetServiceAccount(ctx, client.ObjectKey{Name: serviceAccountRef.Name, Namespace: serviceAccountRef.Namespace}).
			Return(&k8s_core_types.ServiceAccount{
				Secrets: []k8s_core_types.ObjectReference{tokenSecretRef},
			}, nil)

		secretClient.
			EXPECT().
			GetSecret(ctx, client.ObjectKey{Name: tokenSecretRef.Name, Namespace: serviceAccountRef.Namespace}).
			Return(secret, nil)

		newCfg, err := remoteAuthConfigCreator.ConfigFromRemoteServiceAccount(ctx, fileTestKubeConfig, serviceAccountRef)

		Expect(err).NotTo(HaveOccurred())
		Expect(newCfg.TLSClientConfig.CertData).To(BeEmpty())
		Expect(newCfg.TLSClientConfig.CAData).To(BeEmpty())
		Expect(newCfg.TLSClientConfig.CAFile).To(Equal(fileTestKubeConfig.TLSClientConfig.CAFile))
		Expect([]byte(newCfg.BearerToken)).To(Equal(secret.Data[auth.SecretTokenKey]))
	})

	It("returns an error when the secret is malformed", func() {
		saClient := mock_kubernetes_core.NewMockServiceAccountClient(ctrl)
		secretClient := mock_kubernetes_core.NewMockSecretClient(ctrl)

		remoteAuthConfigCreator := auth.NewRemoteAuthorityConfigCreator(secretClient, saClient)

		saClient.
			EXPECT().
			GetServiceAccount(ctx, client.ObjectKey{Name: serviceAccountRef.Name, Namespace: serviceAccountRef.Namespace}).
			Return(&k8s_core_types.ServiceAccount{
				Secrets: []k8s_core_types.ObjectReference{tokenSecretRef},
			}, nil)

		secretClient.
			EXPECT().
			GetSecret(ctx, client.ObjectKey{Name: tokenSecretRef.Name, Namespace: serviceAccountRef.Namespace}).
			Return(&k8s_core_types.Secret{Data: map[string][]byte{"whoops wrong key": []byte("yikes")}}, nil)

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
			GetServiceAccount(ctx, client.ObjectKey{Name: serviceAccountRef.Name, Namespace: serviceAccountRef.Namespace}).
			Return(&k8s_core_types.ServiceAccount{
				Secrets: []k8s_core_types.ObjectReference{tokenSecretRef},
			}, nil).
			AnyTimes()

		testErr := errors.New("not ready yet")

		secretClient.
			EXPECT().
			GetSecret(ctx, client.ObjectKey{Name: tokenSecretRef.Name, Namespace: serviceAccountRef.Namespace}).
			Return(nil, testErr).
			AnyTimes()

		newCfg, err := remoteAuthConfigCreator.ConfigFromRemoteServiceAccount(ctx, testKubeConfig, serviceAccountRef)

		Expect(err).To(HaveInErrorChain(auth.SecretNotReady(errors.New("test-err"))))
		Expect(newCfg).To(BeNil())
	})
})
