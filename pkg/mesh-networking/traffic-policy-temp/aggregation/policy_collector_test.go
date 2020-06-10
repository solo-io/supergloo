package traffic_policy_aggregation_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/aggregation"
	mock_traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/aggregation/mocks"
	mesh_translation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/translators"
	mock_mesh_translation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/translators/mocks"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PolicyCollector", func() {
	var (
		ctrl              *gomock.Controller
		validationSuccess = &smh_core_types.Status{State: smh_core_types.Status_ACCEPTED}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	When("there are no policies in the cluster", func() {
		Context("and the service has no previously-recorded policies", func() {
			It("does not do anything", func() {
				aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
				collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
				meshService := &smh_discovery.MeshService{}

				aggregator.EXPECT().
					PoliciesForService(nil, meshService).
					Return(nil, nil)

				result, err := collector.CollectForService(meshService, nil, nil, nil, nil)
				Expect(err).To(BeNil())
				Expect(result.PoliciesToRecordOnService).To(BeNil())
			})
		})

		Context("and there are previously-recorded policies on the service", func() {
			It("should indicate that all policies be removed from the service", func() {
				aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
				collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
				meshService := &smh_discovery.MeshService{
					Status: smh_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
							{
								Ref: &smh_core_types.ResourceRef{Name: "tp1"},
							},
							{
								Ref: &smh_core_types.ResourceRef{Name: "tp2"},
							},
						},
					},
				}

				aggregator.EXPECT().
					PoliciesForService(nil, meshService).
					Return(nil, nil)

				result, err := collector.CollectForService(meshService, []*smh_discovery.MeshService{meshService}, nil, nil, nil)
				Expect(err).To(BeNil())
				Expect(result.PoliciesToRecordOnService).To(BeNil())
			})
		})
	})

	When("policies have been written for the first time", func() {
		It("ignores invalid or un-validated traffic policies", func() {
			aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
			collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
			validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
			meshService := &smh_discovery.MeshService{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"}}
			mesh := &smh_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
			trafficPolicies := []*smh_networking.TrafficPolicy{
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"}},
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"}},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
					Status: smh_networking_types.TrafficPolicyStatus{
						ValidationStatus: &smh_core_types.Status{State: smh_core_types.Status_INVALID},
					},
				},
			}

			aggregator.EXPECT().
				PoliciesForService(nil, meshService).
				Return(nil, nil)

			result, err := collector.CollectForService(meshService, []*smh_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
			Expect(err).To(BeNil())
			Expect(result.PoliciesToRecordOnService).To(BeNil())
		})

		It("adds the valid policies to the service", func() {
			aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
			collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
			validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
			meshService := &smh_discovery.MeshService{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"}}
			mesh := &smh_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
			trafficPolicies := []*smh_networking.TrafficPolicy{

				// the first one will not be associated with the service
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
					Spec: smh_networking_types.TrafficPolicySpec{
						Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts: 1,
						},
					},
					Status: smh_networking_types.TrafficPolicyStatus{
						ValidationStatus: validationSuccess,
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
					Spec: smh_networking_types.TrafficPolicySpec{
						Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts: 2,
						},
					},
					Status: smh_networking_types.TrafficPolicyStatus{
						ValidationStatus: validationSuccess,
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
					Spec: smh_networking_types.TrafficPolicySpec{
						Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts: 3,
						},
					},
					Status: smh_networking_types.TrafficPolicyStatus{
						ValidationStatus: validationSuccess,
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
					Spec: smh_networking_types.TrafficPolicySpec{
						Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts: 4,
						},
					},
					Status: smh_networking_types.TrafficPolicyStatus{
						// un-validated
					},
				},
			}

			aggregator.EXPECT().
				PoliciesForService(trafficPolicies[0:3], meshService).
				Return(trafficPolicies[1:3], nil)
			aggregator.EXPECT().
				FindMergeConflict(&trafficPolicies[1].Spec, nil, meshService).
				Return(nil)
			validator.EXPECT().
				GetTranslationErrors(meshService, []*smh_discovery.MeshService{meshService}, mesh, []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
					{
						Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
						TrafficPolicySpec: &trafficPolicies[1].Spec,
					},
				}).
				Return(nil)
			aggregator.EXPECT().
				FindMergeConflict(&trafficPolicies[2].Spec, []*smh_networking_types.TrafficPolicySpec{&trafficPolicies[1].Spec}, meshService).
				Return(nil)
			validator.EXPECT().
				GetTranslationErrors(meshService, []*smh_discovery.MeshService{meshService}, mesh, []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
					{
						Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
						TrafficPolicySpec: &trafficPolicies[1].Spec,
					},
					{
						Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
						TrafficPolicySpec: &trafficPolicies[2].Spec,
					},
				}).
				Return(nil)

			result, err := collector.CollectForService(meshService, []*smh_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
			Expect(err).To(BeNil())
			Expect(result.PoliciesToRecordOnService).To(Equal([]*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
					TrafficPolicySpec: &trafficPolicies[1].Spec,
				},
				{
					Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
					TrafficPolicySpec: &trafficPolicies[2].Spec,
				},
			}))
		})
	})

	When("policies must be merged in and validated with existing recorded policies", func() {
		Context("and none of them have changed", func() {
			It("should return the mesh service's previous status verbatim, without running any validations", func() {
				aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
				collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
				validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
				mesh := &smh_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
				trafficPolicies := []*smh_networking.TrafficPolicy{

					// the first one will not be associated with the service
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
						Spec: smh_networking_types.TrafficPolicySpec{
							Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
								Attempts: 1,
							},
						},
						Status: smh_networking_types.TrafficPolicyStatus{
							ValidationStatus: validationSuccess,
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
						Spec: smh_networking_types.TrafficPolicySpec{
							Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
								Attempts: 2,
							},
						},
						Status: smh_networking_types.TrafficPolicyStatus{
							ValidationStatus: validationSuccess,
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
						Spec: smh_networking_types.TrafficPolicySpec{
							Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
								Attempts: 3,
							},
						},
						Status: smh_networking_types.TrafficPolicyStatus{
							ValidationStatus: validationSuccess,
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
						Spec: smh_networking_types.TrafficPolicySpec{
							Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
								Attempts: 4,
							},
						},
						Status: smh_networking_types.TrafficPolicyStatus{
							// un-validated
						},
					},
				}
				meshService := &smh_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"},
					Status: smh_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
							{
								Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
								TrafficPolicySpec: &trafficPolicies[1].Spec,
							},
							{
								Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
								TrafficPolicySpec: &trafficPolicies[2].Spec,
							},
						},
					},
				}

				aggregator.EXPECT().
					PoliciesForService(trafficPolicies[0:3], meshService).
					Return(trafficPolicies[1:3], nil)

				result, err := collector.CollectForService(meshService, []*smh_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
				Expect(err).To(BeNil())
				Expect(result.PoliciesToRecordOnService).To(Equal(meshService.Status.ValidatedTrafficPolicies))
			})
		})

		Context("and there is a mix of updated and non-updated", func() {
			When("all policies pass all validation", func() {
				It("updates the service status appropriately", func() {
					aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
					collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
					validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
					mesh := &smh_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
					trafficPolicies := []*smh_networking.TrafficPolicy{

						// the first one will not be associated with the service
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 1,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 2,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 3,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 4,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								// un-validated
							},
						},
					}
					meshService := &smh_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"},
						Status: smh_discovery_types.MeshServiceStatus{
							ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
								{
									Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
									TrafficPolicySpec: &trafficPolicies[1].Spec,
								},
								{
									Ref: selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
									TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
										Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
											Attempts: 9999, // this is getting updated to the value "2"
										},
									},
								},
							},
						},
					}

					aggregator.EXPECT().
						PoliciesForService(trafficPolicies[0:3], meshService).
						Return(trafficPolicies[1:3], nil)
					aggregator.EXPECT().
						FindMergeConflict(&trafficPolicies[2].Spec, []*smh_networking_types.TrafficPolicySpec{&trafficPolicies[1].Spec}, meshService).
						Return(nil)
					validator.EXPECT().
						GetTranslationErrors(meshService, []*smh_discovery.MeshService{meshService}, mesh, []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
							{
								Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
								TrafficPolicySpec: &trafficPolicies[1].Spec,
							},
							{
								Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
								TrafficPolicySpec: &trafficPolicies[2].Spec,
							},
						}).
						Return(nil)

					result, err := collector.CollectForService(meshService, []*smh_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
					Expect(err).To(BeNil())
					Expect(result.PoliciesToRecordOnService).To(Equal([]*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
						meshService.Status.ValidatedTrafficPolicies[0],
						{
							Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
							TrafficPolicySpec: &trafficPolicies[2].Spec,
						},
					}))
				})

				It("updates the service status appropriately only with observed policies", func() {
					aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
					collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
					validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
					mesh := &smh_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
					validatedSpec := &smh_networking_types.TrafficPolicySpec{
						Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts: 1,
						},
					}
					trafficPolicies := []*smh_networking.TrafficPolicy{
						// the first one will not be associated with the service
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1", Generation: 2},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 2,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								ObservedGeneration: 1,
								ValidationStatus:   validationSuccess,
							},
						},
					}
					meshService := &smh_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"},
						Status: smh_discovery_types.MeshServiceStatus{
							ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
								{
									Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[0].ObjectMeta),
									TrafficPolicySpec: validatedSpec,
								},
							},
						},
					}

					// validatedPolicies will be empty as no validated policies exist (the one we have is ignored as generation != observedGeneration)
					var validatedPolicies []*smh_networking.TrafficPolicy
					aggregator.EXPECT().
						PoliciesForService(validatedPolicies, meshService).
						Return(trafficPolicies, nil)

					result, err := collector.CollectForService(meshService, []*smh_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
					Expect(err).To(BeNil())
					Expect(result.PoliciesToRecordOnService).To(Equal([]*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[0].ObjectMeta),
							TrafficPolicySpec: validatedSpec,
						},
					}))
				})
			})

			When("there are validation errors to report", func() {
				It("reports merge conflicts and keeps the last-known good state", func() {
					aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
					collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
					validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
					mesh := &smh_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
					trafficPolicies := []*smh_networking.TrafficPolicy{

						// the first one will not be associated with the service
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 1,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 2,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 3,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 4,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								// un-validated
							},
						},
					}
					meshService := &smh_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"},
						Status: smh_discovery_types.MeshServiceStatus{
							ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
								{
									Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
									TrafficPolicySpec: &trafficPolicies[1].Spec,
								},
								{
									Ref: selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
									TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
										Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
											Attempts: 9999, // this is getting updated to the value "2"
										},
									},
								},
							},
						},
					}
					mergeConflict := &smh_networking_types.TrafficPolicyStatus_ConflictError{ErrorMessage: "whoops conflict"}

					aggregator.EXPECT().
						PoliciesForService(trafficPolicies[0:3], meshService).
						Return(trafficPolicies[1:3], nil)
					aggregator.EXPECT().
						FindMergeConflict(&trafficPolicies[2].Spec, []*smh_networking_types.TrafficPolicySpec{&trafficPolicies[1].Spec}, meshService).
						Return(mergeConflict)

					result, err := collector.CollectForService(meshService, []*smh_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
					Expect(err).To(BeNil())
					Expect(result.PoliciesToRecordOnService).To(Equal(meshService.Status.ValidatedTrafficPolicies))
					Expect(result.PolicyToConflictErrors).To(Equal(map[*smh_networking.TrafficPolicy][]*smh_networking_types.TrafficPolicyStatus_ConflictError{
						trafficPolicies[2]: {mergeConflict},
					}))
				})

				It("reports translation errors and keeps the last-known good state", func() {
					aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
					collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
					validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
					mesh := &smh_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
					trafficPolicies := []*smh_networking.TrafficPolicy{

						// the first one will not be associated with the service
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 1,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 2,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 3,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
							Spec: smh_networking_types.TrafficPolicySpec{
								Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 4,
								},
							},
							Status: smh_networking_types.TrafficPolicyStatus{
								// un-validated
							},
						},
					}
					meshService := &smh_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"},
						Status: smh_discovery_types.MeshServiceStatus{
							ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
								{
									Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
									TrafficPolicySpec: &trafficPolicies[1].Spec,
								},
								{
									Ref: selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
									TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
										Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
											Attempts: 9999, // this is getting updated to the value "2"
										},
									},
								},
							},
						},
					}
					translationError := &smh_networking_types.TrafficPolicyStatus_TranslatorError{ErrorMessage: "whoops translator error"}

					aggregator.EXPECT().
						PoliciesForService(trafficPolicies[0:3], meshService).
						Return(trafficPolicies[1:3], nil)
					aggregator.EXPECT().
						FindMergeConflict(&trafficPolicies[2].Spec, []*smh_networking_types.TrafficPolicySpec{&trafficPolicies[1].Spec}, meshService).
						Return(nil)
					validator.EXPECT().
						GetTranslationErrors(meshService, []*smh_discovery.MeshService{meshService}, mesh, []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
							{
								Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
								TrafficPolicySpec: &trafficPolicies[1].Spec,
							},
							{
								Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
								TrafficPolicySpec: &trafficPolicies[2].Spec,
							},
						}).
						Return([]*mesh_translation.TranslationError{{
							Policy: &smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
								TrafficPolicySpec: &trafficPolicies[2].Spec,
								Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
							},
							TranslatorErrors: []*smh_networking_types.TrafficPolicyStatus_TranslatorError{translationError},
						}})

					result, err := collector.CollectForService(meshService, []*smh_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
					Expect(err).To(BeNil())
					Expect(result.PoliciesToRecordOnService).To(Equal(meshService.Status.ValidatedTrafficPolicies))
					Expect(result.PolicyToTranslatorErrors).To(Equal(map[*smh_networking.TrafficPolicy][]*smh_networking_types.TrafficPolicyStatus_TranslatorError{
						trafficPolicies[2]: {translationError},
					}))
				})
			})
		})
	})
})
