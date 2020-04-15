package preprocess_test

import (
	"context"

	types1 "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/selector/mocks"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Validator", func() {
	var (
		ctrl                  *gomock.Controller
		ctx                   context.Context
		mockMeshServiceClient *mock_zephyr_discovery.MockMeshServiceClient
		mockResourceSelector  *mock_selector.MockResourceSelector
		validator             preprocess.TrafficPolicyValidator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshServiceClient = mock_zephyr_discovery.NewMockMeshServiceClient(ctrl)
		mockResourceSelector = mock_selector.NewMockResourceSelector(ctrl)
		validator = preprocess.NewTrafficPolicyValidator(mockMeshServiceClient, mockResourceSelector)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should return error if Destination cannot be found", func() {
		name := "name"
		namespace := "namespace"
		cluster := "cluster"
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				DestinationSelector: &zephyr_core_types.ServiceSelector{
					ServiceSelectorType: &zephyr_core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &zephyr_core_types.ServiceSelector_ServiceRefs{
							Services: []*zephyr_core_types.ResourceRef{
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
		mockResourceSelector.
			EXPECT().
			GetMeshServicesByServiceSelector(ctx, tp.Spec.GetDestinationSelector()).
			Return(nil, selector.MeshServiceNotFound(name, namespace, cluster))
		err := validator.Validate(ctx, tp)
		expectSingleErrorOf(err, selector.MeshServiceNotFound(name, namespace, cluster))
	})

	It("should return error for RetryPolicy with negative num attempts", func() {
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{Attempts: -1},
			},
		}
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.InvalidRetryPolicyNumAttempts(-1))))
	})

	It("should return error for RetryPolicy with negative per retry timeout", func() {
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{PerTryTimeout: &types1.Duration{Seconds: -1}},
			},
		}
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.MinDurationError)))
	})

	It("should return error if TrafficShift has destinations that cannot be found", func() {
		name := "name"
		namespace := "namespace"
		cluster := ""
		serviceRef := &zephyr_core_types.ResourceRef{
			Name:      name,
			Namespace: namespace,
			Cluster:   cluster,
		}
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				TrafficShift: &zephyr_networking_types.TrafficPolicySpec_MultiDestination{
					Destinations: []*zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							Destination: serviceRef,
							Weight:      1,
						},
					},
				},
			},
		}
		mockResourceSelector.
			EXPECT().
			GetMeshServiceByRefSelector(ctx, name, namespace, cluster).
			Return(nil, selector.MeshServiceNotFound(name, namespace, cluster))
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(selector.MeshServiceNotFound(name, namespace, cluster))))
	})

	It("should return error if TrafficShift has subsets that can't be found", func() {
		name := "name"
		namespace := "namespace"
		cluster := ""
		serviceRef := &zephyr_core_types.ResourceRef{
			Name:      name,
			Namespace: namespace,
			Cluster:   cluster,
		}
		subset := map[string]string{"env": "dev", "version": "v1"}
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				TrafficShift: &zephyr_networking_types.TrafficPolicySpec_MultiDestination{
					Destinations: []*zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							Destination: serviceRef,
							Subset:      subset,
							Weight:      1,
						},
					},
				},
			},
		}
		backingMeshService := &zephyr_discovery.MeshService{Spec: zephyr_discovery_types.MeshServiceSpec{
			Subsets: map[string]*zephyr_discovery_types.MeshServiceSpec_Subset{
				"env":     {Values: []string{"dev", "prod"}},
				"version": {Values: []string{"v2", "v3"}},
			},
		}}
		mockResourceSelector.
			EXPECT().
			GetMeshServiceByRefSelector(ctx, serviceRef.GetName(), serviceRef.GetNamespace(), serviceRef.GetCluster()).
			Return(backingMeshService, nil)
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.SubsetSelectorNotFound(backingMeshService, "version", "v1"))))
	})

	It("should return error if FaultInjection Abort has invalid percentage", func() {
		invalidPct := 101.0
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				FaultInjection: &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
					Percentage: invalidPct,
				},
			},
		}
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.InvalidPercentageError(invalidPct))))
	})

	It("should return error if FaultInjection Abort has invalid HTTP status", func() {
		invalidHttpStatus := int32(432)
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				FaultInjection: &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
					FaultInjectionType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Abort_{
						Abort: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Abort{
							ErrorType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Abort_HttpStatus{
								HttpStatus: invalidHttpStatus,
							},
						},
					},
					Percentage: 50,
				},
			},
		}
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.InvalidHttpStatus(invalidHttpStatus))))
	})

	It("should return error if FaultInjection Delay has invalid duration", func() {
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				FaultInjection: &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
					FaultInjectionType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_{
						Delay: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay{
							HttpDelayType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_FixedDelay{
								FixedDelay: &types1.Duration{Seconds: -1},
							},
						},
					},
					Percentage: 50,
				},
			},
		}
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.MinDurationError)))
	})

	It("should return error if RequestTimeout has an invalid duration", func() {
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				RequestTimeout: &types1.Duration{Seconds: 0, Nanos: 999999},
			},
		}
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.MinDurationError)))
	})

	It("should return error if CorsPolicy has an invalid max age duration", func() {
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				CorsPolicy: &zephyr_networking_types.TrafficPolicySpec_CorsPolicy{
					MaxAge: &types1.Duration{Seconds: 0, Nanos: 999999},
				},
			},
		}
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.MinDurationError)))
	})

	It("should return error if Mirror has invalid percentage", func() {
		serviceKey := client.ObjectKey{Name: "name", Namespace: "namespace"}
		serviceRef := &zephyr_core_types.ResourceRef{
			Name:      serviceKey.Name,
			Namespace: serviceKey.Namespace,
		}
		invalidPct := 101.0
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				Mirror: &zephyr_networking_types.TrafficPolicySpec_Mirror{
					Destination: serviceRef,
					Percentage:  invalidPct,
				},
			},
		}
		mockResourceSelector.
			EXPECT().
			GetMeshServiceByRefSelector(ctx, serviceKey.Name, serviceKey.Namespace, "").
			Return(nil, nil)
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.InvalidPercentageError(invalidPct))))
	})

	It("should return error if Mirror has destination that cannot be found", func() {
		serviceKey := client.ObjectKey{Name: "name", Namespace: "namespace"}
		serviceRef := &zephyr_core_types.ResourceRef{
			Name:      serviceKey.Name,
			Namespace: serviceKey.Namespace,
		}
		tp := &zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				Mirror: &zephyr_networking_types.TrafficPolicySpec_Mirror{
					Destination: serviceRef,
					Percentage:  50,
				},
			},
		}
		mockResourceSelector.
			EXPECT().
			GetMeshServiceByRefSelector(ctx, serviceKey.Name, serviceKey.Namespace, "").
			Return(nil, selector.MeshServiceNotFound(serviceKey.Name, serviceKey.Namespace, ""))
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(selector.MeshServiceNotFound(serviceKey.Name, serviceKey.Namespace, ""))))
	})
})

func expectSingleErrorOf(err error, expected error) {
	multierr, ok := err.(*multierror.Error)
	Expect(ok).To(BeTrue())
	Expect(multierr.Errors).To(HaveLen(1))
	Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(expected))
}
