package virtualservice_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input/test"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/types"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	mock_decorators "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators/mocks"
	mock_trafficpolicy "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/virtualservice"
	mock_hostutils "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
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
		mockDecorator             *mock_trafficpolicy.MockVirtualServiceDecorator
		virtualServiceTranslator  virtualservice.Translator
		in                        input.Snapshot
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClusterDomainRegistry = mock_hostutils.NewMockClusterDomainRegistry(ctrl)
		mockDecoratorFactory = mock_decorators.NewMockFactory(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockDecorator = mock_trafficpolicy.NewMockVirtualServiceDecorator(ctrl)
		virtualServiceTranslator = virtualservice.NewTranslator(mockClusterDomainRegistry, mockDecoratorFactory)
		in = test.NewInputSnapshotManualBuilder("").Build()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate", func() {
		sourceSelectorLabels := map[string]string{"env": "dev"}
		sourceSelectorNamespaces := []string{"n1", "n2"}

		meshService := &discoveryv1alpha2.MeshService{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mesh-service",
			},
			Spec: discoveryv1alpha2.MeshServiceSpec{
				Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
					KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "mesh-service",
							Namespace:   "mesh-service-namespace",
							ClusterName: "mesh-service-cluster",
						},
						Ports: []*discoveryv1alpha2.MeshServiceSpec_KubeService_KubeServicePort{
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
			Status: discoveryv1alpha2.MeshServiceStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
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
						},
					},
				},
			},
		}

		mockClusterDomainRegistry.
			EXPECT().
			GetServiceLocalFQDN(meshService.Spec.GetKubeService().Ref).
			Return("local-hostname").
			Times(2)

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

		// For each HTTPMatchRequest, for each MeshService KubeService port, produce an networkingv1alpha3spec.HTTPRoute
		var expectedHttpRoutes []*networkingv1alpha3spec.HTTPRoute
		for _, port := range meshService.Spec.GetKubeService().Ports {
			for _, matchRequest := range initializedMatchRequests {
				matchRequestWithPort := matchRequest.DeepCopy()
				matchRequestWithPort.Port = port.Port
				route := &networkingv1alpha3spec.HTTPRoute{
					Match: []*networkingv1alpha3spec.HTTPMatchRequest{
						matchRequestWithPort,
					},
					Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
						Destination: &networkingv1alpha3spec.Destination{
							Host: "local-hostname",
							Port: &networkingv1alpha3spec.PortSelector{
								Number: port.Port,
							},
						},
					}},
				}
				expectedHttpRoutes = append(expectedHttpRoutes, route)
			}
		}

		expectedVirtualService := &networkingv1alpha3.VirtualService{
			ObjectMeta: metautils.TranslatedObjectMeta(
				meshService.Spec.GetKubeService().Ref,
				meshService.Annotations,
			),
			Spec: networkingv1alpha3spec.VirtualService{
				Hosts: []string{"local-hostname"},
				Http:  expectedHttpRoutes,
			},
		}

		mockDecorator.
			EXPECT().
			ApplyToVirtualService(
				meshService.Status.AppliedTrafficPolicies[0],
				meshService,
				&networkingv1alpha3spec.HTTPRoute{
					Match: initializedMatchRequests,
				},
				gomock.Any(),
			).
			Return(nil)

		virtualService := virtualServiceTranslator.Translate(in, meshService, mockReporter)
		Expect(virtualService).To(Equal(expectedVirtualService))
	})
})
