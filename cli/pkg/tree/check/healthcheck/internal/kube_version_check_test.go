package internal_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/internal"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	version2 "github.com/solo-io/service-mesh-hub/pkg/container-runtime/version"
	mock_kubernetes_discovery "github.com/solo-io/service-mesh-hub/pkg/kube/discovery/mocks"
	"k8s.io/apimachinery/pkg/version"
)

var _ = Describe("K8s version check", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("reports an error if the API server is not reachable", func() {
		testErr := eris.New("test-err")
		serverVersionClient := mock_kubernetes_discovery.NewMockServerVersionClient(ctrl)
		serverVersionClient.EXPECT().
			Get().
			Return(nil, testErr)

		runFailure, checkApplies := internal.NewK8sServerVersionCheck().Run(ctx, container_runtime.GetWriteNamespace(), healthcheck_types.Clients{
			ServerVersionClient: serverVersionClient,
		})

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(internal.KubernetesApiServerUnreachable(testErr).Error()))
	})

	It("reports an error if the minor version is unrecognized", func() {
		serverVersionClient := mock_kubernetes_discovery.NewMockServerVersionClient(ctrl)
		serverVersionClient.EXPECT().
			Get().
			Return(&version.Info{
				Minor: "abcd",
			}, nil)

		runFailure, checkApplies := internal.NewK8sServerVersionCheck().Run(ctx, container_runtime.GetWriteNamespace(), healthcheck_types.Clients{
			ServerVersionClient: serverVersionClient,
		})

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(internal.KubernetesServerVersionUnsupported("abcd").Error()))
	})

	It("can recognize and accept slightly nonstandard minor version numbers (eg from GKE)", func() {
		serverVersionClient := mock_kubernetes_discovery.NewMockServerVersionClient(ctrl)
		serverVersionClient.EXPECT().
			Get().
			Return(&version.Info{
				Minor: fmt.Sprintf("%d+", version2.MinimumSupportedKubernetesMinorVersion), // for example: "15+"
			}, nil)

		runFailure, checkApplies := internal.NewK8sServerVersionCheck().Run(ctx, container_runtime.GetWriteNamespace(), healthcheck_types.Clients{
			ServerVersionClient: serverVersionClient,
		})

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).To(BeNil())
	})

	It("can recognize and accept a minor version ahead of the minimum supported version", func() {
		serverVersionClient := mock_kubernetes_discovery.NewMockServerVersionClient(ctrl)
		serverVersionClient.EXPECT().
			Get().
			Return(&version.Info{
				Minor: fmt.Sprintf("%d", version2.MinimumSupportedKubernetesMinorVersion+2), // for example: "15+"
			}, nil)

		runFailure, checkApplies := internal.NewK8sServerVersionCheck().Run(ctx, container_runtime.GetWriteNamespace(), healthcheck_types.Clients{
			ServerVersionClient: serverVersionClient,
		})

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).To(BeNil())
	})
})
