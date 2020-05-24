package internal_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/internal"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	mock_kubernetes_discovery "github.com/solo-io/service-mesh-hub/pkg/kube/discovery/mocks"
)

var _ = Describe("K8s connectivity check", func() {
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

		runFailure, checkApplies := internal.NewKubeConnectivityCheck().Run(ctx, container_runtime.GetWriteNamespace(), healthcheck_types.Clients{
			ServerVersionClient: serverVersionClient,
		})

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(internal.KubernetesApiServerUnreachable(testErr).Error()))
	})

	It("does not report an error if the API server is reachable", func() {
		serverVersionClient := mock_kubernetes_discovery.NewMockServerVersionClient(ctrl)
		serverVersionClient.EXPECT().
			Get().
			Return(nil, nil)

		runFailure, checkApplies := internal.NewKubeConnectivityCheck().Run(ctx, container_runtime.GetWriteNamespace(), healthcheck_types.Clients{
			ServerVersionClient: serverVersionClient,
		})

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).To(BeNil())
	})
})
