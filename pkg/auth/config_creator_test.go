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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("Config creator", func() {
	var (
		ctrl              *gomock.Controller
		serviceAccountRef = &core.ResourceRef{
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
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("works when the service account is immediately ready", func() {
		saClient := mock_auth.NewMockServiceAccountClient(ctrl)
		secretClient := mock_auth.NewMockSecretClient(ctrl)

		remoteAuthConfigCreator := auth.NewRemoteAuthorityConfigCreatorForTest(saClient, secretClient)

		saClient.
			EXPECT().
			Get(serviceAccountRef.Name, v1.GetOptions{}).
			Return(&kubeapiv1.ServiceAccount{
				Secrets: []kubeapiv1.ObjectReference{tokenSecretRef},
			}, nil)

		secretClient.
			EXPECT().
			Get(tokenSecretRef.Name, v1.GetOptions{}).
			Return(secret, nil)

		newCfg, err := remoteAuthConfigCreator.ConfigFromRemoteServiceAccount(testKubeConfig, serviceAccountRef)

		Expect(err).NotTo(HaveOccurred())
		Expect(newCfg.TLSClientConfig.CertData).To(BeEmpty())
		Expect([]byte(newCfg.BearerToken)).To(Equal(secret.Data[auth.SecretTokenKey]))
	})

	It("works when the service account eventually has a secret attached to it", func() {
		saClient := mock_auth.NewMockServiceAccountClient(ctrl)
		secretClient := mock_auth.NewMockSecretClient(ctrl)

		remoteAuthConfigCreator := auth.NewRemoteAuthorityConfigCreatorForTest(saClient, secretClient)

		attemptsRemaining := 3
		saClient.
			EXPECT().
			Get(serviceAccountRef.Name, v1.GetOptions{}).
			DoAndReturn(func(serviceAccountName string, opts v1.GetOptions) (*kubeapiv1.ServiceAccount, error) {
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
			Get(tokenSecretRef.Name, v1.GetOptions{}).
			Return(secret, nil)

		newCfg, err := remoteAuthConfigCreator.ConfigFromRemoteServiceAccount(testKubeConfig, serviceAccountRef)

		Expect(err).NotTo(HaveOccurred())
		Expect(newCfg.TLSClientConfig.CertData).To(BeEmpty())
		Expect([]byte(newCfg.BearerToken)).To(Equal(secret.Data[auth.SecretTokenKey]))
	})

	It("returns an error when the secret is malformed", func() {
		saClient := mock_auth.NewMockServiceAccountClient(ctrl)
		secretClient := mock_auth.NewMockSecretClient(ctrl)

		remoteAuthConfigCreator := auth.NewRemoteAuthorityConfigCreatorForTest(saClient, secretClient)

		saClient.
			EXPECT().
			Get(serviceAccountRef.Name, v1.GetOptions{}).
			Return(&kubeapiv1.ServiceAccount{
				Secrets: []kubeapiv1.ObjectReference{tokenSecretRef},
			}, nil)

		secretClient.
			EXPECT().
			Get(tokenSecretRef.Name, v1.GetOptions{}).
			Return(&kubeapiv1.Secret{Data: map[string][]byte{"whoops wrong key": []byte("yikes")}}, nil)

		newCfg, err := remoteAuthConfigCreator.ConfigFromRemoteServiceAccount(testKubeConfig, serviceAccountRef)

		Expect(err).To(Equal(auth.MalformedSecret))
		Expect(err).To(HaveInErrorChain(auth.MalformedSecret))
		Expect(newCfg).To(BeNil())
	})

	It("returns an error if the secret never appears", func() {
		saClient := mock_auth.NewMockServiceAccountClient(ctrl)
		secretClient := mock_auth.NewMockSecretClient(ctrl)

		remoteAuthConfigCreator := auth.NewRemoteAuthorityConfigCreatorForTest(saClient, secretClient)

		saClient.
			EXPECT().
			Get(serviceAccountRef.Name, v1.GetOptions{}).
			Return(&kubeapiv1.ServiceAccount{
				Secrets: []kubeapiv1.ObjectReference{tokenSecretRef},
			}, nil).
			AnyTimes()

		testErr := errors.New("not ready yet")

		secretClient.
			EXPECT().
			Get(tokenSecretRef.Name, v1.GetOptions{}).
			Return(nil, testErr).
			AnyTimes()

		newCfg, err := remoteAuthConfigCreator.ConfigFromRemoteServiceAccount(testKubeConfig, serviceAccountRef)

		Expect(err).To(HaveInErrorChain(auth.SecretNotReady(errors.New("test-err"))))
		Expect(newCfg).To(BeNil())
	})
})
