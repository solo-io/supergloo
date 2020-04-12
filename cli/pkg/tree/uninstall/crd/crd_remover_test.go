package crd_uninstall_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	crd_uninstall "github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/crd"
	kubernetes_apiext "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/apiext"
	mock_kubernetes_apiext "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/apiext/mocks"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("Crd Uninstaller", func() {
	var (
		ctrl       *gomock.Controller
		restConfig = &rest.Config{
			// these fields aren't relevant to anything
			Host:        "example.com",
			BearerToken: "service-account-token",
		}
		crdClientFactoryBuilder = func(crdClient kubernetes_apiext.CustomResourceDefinitionClient) kubernetes_apiext.CrdClientFactory {
			return func(cfg *rest.Config) (client kubernetes_apiext.CustomResourceDefinitionClient, err error) {
				Expect(cfg).To(Equal(restConfig))
				return crdClient, nil
			}
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("only removes zephyr CRDs", func() {
		crdClient := mock_kubernetes_apiext.NewMockCustomResourceDefinitionClient(ctrl)
		crdClient.EXPECT().
			List().
			Return(&v1beta1.CustomResourceDefinitionList{
				Items: []v1beta1.CustomResourceDefinition{
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "test.abc.zephyr.solo.io",
						},
					},
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "test.def.zephyr.solo.io",
						},
					},
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "unrelated.crd",
						},
					},
				},
			}, nil)

		crdClient.EXPECT().
			Delete("test.abc.zephyr.solo.io").
			Return(nil)

		crdClient.EXPECT().
			Delete("test.def.zephyr.solo.io").
			Return(nil)

		deletedCrds, err := crd_uninstall.NewCrdRemover(crdClientFactoryBuilder(crdClient)).RemoveZephyrCrds("cluster-1", restConfig)
		Expect(deletedCrds).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	It("responds with the appropriate error if the list call fails", func() {
		testErr := eris.New("test-err")
		crdClient := mock_kubernetes_apiext.NewMockCustomResourceDefinitionClient(ctrl)
		crdClient.EXPECT().
			List().
			Return(nil, testErr)

		removedCrds, err := crd_uninstall.NewCrdRemover(crdClientFactoryBuilder(crdClient)).RemoveZephyrCrds("cluster-1", restConfig)
		Expect(removedCrds).To(BeFalse())
		Expect(err).To(testutils.HaveInErrorChain(crd_uninstall.FailedToListCrds(testErr, "cluster-1")))
	})

	It("responds with the appropriate error if the delete call fails", func() {
		testErr := eris.New("test-err")
		crdClient := mock_kubernetes_apiext.NewMockCustomResourceDefinitionClient(ctrl)
		crdClient.EXPECT().
			List().
			Return(&v1beta1.CustomResourceDefinitionList{
				Items: []v1beta1.CustomResourceDefinition{
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "test.abc.zephyr.solo.io",
						},
					},
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "test.def.zephyr.solo.io",
						},
					},
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "unrelated.crd",
						},
					},
				},
			}, nil)

		crdClient.EXPECT().
			Delete("test.abc.zephyr.solo.io").
			Return(testErr)

		removedCrds, err := crd_uninstall.NewCrdRemover(crdClientFactoryBuilder(crdClient)).RemoveZephyrCrds("cluster-1", restConfig)
		Expect(removedCrds).To(BeTrue())
		Expect(err).To(testutils.HaveInErrorChain(crd_uninstall.FailedToDeleteCrd(testErr, "cluster-1", "test.abc.zephyr.solo.io")))
	})
})
