package internal_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/internal"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Federation decision health check", func() {
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

	It("does not consider itself valid if there are no mesh services", func() {
		meshServiceClient := mock_zephyr_discovery.NewMockMeshServiceClient(ctrl)
		meshServiceClient.EXPECT().
			List(ctx).
			Return(&v1alpha1.MeshServiceList{}, nil)

		runFailure, checkApplies := internal.NewFederationDecisionCheck().Run(ctx, env.GetWriteNamespace(), healthcheck_types.Clients{
			MeshServiceClient: meshServiceClient,
		})

		Expect(checkApplies).To(BeFalse())
		Expect(runFailure).To(BeNil())
	})

	It("does not consider itself valid if there are mesh services but they are not federated", func() {
		meshServiceClient := mock_zephyr_discovery.NewMockMeshServiceClient(ctrl)
		meshServiceClient.EXPECT().
			List(ctx).
			Return(&v1alpha1.MeshServiceList{
				Items: []v1alpha1.MeshService{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "test-1"},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "test-2"},
					},
				},
			}, nil)

		runFailure, checkApplies := internal.NewFederationDecisionCheck().Run(ctx, env.GetWriteNamespace(), healthcheck_types.Clients{
			MeshServiceClient: meshServiceClient,
		})

		Expect(checkApplies).To(BeFalse())
		Expect(runFailure).To(BeNil())
	})

	It("reports no issues with successfully federated mesh services", func() {
		meshServiceClient := mock_zephyr_discovery.NewMockMeshServiceClient(ctrl)
		meshServiceClient.EXPECT().
			List(ctx).
			Return(&v1alpha1.MeshServiceList{
				Items: []v1alpha1.MeshService{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "test-1"},
						Status: discovery_types.MeshServiceStatus{
							FederationStatus: &core_types.Status{
								State: core_types.Status_ACCEPTED,
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "test-2"},
						Status: discovery_types.MeshServiceStatus{
							FederationStatus: &core_types.Status{
								State: core_types.Status_ACCEPTED,
							},
						},
					},
				},
			}, nil)

		runFailure, checkApplies := internal.NewFederationDecisionCheck().Run(ctx, env.GetWriteNamespace(), healthcheck_types.Clients{
			MeshServiceClient: meshServiceClient,
		})

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).To(BeNil())
	})

	It("reports an issue when federation has failed to be written to a mesh service", func() {
		meshServiceClient := mock_zephyr_discovery.NewMockMeshServiceClient(ctrl)
		meshServiceClient.EXPECT().
			List(ctx).
			Return(&v1alpha1.MeshServiceList{
				Items: []v1alpha1.MeshService{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "test-1", Namespace: env.GetWriteNamespace()},
						Status: discovery_types.MeshServiceStatus{
							FederationStatus: &core_types.Status{
								State: core_types.Status_ACCEPTED,
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "test-2", Namespace: env.GetWriteNamespace()},
						Status: discovery_types.MeshServiceStatus{
							FederationStatus: &core_types.Status{
								State: core_types.Status_INVALID,
							},
						},
					},
				},
			}, nil)

		federationChecker := internal.NewFederationDecisionCheck()
		clients := healthcheck_types.Clients{
			MeshServiceClient: meshServiceClient,
		}
		runFailure, checkApplies := federationChecker.Run(ctx, env.GetWriteNamespace(), clients)

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(internal.FederationRecordingHasFailed("test-2", env.GetWriteNamespace(), core_types.Status_INVALID).Error()))
		Expect(runFailure.Hint).To(Equal(fmt.Sprintf("get details from the failing MeshService: `kubectl -n %s get meshservice %s -oyaml`", env.GetWriteNamespace(), "test-2")))
	})
})
