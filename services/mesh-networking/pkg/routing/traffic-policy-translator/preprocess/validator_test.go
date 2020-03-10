package preprocess_test

import (
	"context"

	types1 "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
	mock_preprocess "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess/mocks"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Validator", func() {
	var (
		ctrl                    *gomock.Controller
		ctx                     context.Context
		mockMeshServiceClient   *mock_core.MockMeshServiceClient
		mockMeshServiceSelector *mock_preprocess.MockMeshServiceSelector
		validator               preprocess.TrafficPolicyValidator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshServiceClient = mock_core.NewMockMeshServiceClient(ctrl)
		mockMeshServiceSelector = mock_preprocess.NewMockMeshServiceSelector(ctrl)
		validator = preprocess.NewTrafficPolicyValidator(mockMeshServiceClient, mockMeshServiceSelector)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should return error if Destination cannot be found", func() {
		name := "name"
		namespace := "namespace"
		cluster := "cluster"
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				DestinationSelector: &core_types.Selector{
					Refs: []*core_types.ResourceRef{
						{
							Name:      name,
							Namespace: namespace,
							Cluster:   &types1.StringValue{Value: cluster},
						},
					},
				},
			},
		}
		mockMeshServiceSelector.
			EXPECT().
			GetBackingMeshService(ctx, name, namespace, cluster).
			Return(nil, preprocess.MeshServiceNotFound(name, namespace, cluster))
		err := validator.Validate(ctx, tp)
		expectSingleErrorOf(err, preprocess.MeshServiceNotFound(name, namespace, cluster))
	})

	It("should return error for RetryPolicy with negative num attempts", func() {
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				Retries: &networking_types.RetryPolicy{Attempts: -1},
			},
		}
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.InvalidRetryPolicyNumAttempts(-1))))
	})

	It("should return error for RetryPolicy with negative per retry timeout", func() {
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				Retries: &networking_types.RetryPolicy{PerTryTimeout: &types1.Duration{Seconds: -1}},
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
		serviceRef := &core_types.ResourceRef{
			Name:      name,
			Namespace: namespace,
			Cluster:   &types1.StringValue{Value: cluster},
		}
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				TrafficShift: &networking_types.MultiDestination{
					Destinations: []*networking_types.MultiDestination_WeightedDestination{
						{
							Destination: serviceRef,
							Weight:      1,
						},
					},
				},
			},
		}
		mockMeshServiceSelector.
			EXPECT().
			GetBackingMeshService(ctx, name, namespace, cluster).
			Return(nil, preprocess.MeshServiceNotFound(name, namespace, cluster))
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.MeshServiceNotFound(name, namespace, cluster))))
	})

	It("should return error if TrafficShift has subsets that can't be found", func() {
		name := "name"
		namespace := "namespace"
		cluster := ""
		serviceRef := &core_types.ResourceRef{
			Name:      name,
			Namespace: namespace,
			Cluster:   &types1.StringValue{Value: cluster},
		}
		subset := map[string]string{"env": "dev", "version": "v1"}
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				TrafficShift: &networking_types.MultiDestination{
					Destinations: []*networking_types.MultiDestination_WeightedDestination{
						{
							Destination: serviceRef,
							Subset:      subset,
							Weight:      1,
						},
					},
				},
			},
		}
		backingMeshService := &v1alpha1.MeshService{Spec: types.MeshServiceSpec{
			Subsets: map[string]*types.MeshServiceSpec_Subset{
				"env":     {Values: []string{"dev", "prod"}},
				"version": {Values: []string{"v2", "v3"}},
			},
		}}
		mockMeshServiceSelector.
			EXPECT().
			GetBackingMeshService(ctx, serviceRef.GetName(), serviceRef.GetNamespace(), serviceRef.GetCluster().GetValue()).
			Return(backingMeshService, nil)
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.SubsetSelectorNotFound(backingMeshService, "version", "v1"))))
	})

	It("should return error if FaultInjection Abort has invalid percentage", func() {
		invalidPct := 101.0
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				FaultInjection: &networking_types.FaultInjection{
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
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				FaultInjection: &networking_types.FaultInjection{
					FaultInjectionType: &networking_types.FaultInjection_Abort_{
						Abort: &networking_types.FaultInjection_Abort{
							ErrorType: &networking_types.FaultInjection_Abort_HttpStatus{
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
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				FaultInjection: &networking_types.FaultInjection{
					FaultInjectionType: &networking_types.FaultInjection_Delay_{
						Delay: &networking_types.FaultInjection_Delay{
							HttpDelayType: &networking_types.FaultInjection_Delay_FixedDelay{
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
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				RequestTimeout: &types1.Duration{Seconds: 0, Nanos: 999999},
			},
		}
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.MinDurationError)))
	})

	It("should return error if CorsPolicy has an invalid max age duration", func() {
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				CorsPolicy: &networking_types.CorsPolicy{
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
		serviceRef := &core_types.ResourceRef{
			Name:      serviceKey.Name,
			Namespace: serviceKey.Namespace,
		}
		invalidPct := 101.0
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				Mirror: &networking_types.Mirror{
					Destination: serviceRef,
					Percentage:  invalidPct,
				},
			},
		}
		mockMeshServiceSelector.
			EXPECT().
			GetBackingMeshService(ctx, serviceKey.Name, serviceKey.Namespace, "").
			Return(nil, nil)
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.InvalidPercentageError(invalidPct))))
	})

	It("should return error if Mirror has destination that cannot be found", func() {
		serviceKey := client.ObjectKey{Name: "name", Namespace: "namespace"}
		serviceRef := &core_types.ResourceRef{
			Name:      serviceKey.Name,
			Namespace: serviceKey.Namespace,
		}
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: networking_types.TrafficPolicySpec{
				Mirror: &networking_types.Mirror{
					Destination: serviceRef,
					Percentage:  50,
				},
			},
		}
		mockMeshServiceSelector.
			EXPECT().
			GetBackingMeshService(ctx, serviceKey.Name, serviceKey.Namespace, "").
			Return(nil, preprocess.MeshServiceNotFound(serviceKey.Name, serviceKey.Namespace, ""))
		err := validator.Validate(ctx, tp)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(ContainElement(testutils.HaveInErrorChain(preprocess.MeshServiceNotFound(serviceKey.Name, serviceKey.Namespace, ""))))
	})
})

func expectSingleErrorOf(err error, expected error) {
	multierr, ok := err.(*multierror.Error)
	Expect(ok).To(BeTrue())
	Expect(multierr.Errors).To(HaveLen(1))
	Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(expected))
}
