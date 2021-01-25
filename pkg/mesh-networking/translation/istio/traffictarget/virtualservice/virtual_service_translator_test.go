package virtualservice_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	networkingv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2/types"
	settingsv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	mock_decorators "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mocks"
	mock_trafficpolicy "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget/virtualservice"
	mock_hostutils "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VirtualServiceTranslator", func() {
	var (
		ctrl                      *gomock.Controller
		mockClusterDomainRegistry *mock_hostutils.MockClusterDomainRegistry
		mockDecoratorFactory      *mock_decorators.MockFactory
		mockReporter              *mock_reporting.MockReporter
		mockDecorator             *mock_trafficpolicy.MockTrafficPolicyVirtualServiceDecorator
		virtualServiceTranslator  virtualservice.Translator
		in                        input.LocalSnapshot
		ctx                       = context.TODO()
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClusterDomainRegistry = mock_hostutils.NewMockClusterDomainRegistry(ctrl)
		mockDecoratorFactory = mock_decorators.NewMockFactory(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockDecorator = mock_trafficpolicy.NewMockTrafficPolicyVirtualServiceDecorator(ctrl)
		virtualServiceTranslator = virtualservice.NewTranslator(nil, mockClusterDomainRegistry, mockDecoratorFactory)
		in = input.NewInputLocalSnapshotManualBuilder("").AddSettingsMeshGlooSoloIov1Alpha2Settings(settingsv1alpha2.SettingsSlice{{}}).Build()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate and merge traffic policies by request matcher", func() {
		sourceSelectorLabels := map[string]string{"env": "dev"}
		sourceSelectorNamespaces := []string{"n1", "n2"}

		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
							{
								Port:     8080,
								Name:     "http1",
								Protocol: "http",
							},
							{
								Port:     9080,
								Name:     "http2",
								Protocol: "http",
							},
						},
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1alpha2.TrafficPolicySpec{
							SourceSelector: []*networkingv1alpha2.WorkloadSelector{
								{
									Labels:     sourceSelectorLabels,
									Namespaces: sourceSelectorNamespaces,
								},
							},
							HttpRequestMatchers: []*networkingv1alpha2.TrafficPolicySpec_HttpMatcher{
								{
									PathSpecifier: &networkingv1alpha2.TrafficPolicySpec_HttpMatcher_Exact{
										Exact: "path",
									},
									Method: &networkingv1alpha2.TrafficPolicySpec_HttpMethod{Method: types.HttpMethodValue_GET},
								},
								{
									Headers: []*networkingv1alpha2.TrafficPolicySpec_HeaderMatcher{
										{
											Name:        "name3",
											Value:       "[a-z]+",
											Regex:       true,
											InvertMatch: true,
										},
									},
									Method: &networkingv1alpha2.TrafficPolicySpec_HttpMethod{Method: types.HttpMethodValue_POST},
								},
							},
							Retries: &networkingv1alpha2.TrafficPolicySpec_RetryPolicy{
								Attempts: 5,
							},
						},
					},
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-2",
							Namespace: "tp-namespace-2",
						},
						Spec: &networkingv1alpha2.TrafficPolicySpec{
							SourceSelector: []*networkingv1alpha2.WorkloadSelector{
								{
									Labels:     sourceSelectorLabels,
									Namespaces: sourceSelectorNamespaces,
								},
							},
							HttpRequestMatchers: []*networkingv1alpha2.TrafficPolicySpec_HttpMatcher{
								{
									PathSpecifier: &networkingv1alpha2.TrafficPolicySpec_HttpMatcher_Exact{
										Exact: "path",
									},
									Method: &networkingv1alpha2.TrafficPolicySpec_HttpMethod{Method: types.HttpMethodValue_GET},
								},
								{
									Headers: []*networkingv1alpha2.TrafficPolicySpec_HeaderMatcher{
										{
											Name:        "name3",
											Value:       "[a-z]+",
											Regex:       true,
											InvertMatch: true,
										},
									},
									Method: &networkingv1alpha2.TrafficPolicySpec_HttpMethod{Method: types.HttpMethodValue_POST},
								},
							},
							FaultInjection: &networkingv1alpha2.TrafficPolicySpec_FaultInjection{
								FaultInjectionType: &networkingv1alpha2.TrafficPolicySpec_FaultInjection_Abort_{
									Abort: &networkingv1alpha2.TrafficPolicySpec_FaultInjection_Abort{
										HttpStatus: 500,
									},
								},
								Percentage: 50,
							},
						},
					},
				},
			},
		}

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(trafficTarget.Spec.GetKubeService().Ref.ClusterName, trafficTarget.Spec.GetKubeService().Ref).
			Return("local-hostname")

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		initializedMatchRequests := []*networkingv1alpha3spec.HTTPMatchRequest{
			{
				Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
				Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
				SourceLabels:    sourceSelectorLabels,
				SourceNamespace: sourceSelectorNamespaces[0],
			},
			{
				Method: &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
				WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
					"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
				},
				SourceLabels:    sourceSelectorLabels,
				SourceNamespace: sourceSelectorNamespaces[0],
			},
			{
				Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
				Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
				SourceLabels:    sourceSelectorLabels,
				SourceNamespace: sourceSelectorNamespaces[1],
			},
			{
				Method: &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
				WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
					"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
				},
				SourceLabels:    sourceSelectorLabels,
				SourceNamespace: sourceSelectorNamespaces[1],
			},
		}

		httpRetry := &networkingv1alpha3spec.HTTPRetry{
			Attempts: 5,
		}
		faultInjection := &networkingv1alpha3spec.HTTPFaultInjection{
			Abort: &networkingv1alpha3spec.HTTPFaultInjection_Abort{
				ErrorType: &networkingv1alpha3spec.HTTPFaultInjection_Abort_HttpStatus{
					HttpStatus: 500,
				},
				Percentage: &networkingv1alpha3spec.Percent{Value: 50},
			},
		}

		expectedHttpRoutes := []*networkingv1alpha3spec.HTTPRoute{
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
						Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[1],
						Port:            8080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 8080,
						},
					},
				}},
				Retries: httpRetry,
				Fault:   faultInjection,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
						Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[1],
						Port:            9080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 9080,
						},
					},
				}},
				Retries: httpRetry,
				Fault:   faultInjection,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
						Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[0],
						Port:            8080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 8080,
						},
					},
				}},
				Retries: httpRetry,
				Fault:   faultInjection,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
						Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[0],
						Port:            9080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 9080,
						},
					},
				}},
				Retries: httpRetry,
				Fault:   faultInjection,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[1],
						WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
							"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
						},
						Port: 8080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 8080,
						},
					},
				}},
				Retries: httpRetry,
				Fault:   faultInjection,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[1],
						WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
							"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
						},
						Port: 9080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 9080,
						},
					},
				}},
				Retries: httpRetry,
				Fault:   faultInjection,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[0],
						WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
							"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
						},
						Port: 8080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 8080,
						},
					},
				}},
				Retries: httpRetry,
				Fault:   faultInjection,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[0],
						WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
							"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
						},
						Port: 9080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 9080,
						},
					},
				}},
				Retries: httpRetry,
				Fault:   faultInjection,
			},
		}

		expectedVirtualService := &networkingv1alpha3.VirtualService{
			ObjectMeta: metautils.TranslatedObjectMeta(
				trafficTarget.Spec.GetKubeService().Ref,
				trafficTarget.Annotations,
			),
			Spec: networkingv1alpha3spec.VirtualService{
				Hosts: []string{"local-hostname"},
				Http:  expectedHttpRoutes,
			},
		}

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToVirtualService(
				trafficTarget.Status.AppliedTrafficPolicies[0],
				trafficTarget,
				nil,
				&networkingv1alpha3spec.HTTPRoute{
					Match: initializedMatchRequests,
				},
				gomock.Any(),
			).DoAndReturn(
			func(
				appliedPolicy *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
				service *discoveryv1alpha2.TrafficTarget,
				sourceMeshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation,
				output *networkingv1alpha3spec.HTTPRoute,
				registerField decorators.RegisterField,
			) error {
				output.Retries = httpRetry
				return nil
			}).
			Return(nil)

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToVirtualService(
				trafficTarget.Status.AppliedTrafficPolicies[1],
				trafficTarget,
				nil,
				&networkingv1alpha3spec.HTTPRoute{
					Match:   initializedMatchRequests,
					Retries: httpRetry,
				},
				gomock.Any(),
			).DoAndReturn(
			func(
				appliedPolicy *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
				service *discoveryv1alpha2.TrafficTarget,
				sourceMeshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation,
				output *networkingv1alpha3spec.HTTPRoute,
				registerField decorators.RegisterField,
			) error {
				output.Fault = faultInjection
				return nil
			}).
			Return(nil)

		virtualService := virtualServiceTranslator.Translate(ctx, in, trafficTarget, nil, mockReporter)
		Expect(virtualService).To(Equal(expectedVirtualService))
	})

	It("should translate for a federated TrafficTarget", func() {
		sourceSelectorLabels := map[string]string{"env": "dev"}
		sourceSelectorNamespaces := []string{"n1", "n2"}
		meshInstallation := &discoveryv1alpha2.MeshSpec_MeshInstallation{
			Namespace: "foobar",
			Cluster:   "mgmt-cluster",
		}

		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "traffic-target",
				ClusterName: "remote-cluster",
			},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
							{
								Port:     8080,
								Name:     "http1",
								Protocol: "http",
							},
							{
								Port:     9080,
								Name:     "http2",
								Protocol: "http",
							},
						},
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1alpha2.TrafficPolicySpec{
							SourceSelector: []*networkingv1alpha2.WorkloadSelector{
								{
									Labels:     sourceSelectorLabels,
									Namespaces: sourceSelectorNamespaces,
								},
							},
							HttpRequestMatchers: []*networkingv1alpha2.TrafficPolicySpec_HttpMatcher{
								{
									PathSpecifier: &networkingv1alpha2.TrafficPolicySpec_HttpMatcher_Exact{
										Exact: "path",
									},
									Method: &networkingv1alpha2.TrafficPolicySpec_HttpMethod{Method: types.HttpMethodValue_GET},
								},
								{
									Headers: []*networkingv1alpha2.TrafficPolicySpec_HeaderMatcher{
										{
											Name:        "name3",
											Value:       "[a-z]+",
											Regex:       true,
											InvertMatch: true,
										},
									},
									Method: &networkingv1alpha2.TrafficPolicySpec_HttpMethod{Method: types.HttpMethodValue_POST},
								},
							},
							Retries: &networkingv1alpha2.TrafficPolicySpec_RetryPolicy{
								Attempts: 5,
							},
						},
					},
				},
			},
		}

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(meshInstallation.Cluster, trafficTarget.Spec.GetKubeService().Ref).
			Return("local-hostname")

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		initializedMatchRequests := []*networkingv1alpha3spec.HTTPMatchRequest{
			{
				Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
				Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
				SourceLabels:    sourceSelectorLabels,
				SourceNamespace: sourceSelectorNamespaces[0],
			},
			{
				Method: &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
				WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
					"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
				},
				SourceLabels:    sourceSelectorLabels,
				SourceNamespace: sourceSelectorNamespaces[0],
			},
			{
				Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
				Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
				SourceLabels:    sourceSelectorLabels,
				SourceNamespace: sourceSelectorNamespaces[1],
			},
			{
				Method: &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
				WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
					"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
				},
				SourceLabels:    sourceSelectorLabels,
				SourceNamespace: sourceSelectorNamespaces[1],
			},
		}

		httpRetry := &networkingv1alpha3spec.HTTPRetry{
			Attempts: 5,
		}

		expectedHttpRoutes := []*networkingv1alpha3spec.HTTPRoute{
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
						Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[1],
						Port:            8080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 8080,
						},
					},
				}},
				Retries: httpRetry,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
						Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[1],
						Port:            9080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 9080,
						},
					},
				}},
				Retries: httpRetry,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
						Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[0],
						Port:            8080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 8080,
						},
					},
				}},
				Retries: httpRetry,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "GET"}},
						Uri:             &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "path"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[0],
						Port:            9080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 9080,
						},
					},
				}},
				Retries: httpRetry,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[1],
						WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
							"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
						},
						Port: 8080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 8080,
						},
					},
				}},
				Retries: httpRetry,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[1],
						WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
							"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
						},
						Port: 9080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 9080,
						},
					},
				}},
				Retries: httpRetry,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[0],
						WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
							"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
						},
						Port: 8080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 8080,
						},
					},
				}},
				Retries: httpRetry,
			},
			{
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{
					{
						Method:          &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: "POST"}},
						SourceLabels:    sourceSelectorLabels,
						SourceNamespace: sourceSelectorNamespaces[0],
						WithoutHeaders: map[string]*networkingv1alpha3spec.StringMatch{
							"name3": {MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: "[a-z]+"}},
						},
						Port: 9080,
					},
				},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination: &networkingv1alpha3spec.Destination{
						Host: "local-hostname",
						Port: &networkingv1alpha3spec.PortSelector{
							Number: 9080,
						},
					},
				}},
				Retries: httpRetry,
			},
		}

		expectedVirtualService := &networkingv1alpha3.VirtualService{
			ObjectMeta: metautils.FederatedObjectMeta(
				trafficTarget.Spec.GetKubeService().Ref,
				meshInstallation,
				trafficTarget.Annotations,
			),
			Spec: networkingv1alpha3spec.VirtualService{
				Hosts: []string{"local-hostname"},
				Http:  expectedHttpRoutes,
			},
		}

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToVirtualService(
				trafficTarget.Status.AppliedTrafficPolicies[0],
				trafficTarget,
				meshInstallation,
				&networkingv1alpha3spec.HTTPRoute{
					Match: initializedMatchRequests,
				},
				gomock.Any(),
			).DoAndReturn(
			func(
				appliedPolicy *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
				service *discoveryv1alpha2.TrafficTarget,
				sourceMeshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation,
				output *networkingv1alpha3spec.HTTPRoute,
				registerField decorators.RegisterField,
			) error {
				output.Retries = httpRetry
				return nil
			}).
			Return(nil)

		virtualService := virtualServiceTranslator.Translate(
			ctx,
			in,
			trafficTarget,
			meshInstallation,
			mockReporter,
		)
		Expect(virtualService).To(Equal(expectedVirtualService))
	})

	It("should not output a VirtualService if translated VirtualService has no HttpRoutes", func() {
		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
							{
								Port:     8080,
								Name:     "http1",
								Protocol: "http",
							},
						},
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1alpha2.TrafficPolicySpec{
							OutlierDetection: &networkingv1alpha2.TrafficPolicySpec_OutlierDetection{
								ConsecutiveErrors: 5,
							},
						},
					},
				},
			},
		}

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(trafficTarget.Spec.GetKubeService().Ref.ClusterName, trafficTarget.Spec.GetKubeService().Ref).
			Return("local-hostname")

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToVirtualService(
				trafficTarget.Status.AppliedTrafficPolicies[0],
				trafficTarget,
				nil,
				&networkingv1alpha3spec.HTTPRoute{},
				gomock.Any(),
			).
			Return(nil)

		virtualService := virtualServiceTranslator.Translate(ctx, in, trafficTarget, nil, mockReporter)
		Expect(virtualService).To(BeNil())
	})

	It("should not output an HttpRoute if TrafficPolicy's WorkloadSelector does not select source cluster", func() {
		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
							{
								Port:     8080,
								Name:     "http1",
								Protocol: "http",
							},
						},
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1alpha2.TrafficPolicySpec{
							SourceSelector: []*networkingv1alpha2.WorkloadSelector{
								{
									Clusters: []string{"foobar"},
								},
							},
							OutlierDetection: &networkingv1alpha2.TrafficPolicySpec_OutlierDetection{
								ConsecutiveErrors: 5,
							},
						},
					},
				},
			},
		}

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(trafficTarget.Spec.GetKubeService().Ref.ClusterName, trafficTarget.Spec.GetKubeService().Ref).
			Return("local-hostname")

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		virtualService := virtualServiceTranslator.Translate(ctx, in, trafficTarget, nil, mockReporter)
		Expect(virtualService).To(BeNil())
	})

	It("should not output a VirtualService if it contains a host that is already configured by an existing VirtualService not owned by Gloo Mesh", func() {
		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
							{
								Port:     8080,
								Name:     "http1",
								Protocol: "http",
							},
						},
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1alpha2.TrafficPolicySpec{
							SourceSelector: []*networkingv1alpha2.WorkloadSelector{
								{
									Clusters: []string{"traffic-target-cluster"},
								},
							},
							Retries: &networkingv1alpha2.TrafficPolicySpec_RetryPolicy{
								Attempts: 5,
							},
						},
					},
				},
			},
		}

		in = input.NewInputLocalSnapshotManualBuilder("").
			AddSettingsMeshGlooSoloIov1Alpha2Settings(settingsv1alpha2.SettingsSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaults.DefaultSettingsName,
						Namespace: defaults.DefaultPodNamespace,
					},
					Spec: settingsv1alpha2.SettingsSpec{},
				},
			}).
			Build()

		existingVirtualServices := v1alpha3sets.NewVirtualServiceSet(
			// user-supplied, should yield conflict error
			&networkingv1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "user-provided-vs",
					Namespace: "foo",
				},
				Spec: networkingv1alpha3spec.VirtualService{
					Hosts: []string{"*-hostname"},
				},
			},
		)

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(trafficTarget.Spec.GetKubeService().Ref.ClusterName, trafficTarget.Spec.GetKubeService().Ref).
			Return("local-hostname")

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToVirtualService(
				trafficTarget.Status.AppliedTrafficPolicies[0],
				trafficTarget,
				nil,
				gomock.Any(),
				gomock.Any(),
			).DoAndReturn(
			func(
				appliedPolicy *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
				service *discoveryv1alpha2.TrafficTarget,
				sourceMeshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation,
				output *networkingv1alpha3spec.HTTPRoute,
				registerField decorators.RegisterField,
			) error {
				output.Retries = &networkingv1alpha3spec.HTTPRetry{
					Attempts: 5,
				}
				return nil
			}).
			Return(nil)

		mockReporter.
			EXPECT().
			ReportTrafficPolicyToTrafficTarget(
				trafficTarget,
				trafficTarget.Status.AppliedTrafficPolicies[0].Ref,
				gomock.Any()).
			DoAndReturn(func(trafficTarget *discoveryv1alpha2.TrafficTarget, trafficPolicy ezkube.ResourceId, err error) {
				Expect(err).To(testutils.HaveInErrorChain(
					eris.Errorf("Unable to translate AppliedTrafficPolicies to VirtualService, applies to hosts %+v that are already configured by the existing VirtualService %s",
						[]string{"local-hostname"},
						sets.Key(existingVirtualServices.List()[0])),
				),
				)
			})

		virtualServiceTranslator = virtualservice.NewTranslator(existingVirtualServices, mockClusterDomainRegistry, mockDecoratorFactory)
		_ = virtualServiceTranslator.Translate(ctx, in, trafficTarget, nil, mockReporter)
	})
})
