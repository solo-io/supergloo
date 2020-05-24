package aggregation_framework_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation"
	aggregation_framework "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation/framework"
	mock_traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation/mocks"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/meshes"
	mock_mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/meshes/mocks"
	mock_zephyr_discovery_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Traffic Policy Aggregation Reconciler", func() {
	var (
		ctx  context.Context
		ctrl *gomock.Controller
	)

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("when no resources exist in the cluster", func() {
		It("does nothing", func() {
			trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
			meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
			meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
			policyCollector := mock_traffic_policy_aggregation.NewMockPolicyCollector(ctrl)
			validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
			inMemoryStatusMutator := mock_traffic_policy_aggregation.NewMockInMemoryStatusMutator(ctrl)
			reconciler := aggregation_framework.NewAggregationReconciler(
				trafficPolicyClient,
				meshServiceClient,
				meshClient,
				policyCollector,
				map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator{
					zephyr_core_types.MeshType_ISTIO1_5: validator,
				},
				inMemoryStatusMutator,
			)

			trafficPolicyClient.EXPECT().
				ListTrafficPolicy(ctx).
				Return(&zephyr_networking.TrafficPolicyList{}, nil)
			meshServiceClient.EXPECT().
				ListMeshService(ctx).
				Return(&zephyr_discovery.MeshServiceList{}, nil)

			err := reconciler.Reconcile(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when no policies have been written yet, but there are services", func() {
		Context("and the services have no previously-written policies on their status", func() {
			It("does nothing", func() {
				trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
				meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
				meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
				policyCollector := mock_traffic_policy_aggregation.NewMockPolicyCollector(ctrl)
				validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
				inMemoryStatusMutator := mock_traffic_policy_aggregation.NewMockInMemoryStatusMutator(ctrl)
				reconciler := aggregation_framework.NewAggregationReconciler(
					trafficPolicyClient,
					meshServiceClient,
					meshClient,
					policyCollector,
					map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator{
						zephyr_core_types.MeshType_ISTIO1_5: validator,
					},
					inMemoryStatusMutator,
				)
				mesh1 := &zephyr_discovery.Mesh{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "mesh-1",
					},
					Spec: zephyr_discovery_types.MeshSpec{
						Cluster: &zephyr_core_types.ResourceRef{
							Name: "cluster1",
						},
						MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
					},
				}
				mesh2 := &zephyr_discovery.Mesh{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "mesh-2",
					},
					Spec: zephyr_discovery_types.MeshSpec{
						Cluster: &zephyr_core_types.ResourceRef{
							Name: "cluster2",
						},
						MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
					},
				}
				meshServices := []*zephyr_discovery.MeshService{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name: "ms1",
						},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(mesh1.ObjectMeta),
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name: "ms2",
						},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(mesh2.ObjectMeta),
						},
					},
				}

				meshServiceClient.EXPECT().
					ListMeshService(ctx).
					Return(&zephyr_discovery.MeshServiceList{Items: []zephyr_discovery.MeshService{*meshServices[0], *meshServices[1]}}, nil)
				meshClient.EXPECT().
					GetMesh(ctx, selection.ResourceRefToObjectKey(meshServices[0].Spec.Mesh)).
					Return(mesh1, nil)
				meshClient.EXPECT().
					GetMesh(ctx, selection.ResourceRefToObjectKey(meshServices[1].Spec.Mesh)).
					Return(mesh2, nil)
				policyCollector.EXPECT().
					CollectForService(meshServices[0], mesh1, validator, nil).
					Return(&traffic_policy_aggregation.CollectionResult{}, nil)
				policyCollector.EXPECT().
					CollectForService(meshServices[1], mesh2, validator, nil).
					Return(&traffic_policy_aggregation.CollectionResult{}, nil)
				inMemoryStatusMutator.EXPECT().
					MutateServicePolicies(meshServices[0], nil).
					Return(false)
				inMemoryStatusMutator.EXPECT().
					MutateServicePolicies(meshServices[1], nil).
					Return(false)

				trafficPolicyClient.EXPECT().
					ListTrafficPolicy(ctx).
					Return(&zephyr_networking.TrafficPolicyList{}, nil)

				err := reconciler.Reconcile(ctx)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("when there are both policies and services to process", func() {
		It("computes new statuses accordingly", func() {
			trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
			meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
			meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
			policyCollector := mock_traffic_policy_aggregation.NewMockPolicyCollector(ctrl)
			validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
			inMemoryStatusMutator := mock_traffic_policy_aggregation.NewMockInMemoryStatusMutator(ctrl)
			reconciler := aggregation_framework.NewAggregationReconciler(
				trafficPolicyClient,
				meshServiceClient,
				meshClient,
				policyCollector,
				map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator{
					zephyr_core_types.MeshType_ISTIO1_5: validator,
				},
				inMemoryStatusMutator,
			)
			mesh1 := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "mesh-1",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: "cluster1",
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
				},
			}
			mesh2 := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "mesh-2",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: "cluster2",
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
				},
			}
			meshServices := []*zephyr_discovery.MeshService{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "ms1",
					},
					Spec: zephyr_discovery_types.MeshServiceSpec{
						Mesh: selection.ObjectMetaToResourceRef(mesh1.ObjectMeta),
					},
					Status: zephyr_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
							Ref: &zephyr_core_types.ResourceRef{Name: "validated-1"},
						}},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "ms2",
					},
					Spec: zephyr_discovery_types.MeshServiceSpec{
						Mesh: selection.ObjectMetaToResourceRef(mesh2.ObjectMeta),
					},
					Status: zephyr_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
							Ref: &zephyr_core_types.ResourceRef{Name: "validated-2"},
						}},
					},
				},
			}
			trafficPolicies := []*zephyr_networking.TrafficPolicy{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "trafficPolicies[0]"},
					Spec: types.TrafficPolicySpec{
						Retries: &types.TrafficPolicySpec_RetryPolicy{Attempts: 0},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "trafficPolicies[1]"},
					Spec: types.TrafficPolicySpec{
						Retries: &types.TrafficPolicySpec_RetryPolicy{Attempts: 1},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "trafficPolicies[2]"},
					Spec: types.TrafficPolicySpec{
						Retries: &types.TrafficPolicySpec_RetryPolicy{Attempts: 2},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "trafficPolicies[3]"},
					Spec: types.TrafficPolicySpec{
						Retries: &types.TrafficPolicySpec_RetryPolicy{Attempts: 3},
					},
				},
			}
			newlyValidatedTrafficPolicies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[0].ObjectMeta),
					TrafficPolicySpec: &trafficPolicies[0].Spec,
				},
				{
					Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
					TrafficPolicySpec: &trafficPolicies[1].Spec,
				},
				{
					Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
					TrafficPolicySpec: &trafficPolicies[2].Spec,
				},
				{
					Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[3].ObjectMeta),
					TrafficPolicySpec: &trafficPolicies[3].Spec,
				},
			}
			conflictErrors := []*types.TrafficPolicyStatus_ConflictError{{ErrorMessage: "whoops"}}

			trafficPolicyClient.EXPECT().
				ListTrafficPolicy(ctx).
				Return(&zephyr_networking.TrafficPolicyList{Items: []zephyr_networking.TrafficPolicy{
					*trafficPolicies[0],
					*trafficPolicies[1],
					*trafficPolicies[2],
					*trafficPolicies[3],
				}}, nil)
			meshServiceClient.EXPECT().
				ListMeshService(ctx).
				Return(&zephyr_discovery.MeshServiceList{Items: []zephyr_discovery.MeshService{*meshServices[0], *meshServices[1]}}, nil)
			meshClient.EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshServices[0].Spec.Mesh)).
				Return(mesh1, nil)
			meshClient.EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshServices[1].Spec.Mesh)).
				Return(mesh2, nil)
			policyCollector.EXPECT().
				CollectForService(meshServices[0], mesh1, validator, trafficPolicies).
				Return(&traffic_policy_aggregation.CollectionResult{PoliciesToRecordOnService: newlyValidatedTrafficPolicies[0:2]}, nil)
			policyCollector.EXPECT().
				CollectForService(meshServices[1], mesh2, validator, trafficPolicies).
				Return(&traffic_policy_aggregation.CollectionResult{
					PoliciesToRecordOnService: newlyValidatedTrafficPolicies[2:4],
					PolicyToConflictErrors: map[*zephyr_networking.TrafficPolicy][]*types.TrafficPolicyStatus_ConflictError{
						trafficPolicies[3]: conflictErrors,
					},
				}, nil)

			ms1Copy := *meshServices[0]
			ms1Copy.Status = zephyr_discovery_types.MeshServiceStatus{
				ValidatedTrafficPolicies: newlyValidatedTrafficPolicies[0:2],
			}
			ms2Copy := *meshServices[1]
			ms2Copy.Status = zephyr_discovery_types.MeshServiceStatus{
				ValidatedTrafficPolicies: newlyValidatedTrafficPolicies[2:4],
			}
			policy4Copy := *trafficPolicies[3]
			policy4Copy.Status = types.TrafficPolicyStatus{
				ConflictErrors: conflictErrors,
			}

			inMemoryStatusMutator.EXPECT().
				MutateServicePolicies(meshServices[0], newlyValidatedTrafficPolicies[0:2]).
				DoAndReturn(func(
					meshService *zephyr_discovery.MeshService,
					newlyComputedMergeablePolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
				) (policyNeedsUpdating bool) {
					meshService.Status = zephyr_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: newlyComputedMergeablePolicies,
					}
					return true
				})
			inMemoryStatusMutator.EXPECT().
				MutateServicePolicies(meshServices[1], newlyValidatedTrafficPolicies[2:4]).
				DoAndReturn(func(
					meshService *zephyr_discovery.MeshService,
					newlyComputedMergeablePolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
				) (policyNeedsUpdating bool) {
					meshService.Status = zephyr_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: newlyComputedMergeablePolicies,
					}
					return true
				})
			inMemoryStatusMutator.EXPECT().
				MutateConflictAndTranslatorErrors(gomock.Any(), gomock.Any(), nil).
				DoAndReturn(func(
					policy *zephyr_networking.TrafficPolicy,
					newConflictErrors []*types.TrafficPolicyStatus_ConflictError,
					newTranslationErrors []*types.TrafficPolicyStatus_TranslatorError,
				) (policyNeedsUpdating bool) {
					if policy.GetName() != trafficPolicies[3].GetName() {
						return false
					}

					policy.Status = policy4Copy.Status
					return true
				}).
				Times(4)
			meshServiceClient.EXPECT().
				UpdateMeshServiceStatus(ctx, &ms1Copy).
				Return(nil)
			meshServiceClient.EXPECT().
				UpdateMeshServiceStatus(ctx, &ms2Copy).
				Return(nil)
			trafficPolicyClient.EXPECT().
				UpdateTrafficPolicyStatus(ctx, &policy4Copy).
				Return(nil)

			err := reconciler.Reconcile(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
