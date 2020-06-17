package traffic_policy_validation_test

import (
	types1 "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/common/kube/selection/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/routing/traffic-policy-translator/preprocess"
	traffic_policy_validation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Validator", func() {
	var (
		ctrl                 *gomock.Controller
		mockResourceSelector *mock_selector.MockResourceSelector
		validator            traffic_policy_validation.Validator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockResourceSelector = mock_selector.NewMockResourceSelector(ctrl)
		validator = traffic_policy_validation.NewValidator(mockResourceSelector)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should return error if Destination cannot be found", func() {
		name := "name"
		namespace := "namespace"
		cluster := "cluster"
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				DestinationSelector: &smh_core_types.ServiceSelector{
					ServiceSelectorType: &smh_core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &smh_core_types.ServiceSelector_ServiceRefs{
							Services: []*smh_core_types.ResourceRef{
								{
									Name:      name,
									Namespace: namespace,
									Cluster:   cluster,
								},
							},
						},
					},
				},
			},
		}
		meshServices := []*smh_discovery.MeshService{}
		mockResourceSelector.
			EXPECT().
			FilterMeshServicesByServiceSelector(meshServices, tp.Spec.GetDestinationSelector()).
			Return(nil, selection.MeshServiceNotFound(name, namespace, cluster))
		status, err := validator.ValidateTrafficPolicy(tp, meshServices)
		expectSingleErrorOf(err, selection.MeshServiceNotFound(name, namespace, cluster))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("should return error for RetryPolicy with negative num attempts", func() {
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{Attempts: -1},
			},
		}
		meshServices := []*smh_discovery.MeshService{}
		status, err := validator.ValidateTrafficPolicy(tp, meshServices)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.InvalidRetryPolicyNumAttempts(-1))))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("should return error for RetryPolicy with negative per retry timeout", func() {
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{PerTryTimeout: &types1.Duration{Seconds: -1}},
			},
		}
		meshServices := []*smh_discovery.MeshService{}
		status, err := validator.ValidateTrafficPolicy(tp, meshServices)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.MinDurationError)))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("should return error if TrafficShift has destinations that cannot be found", func() {
		name := "name"
		namespace := "namespace"
		cluster := ""
		serviceRef := &smh_core_types.ResourceRef{
			Name:      name,
			Namespace: namespace,
			Cluster:   cluster,
		}
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				TrafficShift: &smh_networking_types.TrafficPolicySpec_MultiDestination{
					Destinations: []*smh_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							Destination: serviceRef,
							Weight:      1,
						},
					},
				},
			},
		}
		meshServices := []*smh_discovery.MeshService{}
		mockResourceSelector.
			EXPECT().
			FindMeshServiceByRefSelector(meshServices, name, namespace, cluster).
			Return(nil)
		status, err := validator.ValidateTrafficPolicy(tp, meshServices)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(traffic_policy_validation.DestinationNotFound(&smh_core_types.ResourceRef{
			Name:      name,
			Namespace: namespace,
			Cluster:   cluster,
		}))))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("should return error if TrafficShift has subsets that can't be found", func() {
		name := "name"
		namespace := "namespace"
		cluster := ""
		serviceRef := &smh_core_types.ResourceRef{
			Name:      name,
			Namespace: namespace,
			Cluster:   cluster,
		}
		subset := map[string]string{"env": "dev", "version": "v1"}
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				TrafficShift: &smh_networking_types.TrafficPolicySpec_MultiDestination{
					Destinations: []*smh_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							Destination: serviceRef,
							Subset:      subset,
							Weight:      1,
						},
					},
				},
			},
		}
		backingMeshService := &smh_discovery.MeshService{Spec: smh_discovery_types.MeshServiceSpec{
			Subsets: map[string]*smh_discovery_types.MeshServiceSpec_Subset{
				"env":     {Values: []string{"dev", "prod"}},
				"version": {Values: []string{"v2", "v3"}},
			},
		}}
		meshServices := []*smh_discovery.MeshService{backingMeshService}
		mockResourceSelector.
			EXPECT().
			FindMeshServiceByRefSelector(meshServices, serviceRef.GetName(), serviceRef.GetNamespace(), serviceRef.GetCluster()).
			Return(backingMeshService)
		status, err := validator.ValidateTrafficPolicy(tp, meshServices)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.SubsetSelectorNotFound(backingMeshService, "version", "v1"))))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("should return error if FaultInjection Abort has invalid percentage", func() {
		invalidPct := 101.0
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				FaultInjection: &smh_networking_types.TrafficPolicySpec_FaultInjection{
					Percentage: invalidPct,
				},
			},
		}
		meshServices := []*smh_discovery.MeshService{}
		status, err := validator.ValidateTrafficPolicy(tp, meshServices)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.InvalidPercentageError(invalidPct))))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("should return error if FaultInjection Abort has invalid HTTP status", func() {
		invalidHttpStatus := int32(432)
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				FaultInjection: &smh_networking_types.TrafficPolicySpec_FaultInjection{
					FaultInjectionType: &smh_networking_types.TrafficPolicySpec_FaultInjection_Abort_{
						Abort: &smh_networking_types.TrafficPolicySpec_FaultInjection_Abort{
							ErrorType: &smh_networking_types.TrafficPolicySpec_FaultInjection_Abort_HttpStatus{
								HttpStatus: invalidHttpStatus,
							},
						},
					},
					Percentage: 50,
				},
			},
		}
		status, err := validator.ValidateTrafficPolicy(tp, nil)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.InvalidHttpStatus(invalidHttpStatus))))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("should return error if FaultInjection Delay has invalid duration", func() {
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				FaultInjection: &smh_networking_types.TrafficPolicySpec_FaultInjection{
					FaultInjectionType: &smh_networking_types.TrafficPolicySpec_FaultInjection_Delay_{
						Delay: &smh_networking_types.TrafficPolicySpec_FaultInjection_Delay{
							HttpDelayType: &smh_networking_types.TrafficPolicySpec_FaultInjection_Delay_FixedDelay{
								FixedDelay: &types1.Duration{Seconds: -1},
							},
						},
					},
					Percentage: 50,
				},
			},
		}
		status, err := validator.ValidateTrafficPolicy(tp, nil)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.MinDurationError)))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("should return error if RequestTimeout has an invalid duration", func() {
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				RequestTimeout: &types1.Duration{Seconds: 0, Nanos: 999999},
			},
		}
		status, err := validator.ValidateTrafficPolicy(tp, nil)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.MinDurationError)))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("should return error if CorsPolicy has an invalid max age duration", func() {
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				CorsPolicy: &smh_networking_types.TrafficPolicySpec_CorsPolicy{
					MaxAge: &types1.Duration{Seconds: 0, Nanos: 999999},
				},
			},
		}
		status, err := validator.ValidateTrafficPolicy(tp, nil)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.MinDurationError)))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("should return error if Mirror has invalid percentage", func() {
		serviceKey := client.ObjectKey{Name: "name", Namespace: "namespace"}
		serviceRef := &smh_core_types.ResourceRef{
			Name:      serviceKey.Name,
			Namespace: serviceKey.Namespace,
		}
		invalidPct := 101.0
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				Mirror: &smh_networking_types.TrafficPolicySpec_Mirror{
					Destination: serviceRef,
					Percentage:  invalidPct,
				},
			},
		}
		backingMeshService := &smh_discovery.MeshService{Spec: smh_discovery_types.MeshServiceSpec{
			Subsets: map[string]*smh_discovery_types.MeshServiceSpec_Subset{
				"env":     {Values: []string{"dev", "prod"}},
				"version": {Values: []string{"v2", "v3"}},
			},
		}}
		meshServices := []*smh_discovery.MeshService{backingMeshService}
		mockResourceSelector.
			EXPECT().
			FindMeshServiceByRefSelector(meshServices, serviceKey.Name, serviceKey.Namespace, "").
			Return(backingMeshService)
		status, err := validator.ValidateTrafficPolicy(tp, meshServices)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.InvalidPercentageError(invalidPct))))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("should return error if Mirror has destination that cannot be found", func() {
		serviceKey := client.ObjectKey{Name: "name", Namespace: "namespace"}
		serviceRef := &smh_core_types.ResourceRef{
			Name:      serviceKey.Name,
			Namespace: serviceKey.Namespace,
		}
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				Mirror: &smh_networking_types.TrafficPolicySpec_Mirror{
					Destination: serviceRef,
					Percentage:  50,
				},
			},
		}
		mockResourceSelector.
			EXPECT().
			FindMeshServiceByRefSelector(nil, serviceKey.Name, serviceKey.Namespace, "").
			Return(nil)
		status, err := validator.ValidateTrafficPolicy(tp, nil)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(traffic_policy_validation.DestinationNotFound(serviceRef))))
		Expect(status.State).To(Equal(smh_core_types.Status_INVALID))
	})
})

func expectSingleErrorOf(err error, expected error) {
	multierr, ok := err.(*multierror.Error)
	Expect(ok).To(BeTrue())
	Expect(multierr.Errors).To(HaveLen(1))
	Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(expected))
}
