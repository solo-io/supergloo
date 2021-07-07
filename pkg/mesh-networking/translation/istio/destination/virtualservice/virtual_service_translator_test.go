package virtualservice_test

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes/duration"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	mock_decorators "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mocks"
	mock_trafficpolicy "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/virtualservice"
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
		in = input.NewInputLocalSnapshotManualBuilder("").AddSettings(settingsv1.SettingsSlice{{}}).Build()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate and merge traffic policies by request matcher", func() {
		sourceSelectorLabels := map[string]string{"env": "dev"}
		sourceSelectorNamespaces := []string{"n1", "n2"}

		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
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
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*networkingv1.AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1.TrafficPolicySpec{
							SourceSelector: []*commonv1.WorkloadSelector{
								{
									KubeWorkloadMatcher: &commonv1.WorkloadSelector_KubeWorkloadMatcher{
										Labels:     sourceSelectorLabels,
										Namespaces: sourceSelectorNamespaces,
									},
								},
							},
							HttpRequestMatchers: []*networkingv1.HttpMatcher{
								{
									Uri: &commonv1.StringMatch{
										MatchType: &commonv1.StringMatch_Exact{
											Exact: "path",
										},
									},
									Method: "GET",
								},
								{
									Headers: []*networkingv1.HeaderMatcher{
										{
											Name:        "name3",
											Value:       "[a-z]+",
											Regex:       true,
											InvertMatch: true,
										},
									},
									Method: "POST",
								},
							},
							Policy: &networkingv1.TrafficPolicySpec_Policy{
								Retries: &networkingv1.TrafficPolicySpec_Policy_RetryPolicy{
									Attempts: 5,
								},
							},
						},
					},
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-2",
							Namespace: "tp-namespace-2",
						},
						Spec: &networkingv1.TrafficPolicySpec{
							SourceSelector: []*commonv1.WorkloadSelector{
								{
									KubeWorkloadMatcher: &commonv1.WorkloadSelector_KubeWorkloadMatcher{
										Labels:     sourceSelectorLabels,
										Namespaces: sourceSelectorNamespaces,
									},
								},
							},
							HttpRequestMatchers: []*networkingv1.HttpMatcher{
								{
									Uri: &commonv1.StringMatch{
										MatchType: &commonv1.StringMatch_Exact{
											Exact: "path",
										},
									},
									Method: "GET",
								},
								{
									Headers: []*networkingv1.HeaderMatcher{
										{
											Name:        "name3",
											Value:       "[a-z]+",
											Regex:       true,
											InvertMatch: true,
										},
									},
									Method: "POST",
								},
							},
							Policy: &networkingv1.TrafficPolicySpec_Policy{
								FaultInjection: &networkingv1.TrafficPolicySpec_Policy_FaultInjection{
									FaultInjectionType: &networkingv1.TrafficPolicySpec_Policy_FaultInjection_Abort_{
										Abort: &networkingv1.TrafficPolicySpec_Policy_FaultInjection_Abort{
											HttpStatus: 500,
										},
									},
									Percentage: 50,
								},
							},
						},
					},
				},
			},
		}

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(destination.Spec.GetKubeService().Ref.ClusterName, destination.Spec.GetKubeService().Ref).
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
				destination.Spec.GetKubeService().Ref,
				destination.Annotations,
			),
			Spec: networkingv1alpha3spec.VirtualService{
				Hosts: []string{"local-hostname"},
				Http:  expectedHttpRoutes,
			},
		}

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToVirtualService(
				destination.Status.AppliedTrafficPolicies[0],
				destination,
				nil,
				&networkingv1alpha3spec.HTTPRoute{
					Match: initializedMatchRequests,
				},
				gomock.Any(),
			).DoAndReturn(
			func(
				appliedPolicy *networkingv1.AppliedTrafficPolicy,
				service *discoveryv1.Destination,
				sourceMeshInstallation *discoveryv1.MeshInstallation,
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
				destination.Status.AppliedTrafficPolicies[1],
				destination,
				nil,
				&networkingv1alpha3spec.HTTPRoute{
					Match:   initializedMatchRequests,
					Retries: httpRetry,
				},
				gomock.Any(),
			).DoAndReturn(
			func(
				appliedPolicy *networkingv1.AppliedTrafficPolicy,
				service *discoveryv1.Destination,
				sourceMeshInstallation *discoveryv1.MeshInstallation,
				output *networkingv1alpha3spec.HTTPRoute,
				registerField decorators.RegisterField,
			) error {
				output.Fault = faultInjection
				return nil
			}).
			Return(nil)

		virtualService := virtualServiceTranslator.Translate(ctx, in, destination, nil, mockReporter)
		Expect(virtualService).To(Equal(expectedVirtualService))
	})

	It("should translate for a federated Destination", func() {
		sourceSelectorLabels := map[string]string{"env": "dev"}
		sourceSelectorNamespaces := []string{"n1", "n2"}
		meshInstallation := &discoveryv1.MeshInstallation{
			Namespace: "foobar",
			Cluster:   "mgmt-cluster",
		}

		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "traffic-target",
				ClusterName: "remote-cluster",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
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
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*networkingv1.AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1.TrafficPolicySpec{
							SourceSelector: []*commonv1.WorkloadSelector{
								{
									KubeWorkloadMatcher: &commonv1.WorkloadSelector_KubeWorkloadMatcher{
										Labels:     sourceSelectorLabels,
										Namespaces: sourceSelectorNamespaces,
									},
								},
							},
							HttpRequestMatchers: []*networkingv1.HttpMatcher{
								{
									Uri: &commonv1.StringMatch{
										MatchType: &commonv1.StringMatch_Exact{
											Exact: "path",
										},
									},
									Method: "GET",
								},
								{
									Headers: []*networkingv1.HeaderMatcher{
										{
											Name:        "name3",
											Value:       "[a-z]+",
											Regex:       true,
											InvertMatch: true,
										},
									},
									Method: "POST",
								},
							},
							Policy: &networkingv1.TrafficPolicySpec_Policy{
								Retries: &networkingv1.TrafficPolicySpec_Policy_RetryPolicy{
									Attempts: 5,
								},
							},
						},
					},
				},
			},
		}

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(meshInstallation.Cluster, destination.Spec.GetKubeService().Ref).
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
				destination.Spec.GetKubeService().Ref,
				meshInstallation,
				destination.Annotations,
			),
			Spec: networkingv1alpha3spec.VirtualService{
				Hosts: []string{"local-hostname"},
				Http:  expectedHttpRoutes,
			},
		}

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToVirtualService(
				destination.Status.AppliedTrafficPolicies[0],
				destination,
				meshInstallation,
				&networkingv1alpha3spec.HTTPRoute{
					Match: initializedMatchRequests,
				},
				gomock.Any(),
			).DoAndReturn(
			func(
				appliedPolicy *networkingv1.AppliedTrafficPolicy,
				service *discoveryv1.Destination,
				sourceMeshInstallation *discoveryv1.MeshInstallation,
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
			destination,
			meshInstallation,
			mockReporter,
		)
		Expect(virtualService).To(Equal(expectedVirtualService))
	})

	It("should not output a VirtualService if translated VirtualService has no HttpRoutes", func() {
		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
							{
								Port:     8080,
								Name:     "http1",
								Protocol: "http",
							},
						},
					},
				},
			},
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*networkingv1.AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1.TrafficPolicySpec{
							Policy: &networkingv1.TrafficPolicySpec_Policy{
								OutlierDetection: &networkingv1.TrafficPolicySpec_Policy_OutlierDetection{
									ConsecutiveErrors: 5,
								},
							},
						},
					},
				},
			},
		}

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(destination.Spec.GetKubeService().Ref.ClusterName, destination.Spec.GetKubeService().Ref).
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
				destination.Status.AppliedTrafficPolicies[0],
				destination,
				nil,
				&networkingv1alpha3spec.HTTPRoute{},
				gomock.Any(),
			).
			Return(nil)

		virtualService := virtualServiceTranslator.Translate(ctx, in, destination, nil, mockReporter)
		Expect(virtualService).To(BeNil())
	})

	It("should not output an HttpRoute if TrafficPolicy's WorkloadSelector does not select source cluster", func() {
		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
							{
								Port:     8080,
								Name:     "http1",
								Protocol: "http",
							},
						},
					},
				},
			},
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*networkingv1.AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1.TrafficPolicySpec{
							SourceSelector: []*commonv1.WorkloadSelector{
								{
									KubeWorkloadMatcher: &commonv1.WorkloadSelector_KubeWorkloadMatcher{
										Clusters: []string{"foobar"},
									},
								},
							},
							Policy: &networkingv1.TrafficPolicySpec_Policy{
								OutlierDetection: &networkingv1.TrafficPolicySpec_Policy_OutlierDetection{
									ConsecutiveErrors: 5,
								},
							},
						},
					},
				},
			},
		}

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(destination.Spec.GetKubeService().Ref.ClusterName, destination.Spec.GetKubeService().Ref).
			Return("local-hostname")

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		virtualService := virtualServiceTranslator.Translate(ctx, in, destination, nil, mockReporter)
		Expect(virtualService).To(BeNil())
	})

	It("should not output a VirtualService if it contains a host that is already configured by an existing VirtualService not owned by Gloo Mesh", func() {
		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
							{
								Port:     8080,
								Name:     "http1",
								Protocol: "http",
							},
						},
					},
				},
			},
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*networkingv1.AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1.TrafficPolicySpec{
							SourceSelector: []*commonv1.WorkloadSelector{
								{
									KubeWorkloadMatcher: &commonv1.WorkloadSelector_KubeWorkloadMatcher{
										Clusters: []string{"traffic-target-cluster"},
									},
								},
							},
							Policy: &networkingv1.TrafficPolicySpec_Policy{
								Retries: &networkingv1.TrafficPolicySpec_Policy_RetryPolicy{
									Attempts: 5,
								},
							},
						},
					},
				},
			},
		}

		in = input.NewInputLocalSnapshotManualBuilder("").
			AddSettings(settingsv1.SettingsSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaults.DefaultSettingsName,
						Namespace: defaults.DefaultPodNamespace,
					},
					Spec: settingsv1.SettingsSpec{},
				},
			}).
			Build()

		existingVirtualServices := v1alpha3sets.NewVirtualServiceSet(
			// user-supplied, should yield conflict error
			&networkingv1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "user-provided-vs",
					Namespace:   "foo",
					ClusterName: "traffic-target-cluster",
				},
				Spec: networkingv1alpha3spec.VirtualService{
					Hosts: []string{"*-hostname"},
				},
			},
		)

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(destination.Spec.GetKubeService().Ref.ClusterName, destination.Spec.GetKubeService().Ref).
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
				destination.Status.AppliedTrafficPolicies[0],
				destination,
				nil,
				gomock.Any(),
				gomock.Any(),
			).DoAndReturn(
			func(
				appliedPolicy *networkingv1.AppliedTrafficPolicy,
				service *discoveryv1.Destination,
				sourceMeshInstallation *discoveryv1.MeshInstallation,
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
			ReportTrafficPolicyToDestination(
				destination,
				destination.Status.AppliedTrafficPolicies[0].Ref,
				gomock.Any()).
			DoAndReturn(func(destination *discoveryv1.Destination, trafficPolicy ezkube.ResourceId, err error) {
				Expect(err).To(testutils.HaveInErrorChain(
					eris.Errorf("Unable to translate AppliedTrafficPolicies to VirtualService, applies to hosts %+v that are already configured by the existing VirtualService %s",
						[]string{"local-hostname"},
						sets.Key(existingVirtualServices.List()[0])),
				),
				)
			})

		virtualServiceTranslator = virtualservice.NewTranslator(existingVirtualServices, mockClusterDomainRegistry, mockDecoratorFactory)
		_ = virtualServiceTranslator.Translate(ctx, in, destination, nil, mockReporter)
	})

	It("should correctly order HttpRoutes according to presence of HttpMatchRequest", func() {
		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
							{
								Port:     8080,
								Name:     "http1",
								Protocol: "http",
							},
						},
					},
				},
			},
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*networkingv1.AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1.TrafficPolicySpec{
							DestinationSelector: []*commonv1.DestinationSelector{
								{
									KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
										Services: []*v1.ClusterObjectRef{
											{
												Name:        "reviews",
												Namespace:   "bookinfo",
												ClusterName: "mgmt-cluster",
											},
										},
									},
								},
							},
							HttpRequestMatchers: []*networkingv1.HttpMatcher{
								{
									Headers: []*networkingv1.HeaderMatcher{
										{
											Name:  "user-agent",
											Value: "'.*Firefox.*'",
											Regex: true,
										},
									},
								},
							},
							Policy: &networkingv1.TrafficPolicySpec_Policy{
								Retries: &networkingv1.TrafficPolicySpec_Policy_RetryPolicy{
									Attempts: 5,
								},
							},
						},
					},
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &networkingv1.TrafficPolicySpec{
							DestinationSelector: []*commonv1.DestinationSelector{
								{
									KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
										Services: []*v1.ClusterObjectRef{
											{
												Name:        "reviews",
												Namespace:   "bookinfo",
												ClusterName: "mgmt-cluster",
											},
										},
									},
								},
							},
							Policy: &networkingv1.TrafficPolicySpec_Policy{
								Retries: &networkingv1.TrafficPolicySpec_Policy_RetryPolicy{
									Attempts: 1,
								},
							},
						},
					},
				},
			},
		}

		in = input.NewInputLocalSnapshotManualBuilder("").
			AddSettings(settingsv1.SettingsSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaults.DefaultSettingsName,
						Namespace: defaults.DefaultPodNamespace,
					},
					Spec: settingsv1.SettingsSpec{},
				},
			}).
			Build()

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(destination.Spec.GetKubeService().Ref.ClusterName, destination.Spec.GetKubeService().Ref).
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
				destination.Status.AppliedTrafficPolicies[0],
				destination,
				nil,
				gomock.Any(),
				gomock.Any(),
			).DoAndReturn(
			func(
				appliedPolicy *networkingv1.AppliedTrafficPolicy,
				service *discoveryv1.Destination,
				sourceMeshInstallation *discoveryv1.MeshInstallation,
				output *networkingv1alpha3spec.HTTPRoute,
				registerField decorators.RegisterField,
			) error {
				output.Retries = &networkingv1alpha3spec.HTTPRetry{
					Attempts: destination.Status.AppliedTrafficPolicies[0].Spec.Policy.Retries.Attempts,
				}
				return nil
			}).
			Return(nil)

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToVirtualService(
				destination.Status.AppliedTrafficPolicies[1],
				destination,
				nil,
				gomock.Any(),
				gomock.Any(),
			).DoAndReturn(
			func(
				appliedPolicy *networkingv1.AppliedTrafficPolicy,
				service *discoveryv1.Destination,
				sourceMeshInstallation *discoveryv1.MeshInstallation,
				output *networkingv1alpha3spec.HTTPRoute,
				registerField decorators.RegisterField,
			) error {
				output.Retries = &networkingv1alpha3spec.HTTPRetry{
					Attempts: destination.Status.AppliedTrafficPolicies[1].Spec.Policy.Retries.Attempts,
				}
				return nil
			}).
			Return(nil)

		expectedVirtualService := &networkingv1alpha3.VirtualService{
			ObjectMeta: metautils.TranslatedObjectMeta(
				destination.Spec.GetKubeService().Ref,
				destination.Annotations,
			),
			Spec: networkingv1alpha3spec.VirtualService{
				Hosts: []string{"local-hostname"},
				Http: []*networkingv1alpha3spec.HTTPRoute{
					{
						Match: []*networkingv1alpha3spec.HTTPMatchRequest{
							{
								Headers: map[string]*networkingv1alpha3spec.StringMatch{
									"user-agent": {
										MatchType: &networkingv1alpha3spec.StringMatch_Regex{
											Regex: "'.*Firefox.*'",
										},
									},
								},
								Port: 8080,
							},
						},
						Route: []*networkingv1alpha3spec.HTTPRouteDestination{
							{
								Destination: &networkingv1alpha3spec.Destination{
									Host: "local-hostname",
									Port: &networkingv1alpha3spec.PortSelector{
										Number: 8080,
									},
								},
							},
						},
						Retries: &networkingv1alpha3spec.HTTPRetry{
							Attempts: 5,
						},
					},
					{
						Match: []*networkingv1alpha3spec.HTTPMatchRequest{
							{
								Port: 8080,
							},
						},
						Route: []*networkingv1alpha3spec.HTTPRouteDestination{
							{
								Destination: &networkingv1alpha3spec.Destination{
									Host: "local-hostname",
									Port: &networkingv1alpha3spec.PortSelector{
										Number: 8080,
									},
								},
							},
						},
						Retries: &networkingv1alpha3spec.HTTPRetry{
							Attempts: 1,
						},
					},
				},
			},
		}

		virtualService := virtualServiceTranslator.Translate(ctx, in, destination, nil, mockReporter)
		Expect(virtualService).To(Equal(expectedVirtualService))
	})

	It("should add HttpMatchRequest with a matcher and route for each port, without overwriting intended route port", func() {
		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
						Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
							{
								Port:     9000,
								Name:     "http2",
								Protocol: "http",
							},
							{
								Port:     9443,
								Name:     "http2",
								Protocol: "https",
							},
						},
					},
				},
			},
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*networkingv1.AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "backend-timeout",
							Namespace: "gloo-mesh",
						},
						Spec: &networkingv1.TrafficPolicySpec{
							DestinationSelector: []*commonv1.DestinationSelector{
								{
									KubeServiceMatcher: &commonv1.DestinationSelector_KubeServiceMatcher{
										Labels: map[string]string{"app": "backend"},
									},
								},
							},
							Policy: &networkingv1.TrafficPolicySpec_Policy{RequestTimeout: &duration.Duration{Seconds: 1}},
						},
					},
				},
			},
		}

		virtualMesh := &networkingv1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "virtual-mesh",
				Namespace: "gloo-mesh",
			},
			Spec: networkingv1.VirtualMeshSpec{
				Meshes: []*v1.ObjectRef{
					{Name: "istiod-istio-system-cluster-0", Namespace: "gloo-mesh"},
					{Name: "istiod-istio-system-cluster-1", Namespace: "gloo-mesh"},
				},
			},
		}

		in = input.NewInputLocalSnapshotManualBuilder("").
			AddSettings(settingsv1.SettingsSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaults.DefaultSettingsName,
						Namespace: defaults.DefaultPodNamespace,
					},
					Spec: settingsv1.SettingsSpec{},
				},
			}).
			AddVirtualMeshes(networkingv1.VirtualMeshSlice{virtualMesh}).
			Build()

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(destination.Spec.GetKubeService().Ref.ClusterName, destination.Spec.GetKubeService().Ref).
			Return("local-hostname")

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		routeDestination := &networkingv1alpha3spec.HTTPRouteDestination{
			Destination: &networkingv1alpha3spec.Destination{
				Host: "traffic-shift-host",
				// this port should not be overwritten
				Port: &networkingv1alpha3spec.PortSelector{
					Number: 1234,
				},
			},
		}

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToVirtualService(
				destination.Status.AppliedTrafficPolicies[0],
				destination,
				nil,
				gomock.Any(),
				gomock.Any(),
			).DoAndReturn(
			func(
				appliedPolicy *networkingv1.AppliedTrafficPolicy,
				service *discoveryv1.Destination,
				sourceMeshInstallation *discoveryv1.MeshInstallation,
				output *networkingv1alpha3spec.HTTPRoute,
				registerField decorators.RegisterField,
			) error {
				reqTimeout := destination.Status.AppliedTrafficPolicies[0].Spec.Policy.RequestTimeout
				output.Timeout = &types.Duration{Seconds: reqTimeout.GetSeconds(), Nanos: reqTimeout.GetNanos()}
				output.Route = []*networkingv1alpha3spec.HTTPRouteDestination{routeDestination}
				return nil
			}).
			Return(nil)

		expectedVirtualService := &networkingv1alpha3.VirtualService{
			ObjectMeta: metautils.TranslatedObjectMeta(
				destination.Spec.GetKubeService().Ref,
				destination.Annotations,
			),
			Spec: networkingv1alpha3spec.VirtualService{
				Hosts: []string{"local-hostname"},
				Http: []*networkingv1alpha3spec.HTTPRoute{
					{
						Route: []*networkingv1alpha3spec.HTTPRouteDestination{
							routeDestination,
						},
						Match: []*networkingv1alpha3spec.HTTPMatchRequest{
							{
								Port: 9000,
							},
						},
						Timeout: &types.Duration{
							Seconds: 1,
						},
					},
					{
						Route: []*networkingv1alpha3spec.HTTPRouteDestination{
							routeDestination,
						},
						Match: []*networkingv1alpha3spec.HTTPMatchRequest{
							{
								Port: 9443,
							},
						},
						Timeout: &types.Duration{
							Seconds: 1,
						},
					},
				},
			},
		}

		virtualService := virtualServiceTranslator.Translate(ctx, in, destination, nil, mockReporter)
		Expect(virtualService).To(Equal(expectedVirtualService))
	})
})
