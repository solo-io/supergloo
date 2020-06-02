package traffic_policy_aggregation_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation"
	mock_traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation/mocks"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators"
	mock_mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators/mocks"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PolicyCollector", func() {
	var (
		ctrl              *gomock.Controller
		validationSuccess = &zephyr_core_types.Status{State: zephyr_core_types.Status_ACCEPTED}
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
				meshService := &zephyr_discovery.MeshService{}

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
				meshService := &zephyr_discovery.MeshService{
					Status: zephyr_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
							{
								Ref: &zephyr_core_types.ResourceRef{Name: "tp1"},
							},
							{
								Ref: &zephyr_core_types.ResourceRef{Name: "tp2"},
							},
						},
					},
				}

				aggregator.EXPECT().
					PoliciesForService(nil, meshService).
					Return(nil, nil)

				result, err := collector.CollectForService(meshService, []*zephyr_discovery.MeshService{meshService}, nil, nil, nil)
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
			meshService := &zephyr_discovery.MeshService{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"}}
			mesh := &zephyr_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
			trafficPolicies := []*zephyr_networking.TrafficPolicy{
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"}},
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"}},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
					Status: zephyr_networking_types.TrafficPolicyStatus{
						ValidationStatus: &zephyr_core_types.Status{State: zephyr_core_types.Status_INVALID},
					},
				},
			}

			aggregator.EXPECT().
				PoliciesForService(nil, meshService).
				Return(nil, nil)

			result, err := collector.CollectForService(meshService, []*zephyr_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
			Expect(err).To(BeNil())
			Expect(result.PoliciesToRecordOnService).To(BeNil())
		})

		It("adds the valid policies to the service", func() {
			aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
			collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
			validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
			meshService := &zephyr_discovery.MeshService{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"}}
			mesh := &zephyr_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
			trafficPolicies := []*zephyr_networking.TrafficPolicy{

				// the first one will not be associated with the service
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
					Spec: zephyr_networking_types.TrafficPolicySpec{
						Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts: 1,
						},
					},
					Status: zephyr_networking_types.TrafficPolicyStatus{
						ValidationStatus: validationSuccess,
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
					Spec: zephyr_networking_types.TrafficPolicySpec{
						Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts: 2,
						},
					},
					Status: zephyr_networking_types.TrafficPolicyStatus{
						ValidationStatus: validationSuccess,
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
					Spec: zephyr_networking_types.TrafficPolicySpec{
						Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts: 3,
						},
					},
					Status: zephyr_networking_types.TrafficPolicyStatus{
						ValidationStatus: validationSuccess,
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
					Spec: zephyr_networking_types.TrafficPolicySpec{
						Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts: 4,
						},
					},
					Status: zephyr_networking_types.TrafficPolicyStatus{
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
				GetTranslationErrors(meshService, []*zephyr_discovery.MeshService{meshService}, mesh, []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
					{
						Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
						TrafficPolicySpec: &trafficPolicies[1].Spec,
					},
				}).
				Return(nil)
			aggregator.EXPECT().
				FindMergeConflict(&trafficPolicies[2].Spec, []*zephyr_networking_types.TrafficPolicySpec{&trafficPolicies[1].Spec}, meshService).
				Return(nil)
			validator.EXPECT().
				GetTranslationErrors(meshService, []*zephyr_discovery.MeshService{meshService}, mesh, []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
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

			result, err := collector.CollectForService(meshService, []*zephyr_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
			Expect(err).To(BeNil())
			Expect(result.PoliciesToRecordOnService).To(Equal([]*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
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
				mesh := &zephyr_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
				trafficPolicies := []*zephyr_networking.TrafficPolicy{

					// the first one will not be associated with the service
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
						Spec: zephyr_networking_types.TrafficPolicySpec{
							Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
								Attempts: 1,
							},
						},
						Status: zephyr_networking_types.TrafficPolicyStatus{
							ValidationStatus: validationSuccess,
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
						Spec: zephyr_networking_types.TrafficPolicySpec{
							Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
								Attempts: 2,
							},
						},
						Status: zephyr_networking_types.TrafficPolicyStatus{
							ValidationStatus: validationSuccess,
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
						Spec: zephyr_networking_types.TrafficPolicySpec{
							Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
								Attempts: 3,
							},
						},
						Status: zephyr_networking_types.TrafficPolicyStatus{
							ValidationStatus: validationSuccess,
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
						Spec: zephyr_networking_types.TrafficPolicySpec{
							Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
								Attempts: 4,
							},
						},
						Status: zephyr_networking_types.TrafficPolicyStatus{
							// un-validated
						},
					},
				}
				meshService := &zephyr_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"},
					Status: zephyr_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
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

				result, err := collector.CollectForService(meshService, []*zephyr_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
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
					mesh := &zephyr_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
					trafficPolicies := []*zephyr_networking.TrafficPolicy{

						// the first one will not be associated with the service
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 1,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 2,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 3,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 4,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								// un-validated
							},
						},
					}
					meshService := &zephyr_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"},
						Status: zephyr_discovery_types.MeshServiceStatus{
							ValidatedTrafficPolicies: []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
								{
									Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
									TrafficPolicySpec: &trafficPolicies[1].Spec,
								},
								{
									Ref: selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
									TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
										Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
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
						FindMergeConflict(&trafficPolicies[2].Spec, []*zephyr_networking_types.TrafficPolicySpec{&trafficPolicies[1].Spec}, meshService).
						Return(nil)
					validator.EXPECT().
						GetTranslationErrors(meshService, []*zephyr_discovery.MeshService{meshService}, mesh, []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
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

					result, err := collector.CollectForService(meshService, []*zephyr_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
					Expect(err).To(BeNil())
					Expect(result.PoliciesToRecordOnService).To(Equal([]*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
						meshService.Status.ValidatedTrafficPolicies[0],
						{
							Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
							TrafficPolicySpec: &trafficPolicies[2].Spec,
						},
					}))
				})
			})

			When("there are validation errors to report", func() {
				It("reports merge conflicts and keeps the last-known good state", func() {
					aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
					collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
					validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
					mesh := &zephyr_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
					trafficPolicies := []*zephyr_networking.TrafficPolicy{

						// the first one will not be associated with the service
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 1,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 2,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 3,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 4,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								// un-validated
							},
						},
					}
					meshService := &zephyr_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"},
						Status: zephyr_discovery_types.MeshServiceStatus{
							ValidatedTrafficPolicies: []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
								{
									Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
									TrafficPolicySpec: &trafficPolicies[1].Spec,
								},
								{
									Ref: selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
									TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
										Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
											Attempts: 9999, // this is getting updated to the value "2"
										},
									},
								},
							},
						},
					}
					mergeConflict := &zephyr_networking_types.TrafficPolicyStatus_ConflictError{ErrorMessage: "whoops conflict"}

					aggregator.EXPECT().
						PoliciesForService(trafficPolicies[0:3], meshService).
						Return(trafficPolicies[1:3], nil)
					aggregator.EXPECT().
						FindMergeConflict(&trafficPolicies[2].Spec, []*zephyr_networking_types.TrafficPolicySpec{&trafficPolicies[1].Spec}, meshService).
						Return(mergeConflict)

					result, err := collector.CollectForService(meshService, []*zephyr_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
					Expect(err).To(BeNil())
					Expect(result.PoliciesToRecordOnService).To(Equal(meshService.Status.ValidatedTrafficPolicies))
					Expect(result.PolicyToConflictErrors).To(Equal(map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError{
						trafficPolicies[2]: {mergeConflict},
					}))
				})

				It("reports translation errors and keeps the last-known good state", func() {
					aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
					collector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
					validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)
					mesh := &zephyr_discovery.Mesh{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh"}}
					trafficPolicies := []*zephyr_networking.TrafficPolicy{

						// the first one will not be associated with the service
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 1,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 2,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 3,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								ValidationStatus: validationSuccess,
							},
						},
						{
							ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
							Spec: zephyr_networking_types.TrafficPolicySpec{
								Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
									Attempts: 4,
								},
							},
							Status: zephyr_networking_types.TrafficPolicyStatus{
								// un-validated
							},
						},
					}
					meshService := &zephyr_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "test-mesh-service"},
						Status: zephyr_discovery_types.MeshServiceStatus{
							ValidatedTrafficPolicies: []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
								{
									Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[1].ObjectMeta),
									TrafficPolicySpec: &trafficPolicies[1].Spec,
								},
								{
									Ref: selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
									TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
										Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
											Attempts: 9999, // this is getting updated to the value "2"
										},
									},
								},
							},
						},
					}
					translationError := &zephyr_networking_types.TrafficPolicyStatus_TranslatorError{ErrorMessage: "whoops translator error"}

					aggregator.EXPECT().
						PoliciesForService(trafficPolicies[0:3], meshService).
						Return(trafficPolicies[1:3], nil)
					aggregator.EXPECT().
						FindMergeConflict(&trafficPolicies[2].Spec, []*zephyr_networking_types.TrafficPolicySpec{&trafficPolicies[1].Spec}, meshService).
						Return(nil)
					validator.EXPECT().
						GetTranslationErrors(meshService, []*zephyr_discovery.MeshService{meshService}, mesh, []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
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
							Policy: &zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
								TrafficPolicySpec: &trafficPolicies[2].Spec,
								Ref:               selection.ObjectMetaToResourceRef(trafficPolicies[2].ObjectMeta),
							},
							TranslatorErrors: []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError{translationError},
						}})

					result, err := collector.CollectForService(meshService, []*zephyr_discovery.MeshService{meshService}, mesh, validator, trafficPolicies)
					Expect(err).To(BeNil())
					Expect(result.PoliciesToRecordOnService).To(Equal(meshService.Status.ValidatedTrafficPolicies))
					Expect(result.PolicyToTranslatorErrors).To(Equal(map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError{
						trafficPolicies[2]: {translationError},
					}))
				})
			})
		})
	})
})
