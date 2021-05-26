package trafficpolicyutils

import (
	"github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	mock_hostutils "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils/mocks"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("Cors", func() {

	It("should set cors policy", func() {
		corsPolicy := &v1.TrafficPolicySpec_Policy_CorsPolicy{
			AllowOrigins: []*v1.StringMatch{
				{MatchType: &v1.StringMatch_Exact{Exact: "exact"}},
				{MatchType: &v1.StringMatch_Prefix{Prefix: "prefix"}},
				{MatchType: &v1.StringMatch_Regex{Regex: "regex"}},
			},
			AllowMethods:     []string{"GET", "POST"},
			AllowHeaders:     []string{"Header1", "Header2"},
			ExposeHeaders:    []string{"ExposedHeader1", "ExposedHeader2"},
			MaxAge:           &duration.Duration{Seconds: 1},
			AllowCredentials: &wrappers.BoolValue{Value: false},
		}
		expectedCorsPolicy := &v1alpha3.CorsPolicy{
			AllowOrigins: []*v1alpha3.StringMatch{
				{MatchType: &v1alpha3.StringMatch_Exact{Exact: "exact"}},
				{MatchType: &v1alpha3.StringMatch_Prefix{Prefix: "prefix"}},
				{MatchType: &v1alpha3.StringMatch_Regex{Regex: "regex"}},
			},
			AllowMethods:     []string{"GET", "POST"},
			AllowHeaders:     []string{"Header1", "Header2"},
			ExposeHeaders:    []string{"ExposedHeader1", "ExposedHeader2"},
			MaxAge:           &types.Duration{Seconds: 1},
			AllowCredentials: &types.BoolValue{Value: false},
		}
		corsResult, err := TranslateCorsPolicy(corsPolicy)
		Expect(err).ToNot(HaveOccurred())
		Expect(corsResult).To(Equal(expectedCorsPolicy))
	})
})

var _ = Describe("FaultInjection", func() {

	It("should set fault injection of type abort", func() {

		faultPolicy := &v1.TrafficPolicySpec_Policy_FaultInjection{
			FaultInjectionType: &v1.TrafficPolicySpec_Policy_FaultInjection_Abort_{
				Abort: &v1.TrafficPolicySpec_Policy_FaultInjection_Abort{
					HttpStatus: 404,
				},
			},
			Percentage: 50,
		}
		expectedFaultInjection := &v1alpha3.HTTPFaultInjection{
			Abort: &v1alpha3.HTTPFaultInjection_Abort{
				ErrorType:  &v1alpha3.HTTPFaultInjection_Abort_HttpStatus{HttpStatus: 404},
				Percentage: &v1alpha3.Percent{Value: 50},
			},
		}
		faultResult, err := TranslateFault(faultPolicy)
		Expect(err).ToNot(HaveOccurred())
		Expect(faultResult).To(Equal(expectedFaultInjection))
	})

	It("should set fault injection of type fixed delay", func() {
		faultPolicy := &v1.TrafficPolicySpec_Policy_FaultInjection{
			FaultInjectionType: &v1.TrafficPolicySpec_Policy_FaultInjection_FixedDelay{
				FixedDelay: &duration.Duration{Seconds: 2},
			},
			Percentage: 50,
		}
		expectedFaultInjection := &v1alpha3.HTTPFaultInjection{
			Delay: &v1alpha3.HTTPFaultInjection_Delay{
				HttpDelayType: &v1alpha3.HTTPFaultInjection_Delay_FixedDelay{FixedDelay: &types.Duration{Seconds: 2}},
				Percentage:    &v1alpha3.Percent{Value: 50},
			},
		}
		faultResult, err := TranslateFault(faultPolicy)
		Expect(err).ToNot(HaveOccurred())
		Expect(faultResult).To(Equal(expectedFaultInjection))
	})

	It("should return error if fault injection type not specified", func() {
		faultPolicy := &v1.TrafficPolicySpec_Policy_FaultInjection{
			Percentage: 50,
		}
		faultResult, err := TranslateFault(faultPolicy)
		Expect(err.Error()).To(ContainSubstring("FaultInjection type must be specified"))
		Expect(faultResult).To(BeNil())
	})
})

var _ = Describe("HeaderManipulation", func() {

	It("should set headers", func() {
		headerManipulationPolicy := &v1.HeaderManipulation{
			AppendRequestHeaders:  map[string]string{"a": "b"},
			RemoveRequestHeaders:  []string{"3", "4"},
			AppendResponseHeaders: map[string]string{"foo": "bar"},
			RemoveResponseHeaders: []string{"1", "2"},
		}
		expectedHeaderManipulation := &v1alpha3.Headers{
			Request: &v1alpha3.Headers_HeaderOperations{
				Add:    map[string]string{"a": "b"},
				Remove: []string{"3", "4"},
			},
			Response: &v1alpha3.Headers_HeaderOperations{
				Add:    map[string]string{"foo": "bar"},
				Remove: []string{"1", "2"},
			},
		}
		headerManipulationResult := TranslateHeaderManipulation(headerManipulationPolicy)
		Expect(headerManipulationResult).To(Equal(expectedHeaderManipulation))
	})
})

