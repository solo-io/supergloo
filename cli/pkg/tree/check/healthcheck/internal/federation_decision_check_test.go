package internal_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/internal"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	mock_smh_discovery "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		meshServiceClient := mock_smh_discovery.NewMockMeshServiceClient(ctrl)
		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&smh_discovery.MeshServiceList{}, nil)

		runFailure, checkApplies := internal.NewFederationDecisionCheck().Run(ctx, container_runtime.GetWriteNamespace(), healthcheck_types.Clients{
			MeshServiceClient: meshServiceClient,
		})

		Expect(checkApplies).To(BeFalse())
		Expect(runFailure).To(BeNil())
	})

	It("does not consider itself valid if there are mesh services but they are not federated", func() {
		meshServiceClient := mock_smh_discovery.NewMockMeshServiceClient(ctrl)
		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&smh_discovery.MeshServiceList{
				Items: []smh_discovery.MeshService{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-1"},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-2"},
					},
				},
			}, nil)

		runFailure, checkApplies := internal.NewFederationDecisionCheck().Run(ctx, container_runtime.GetWriteNamespace(), healthcheck_types.Clients{
			MeshServiceClient: meshServiceClient,
		})

		Expect(checkApplies).To(BeFalse())
		Expect(runFailure).To(BeNil())
	})

	It("reports no issues with successfully federated mesh services", func() {
		meshServiceClient := mock_smh_discovery.NewMockMeshServiceClient(ctrl)
		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&smh_discovery.MeshServiceList{
				Items: []smh_discovery.MeshService{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-1"},
						Status: smh_discovery_types.MeshServiceStatus{
							FederationStatus: &smh_core_types.Status{
								State: smh_core_types.Status_ACCEPTED,
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-2"},
						Status: smh_discovery_types.MeshServiceStatus{
							FederationStatus: &smh_core_types.Status{
								State: smh_core_types.Status_ACCEPTED,
							},
						},
					},
				},
			}, nil)

		runFailure, checkApplies := internal.NewFederationDecisionCheck().Run(ctx, container_runtime.GetWriteNamespace(), healthcheck_types.Clients{
			MeshServiceClient: meshServiceClient,
		})

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).To(BeNil())
	})

	It("reports an issue when federation has failed to be written to a mesh service", func() {
		meshServiceClient := mock_smh_discovery.NewMockMeshServiceClient(ctrl)
		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&smh_discovery.MeshServiceList{
				Items: []smh_discovery.MeshService{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-1", Namespace: container_runtime.GetWriteNamespace()},
						Status: smh_discovery_types.MeshServiceStatus{
							FederationStatus: &smh_core_types.Status{
								State: smh_core_types.Status_ACCEPTED,
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-2", Namespace: container_runtime.GetWriteNamespace()},
						Status: smh_discovery_types.MeshServiceStatus{
							FederationStatus: &smh_core_types.Status{
								State: smh_core_types.Status_INVALID,
							},
						},
					},
				},
			}, nil)

		federationChecker := internal.NewFederationDecisionCheck()
		clients := healthcheck_types.Clients{
			MeshServiceClient: meshServiceClient,
		}
		runFailure, checkApplies := federationChecker.Run(ctx, container_runtime.GetWriteNamespace(), clients)

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(internal.FederationRecordingHasFailed("test-2", container_runtime.GetWriteNamespace(), smh_core_types.Status_INVALID).Error()))
		Expect(runFailure.Hint).To(Equal(fmt.Sprintf("get details from the failing MeshService: `kubectl -n %s get meshservice %s -oyaml`", container_runtime.GetWriteNamespace(), "test-2")))
	})
})
