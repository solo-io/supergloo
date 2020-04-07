package check_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	cli_mocks "github.com/solo-io/service-mesh-hub/cli/pkg/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/status"
	mock_status "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/status/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
)

var _ = Describe("Meshctl check command", func() {
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
		Expect(nil)
	})

	It("works", func() {
		kubeLoader := cli_mocks.NewMockKubeLoader(ctrl)
		kubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(nil, nil)

		var healthCheckSuite healthcheck_types.HealthCheckSuite = map[healthcheck_types.Category][]healthcheck_types.HealthCheck{}

		statusClient := mock_status.NewMockStatusClient(ctrl)
		statusClient.EXPECT().
			Check(ctx, env.DefaultWriteNamespace, healthCheckSuite).
			Return(&status.StatusReport{
				Success: true,
			})

		meshctl := &cli_test.MockMeshctl{
			Ctx:            ctx,
			MockController: ctrl,
			Clients: common.Clients{
				StatusClientFactory: func(_ healthcheck_types.Clients) status.StatusClient {
					return statusClient
				},
				HealthCheckSuite: healthCheckSuite,
			},
			KubeLoader: kubeLoader,
		}

		// not testing stdout here, unit tests on the printers have that covered
		_, err := meshctl.Invoke("check")
		Expect(err).NotTo(HaveOccurred())
	})

	It("åccepts 'pretty' and 'json' as print formats", func() {
		kubeLoader := cli_mocks.NewMockKubeLoader(ctrl)
		kubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(nil, nil).
			Times(2)

		var healthCheckSuite healthcheck_types.HealthCheckSuite = map[healthcheck_types.Category][]healthcheck_types.HealthCheck{}

		statusClient := mock_status.NewMockStatusClient(ctrl)
		statusClient.EXPECT().
			Check(ctx, env.DefaultWriteNamespace, healthCheckSuite).
			Return(&status.StatusReport{
				Success: true,
			}).
			Times(2)

		meshctl := &cli_test.MockMeshctl{
			Ctx:            ctx,
			MockController: ctrl,
			Clients: common.Clients{
				StatusClientFactory: func(_ healthcheck_types.Clients) status.StatusClient {
					return statusClient
				},
				HealthCheckSuite: healthCheckSuite,
			},
			KubeLoader: kubeLoader,
		}

		_, err := meshctl.Invoke("check -opretty")
		Expect(err).NotTo(HaveOccurred())

		_, err = meshctl.Invoke("check -ojson")
		Expect(err).NotTo(HaveOccurred())
	})

	It("complains about unrecognized print format", func() {
		kubeLoader := cli_mocks.NewMockKubeLoader(ctrl)
		var healthCheckSuite healthcheck_types.HealthCheckSuite = map[healthcheck_types.Category][]healthcheck_types.HealthCheck{}
		statusClient := mock_status.NewMockStatusClient(ctrl)

		meshctl := &cli_test.MockMeshctl{
			Ctx:            ctx,
			MockController: ctrl,
			Clients: common.Clients{
				StatusClientFactory: func(_ healthcheck_types.Clients) status.StatusClient {
					return statusClient
				},
				HealthCheckSuite: healthCheckSuite,
			},
			KubeLoader: kubeLoader,
		}

		_, err := meshctl.Invoke("check -oabcdefg")
		Expect(err).To(testutils.HaveInErrorChain(check.UnrecognizedPrintFormat("abcdefg")))
	})
})