var _ = Describe("Mirror", func() {
	var (
		ctrl                      *gomock.Controller
		mockClusterDomainRegistry *mock_hostutils.MockClusterDomainRegistry
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClusterDomainRegistry = mock_hostutils.NewMockClusterDomainRegistry(ctrl)
	})

	It("should translate mirror", func() {
		destinations := discoveryv1sets.NewDestinationSet(
			&discoveryv1.Destination{
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "http1",
									Protocol: "http",
								},
							},
						},
					},
				},
			})
		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		mirrorPolicy := &v1.TrafficPolicySpec_Policy_Mirror{
			DestinationType: &v1.TrafficPolicySpec_Policy_Mirror_KubeService{
				KubeService: &skv2corev1.ClusterObjectRef{
					Name:        "mirror",
					Namespace:   "namespace",
					ClusterName: "local-cluster",
				},
			},
			Percentage: 50,
			// Not specifying port should default to the single port on the Mirror destination
		}

		localHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(originalService.ClusterName, mirrorPolicy.GetKubeService()).
			Return(localHostname)
		expectedMirror := &v1alpha3.Destination{
			Host: localHostname,
		}
		expectedMirrorPercentage := &v1alpha3.Percent{
			Value: mirrorPolicy.Percentage,
		}
		mirrorResult, mirrorPrecentResult, err := TranslateMirror(mirrorPolicy, originalService.ClusterName, mockClusterDomainRegistry, destinations)

		Expect(err).ToNot(HaveOccurred())
		Expect(mirrorResult).To(Equal(expectedMirror))
		Expect(mirrorPrecentResult).To(Equal(expectedMirrorPercentage))
	})

	It("should decorate mirror for federated Destination", func() {
		destinations := discoveryv1sets.NewDestinationSet(
			&discoveryv1.Destination{
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "http1",
									Protocol: "http",
								},
							},
						},
					},
				},
			})
		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		mirrorPolicy := &v1.TrafficPolicySpec_Policy_Mirror{
			DestinationType: &v1.TrafficPolicySpec_Policy_Mirror_KubeService{
				KubeService: &skv2corev1.ClusterObjectRef{
					Name:        "mirror",
					Namespace:   "namespace",
					ClusterName: "local-cluster",
				},
			},
			Percentage: 50,
			// Not specifying port should default to the single port on the Mirror destination
		}

		globalHostname := "name.namespace.svc.local-cluster.global"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(originalService.ClusterName, mirrorPolicy.GetKubeService()).
			Return(globalHostname)

		expectedMirror := &v1alpha3.Destination{
			Host: globalHostname,
		}
		expectedMirrorPercentage := &v1alpha3.Percent{
			Value: mirrorPolicy.Percentage,
		}

		mirrorResult, mirrorPrecentResult, err := TranslateMirror(mirrorPolicy, originalService.ClusterName, mockClusterDomainRegistry, destinations)

		Expect(err).ToNot(HaveOccurred())
		Expect(mirrorResult).To(Equal(expectedMirror))
		Expect(mirrorPrecentResult).To(Equal(expectedMirrorPercentage))
	})

	It("should throw error if mirror destination has multiple ports but port is not specified", func() {
		destinations := discoveryv1sets.NewDestinationSet(
			&discoveryv1.Destination{
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "http1",
									Protocol: "http",
								},
								{
									Port:     9081,
									Name:     "http2",
									Protocol: "http",
								},
							},
						},
					},
				},
			})
		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		mirrorPolicyMissingPort := &v1.TrafficPolicySpec_Policy_Mirror{
			DestinationType: &v1.TrafficPolicySpec_Policy_Mirror_KubeService{
				KubeService: &skv2corev1.ClusterObjectRef{
					Name:        "mirror",
					Namespace:   "namespace",
					ClusterName: "local-cluster",
				},
			},
			Percentage: 50,
			// Not specifying port should result in error
		}
		mirrorPolicyNonexistentPort := &v1.TrafficPolicySpec_Policy_Mirror{
			DestinationType: &v1.TrafficPolicySpec_Policy_Mirror_KubeService{
				KubeService: &skv2corev1.ClusterObjectRef{
					Name:        "mirror",
					Namespace:   "namespace",
					ClusterName: "local-cluster",
				},
			},
			Percentage: 50,
			Port:       1,
		}

		localHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(originalService.ClusterName, mirrorPolicyMissingPort.GetKubeService()).
			Return(localHostname).
			Times(2)

		_, _, err := TranslateMirror(mirrorPolicyMissingPort, originalService.ClusterName, mockClusterDomainRegistry, destinations)
		Expect(err.Error()).To(ContainSubstring("must provide port for mirror destination service"))

		_, _, err = TranslateMirror(mirrorPolicyNonexistentPort, originalService.ClusterName, mockClusterDomainRegistry, destinations)
		Expect(err.Error()).To(ContainSubstring("does not exist for mirror destination service"))

	})

	var _ = Describe("Retries", func() {
		It("should set retries", func() {
			retriesPolicy := &v1.TrafficPolicySpec_Policy_RetryPolicy{
				Attempts:      5,
				PerTryTimeout: &duration.Duration{Seconds: 2},
			}
			expectedRetries := &v1alpha3.HTTPRetry{
				Attempts:      5,
				PerTryTimeout: &types.Duration{Seconds: 2},
			}
			retriesResult := TranslateRetries(retriesPolicy)
			Expect(retriesResult).To(Equal(expectedRetries))
		})

	})

	var _ = Describe("Timeout", func() {
		It("should set retries", func() {
			timeoutPolicy := &duration.Duration{Seconds: 5}
			expectedTimeout := &types.Duration{Seconds: 5}
			timeoutResult := TranslateTimeout(timeoutPolicy)
			Expect(timeoutResult).To(Equal(expectedTimeout))
		})
	})

})
