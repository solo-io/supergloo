package istio_translator_test

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	istio_networking "github.com/solo-io/service-mesh-hub/pkg/api/istio/networking/v1alpha3"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/selector/mocks"
	mock_mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s/mocks"
	istio_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/istio-translator"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_istio_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/istio/networking/v1beta1"
	api_v1alpha3 "istio.io/api/networking/v1alpha3"
	client_v1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type testContext struct {
	clusterName            string
	meshObjKey             client.ObjectKey
	meshServiceObjKey      client.ObjectKey
	kubeServiceObjKey      client.ObjectKey
	mesh                   *zephyr_discovery.Mesh
	meshService            *zephyr_discovery.MeshService
	trafficPolicy          []*zephyr_networking.TrafficPolicy
	computedVirtualService *client_v1alpha3.VirtualService
	baseMatchRequest       *api_v1alpha3.HTTPMatchRequest
	defaultRoute           []*api_v1alpha3.HTTPRouteDestination
}

var _ = Describe("IstioTranslator", func() {
	var (
		ctrl                         *gomock.Controller
		istioTrafficPolicyTranslator istio_translator.IstioTranslator
		ctx                          context.Context
		mockDynamicClientGetter      *mock_mc_manager.MockDynamicClientGetter
		mockMeshClient               *mock_core.MockMeshClient
		mockMeshServiceClient        *mock_core.MockMeshServiceClient
		mockVirtualServiceClient     *mock_istio_networking.MockVirtualServiceClient
		mockDestinationRuleClient    *mock_istio_networking.MockDestinationRuleClient
		mockResourceSelector         *mock_selector.MockResourceSelector
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockDynamicClientGetter = mock_mc_manager.NewMockDynamicClientGetter(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockMeshServiceClient = mock_core.NewMockMeshServiceClient(ctrl)
		mockVirtualServiceClient = mock_istio_networking.NewMockVirtualServiceClient(ctrl)
		mockResourceSelector = mock_selector.NewMockResourceSelector(ctrl)
		mockDestinationRuleClient = mock_istio_networking.NewMockDestinationRuleClient(ctrl)
		istioTrafficPolicyTranslator = istio_translator.NewIstioTrafficPolicyTranslator(
			mockDynamicClientGetter,
			mockMeshClient,
			mockMeshServiceClient,
			mockResourceSelector,
			func(client client.Client) istio_networking.VirtualServiceClient {
				return mockVirtualServiceClient
			},
			func(client client.Client) istio_networking.DestinationRuleClient {
				return mockDestinationRuleClient
			},
		)
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Context("should translate TrafficPolicies into VirtualService and DestinationRule and upsert", func() {
		setupTestContext := func() *testContext {
			clusterName := "clusterName"
			sourceNamespace := "source-namespace"
			meshObjKey := client.ObjectKey{Name: "mesh-name", Namespace: "mesh-namespace"}
			meshServiceObjKey := client.ObjectKey{Name: "mesh-service-name", Namespace: "mesh-service-namespace"}
			kubeServiceObjKey := client.ObjectKey{Name: "kube-service-name", Namespace: "kube-service-namespace"}
			meshServiceFederationMCDnsName := "multiclusterDNSname"
			meshService := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:        meshServiceObjKey.Name,
					Namespace:   meshServiceObjKey.Namespace,
					ClusterName: clusterName,
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      meshObjKey.Name,
						Namespace: meshObjKey.Namespace,
					},
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      kubeServiceObjKey.Name,
							Namespace: kubeServiceObjKey.Namespace,
							Cluster:   clusterName,
						},
						Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
							{
								Port: 9080,
								Name: "http",
							},
						},
					},
					Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{
						MulticlusterDnsName: meshServiceFederationMCDnsName,
					},
				},
			}
			mesh := &zephyr_discovery.Mesh{
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio{
						Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{},
					},
				},
			}
			trafficPolicy := []*zephyr_networking.TrafficPolicy{{
				Spec: zephyr_networking_types.TrafficPolicySpec{
					SourceSelector: &zephyr_core_types.WorkloadSelector{
						Namespaces: []string{sourceNamespace},
					},
					HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
						{}, {}, {},
					},
				}},
			}
			baseMatchRequest := &api_v1alpha3.HTTPMatchRequest{SourceNamespace: sourceNamespace}
			defaultRoute := []*api_v1alpha3.HTTPRouteDestination{
				{
					Destination: &api_v1alpha3.Destination{
						Host: meshService.Spec.GetKubeService().GetRef().GetName(),
						Port: &api_v1alpha3.PortSelector{
							Number: 9080,
						},
					},
				},
			}
			computedVirtualService := &client_v1alpha3.VirtualService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      meshService.Spec.GetKubeService().GetRef().GetName(),
					Namespace: meshService.Spec.GetKubeService().GetRef().GetNamespace(),
				},
				Spec: api_v1alpha3.VirtualService{
					Hosts: []string{kubeServiceObjKey.Name},
					Http: []*api_v1alpha3.HTTPRoute{
						{
							Match: []*api_v1alpha3.HTTPMatchRequest{baseMatchRequest},
							Route: defaultRoute,
						},
						{
							Match: []*api_v1alpha3.HTTPMatchRequest{baseMatchRequest},
							Route: defaultRoute,
						},
						{
							Match: []*api_v1alpha3.HTTPMatchRequest{baseMatchRequest},
							Route: defaultRoute,
						},
					},
				},
			}
			mockMeshClient.EXPECT().GetMesh(ctx, meshObjKey).Return(mesh, nil)
			mockDynamicClientGetter.EXPECT().GetClientForCluster(ctx, clusterName).Return(nil, nil)
			// computed DestinationRule
			computedDestinationRule := &client_v1alpha3.DestinationRule{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      meshService.Spec.GetKubeService().GetRef().GetName(),
					Namespace: meshService.Spec.GetKubeService().GetRef().GetNamespace(),
				},
				Spec: api_v1alpha3.DestinationRule{
					Host: kubeServiceObjKey.Name,
					TrafficPolicy: &api_v1alpha3.TrafficPolicy{
						Tls: &api_v1alpha3.TLSSettings{
							Mode: api_v1alpha3.TLSSettings_ISTIO_MUTUAL,
						},
					},
				},
			}
			mockDestinationRuleClient.EXPECT().CreateDestinationRule(ctx, computedDestinationRule).Return(nil)
			return &testContext{
				clusterName:            clusterName,
				meshObjKey:             meshObjKey,
				meshServiceObjKey:      meshServiceObjKey,
				kubeServiceObjKey:      kubeServiceObjKey,
				mesh:                   mesh,
				meshService:            meshService,
				trafficPolicy:          trafficPolicy,
				computedVirtualService: computedVirtualService,
				baseMatchRequest:       baseMatchRequest,
				defaultRoute:           defaultRoute,
			}
		}

		It("should upsert VirtualService", func() {
			testContext := setupTestContext()
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				testContext.meshService,
				testContext.mesh,
				testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should error if no destination is specified, and multiple ports are available on service", func() {
			testContext := setupTestContext()
			testContext.meshService.Spec.KubeService.Ports =
				append(testContext.meshService.Spec.KubeService.Ports, &zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
					Port: 8080,
					Name: "will fail",
				})
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				testContext.meshService,
				testContext.mesh,
				testContext.trafficPolicy)
			Expect(translatorError.ErrorMessage).
				To(ContainSubstring(istio_translator.NoSpecifiedPortError(testContext.meshService).Error()))
		})

		It("should translate Retries", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.Retries = &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
				Attempts:      5,
				PerTryTimeout: &types.Duration{Seconds: 2},
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.Retries = &api_v1alpha3.HTTPRetry{
					Attempts:      5,
					PerTryTimeout: &types.Duration{Seconds: 2},
				}
			}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate CorsPolicy", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.CorsPolicy = &zephyr_networking_types.TrafficPolicySpec_CorsPolicy{
				AllowOrigins: []*zephyr_networking_types.TrafficPolicySpec_StringMatch{
					{MatchType: &zephyr_networking_types.TrafficPolicySpec_StringMatch_Exact{Exact: "exact"}},
					{MatchType: &zephyr_networking_types.TrafficPolicySpec_StringMatch_Prefix{Prefix: "prefix"}},
					{MatchType: &zephyr_networking_types.TrafficPolicySpec_StringMatch_Regex{Regex: "regex"}},
				},
				AllowMethods:     []string{"GET", "POST"},
				AllowHeaders:     []string{"Header1", "Header2"},
				ExposeHeaders:    []string{"ExposedHeader1", "ExposedHeader2"},
				MaxAge:           &types.Duration{Seconds: 1},
				AllowCredentials: &types.BoolValue{Value: false},
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.CorsPolicy = &api_v1alpha3.CorsPolicy{
					AllowOrigins: []*api_v1alpha3.StringMatch{
						{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "exact"}},
						{MatchType: &api_v1alpha3.StringMatch_Prefix{Prefix: "prefix"}},
						{MatchType: &api_v1alpha3.StringMatch_Regex{Regex: "regex"}},
					},
					AllowMethods:     []string{"GET", "POST"},
					AllowHeaders:     []string{"Header1", "Header2"},
					ExposeHeaders:    []string{"ExposedHeader1", "ExposedHeader2"},
					MaxAge:           &types.Duration{Seconds: 1},
					AllowCredentials: &types.BoolValue{Value: false},
				}
			}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HeaderManipulation", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HeaderManipulation = &zephyr_networking_types.TrafficPolicySpec_HeaderManipulation{
				AppendRequestHeaders:  map[string]string{"a": "b"},
				RemoveRequestHeaders:  []string{"3", "4"},
				AppendResponseHeaders: map[string]string{"foo": "bar"},
				RemoveResponseHeaders: []string{"1", "2"},
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.Headers = &api_v1alpha3.Headers{
					Request: &api_v1alpha3.Headers_HeaderOperations{
						Add:    map[string]string{"a": "b"},
						Remove: []string{"3", "4"},
					},
					Response: &api_v1alpha3.Headers_HeaderOperations{
						Add:    map[string]string{"foo": "bar"},
						Remove: []string{"1", "2"},
					},
				}
			}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate Mirror destination on same cluster", func() {
			testContext := setupTestContext()
			destName := "name"
			destNamespace := "namespace"
			port := uint32(9080)
			destCluster := testContext.clusterName
			testContext.trafficPolicy[0].Spec.Mirror = &zephyr_networking_types.TrafficPolicySpec_Mirror{
				Destination: &zephyr_core_types.ResourceRef{
					Name:      destName,
					Namespace: destNamespace,
					Cluster:   destCluster,
				},
				Percentage: 50,
				Port:       port,
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.Mirror = &api_v1alpha3.Destination{
					Host: destName,
					Port: &api_v1alpha3.PortSelector{
						Number: port,
					},
				}
				httpRoute.MirrorPercentage = &api_v1alpha3.Percent{Value: 50.0}
			}
			backingMeshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
						},
					},
				},
			}
			mockResourceSelector.
				EXPECT().
				GetMeshServiceByRefSelector(ctx, destName, destNamespace, destCluster).
				Return(backingMeshService, nil)
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate Mirror destination on remote cluster", func() {
			testContext := setupTestContext()
			multiClusterDnsName := "multicluster-dns-name"
			destName := "name"
			destNamespace := "namespace"
			remoteClusterName := "remote-cluster"
			testContext.trafficPolicy[0].Spec.Mirror = &zephyr_networking_types.TrafficPolicySpec_Mirror{
				Destination: &zephyr_core_types.ResourceRef{
					Name:      destName,
					Namespace: destNamespace,
					Cluster:   remoteClusterName,
				},
				Percentage: 50,
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.Mirror = &api_v1alpha3.Destination{
					Host: multiClusterDnsName,
				}
				httpRoute.MirrorPercentage = &api_v1alpha3.Percent{Value: 50.0}
			}
			backingMeshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
						},
					},
					Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{MulticlusterDnsName: multiClusterDnsName},
				},
			}
			mockResourceSelector.
				EXPECT().
				GetMeshServiceByRefSelector(ctx, destName, destNamespace, remoteClusterName).
				Return(backingMeshService, nil)
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate FaultInjection of type Abort", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.FaultInjection = &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
				FaultInjectionType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Abort_{
					Abort: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Abort{
						ErrorType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Abort_HttpStatus{HttpStatus: 404},
					},
				},
				Percentage: 50,
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.Fault = &api_v1alpha3.HTTPFaultInjection{
					Abort: &api_v1alpha3.HTTPFaultInjection_Abort{
						ErrorType:  &api_v1alpha3.HTTPFaultInjection_Abort_HttpStatus{HttpStatus: 404},
						Percentage: &api_v1alpha3.Percent{Value: 50},
					},
				}
			}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate FaultInjection of type Delay of type Fixed", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.FaultInjection = &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
				FaultInjectionType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_{
					Delay: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay{
						HttpDelayType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_FixedDelay{
							FixedDelay: &types.Duration{Seconds: 2},
						},
					},
				},
				Percentage: 50,
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.Fault = &api_v1alpha3.HTTPFaultInjection{
					Delay: &api_v1alpha3.HTTPFaultInjection_Delay{
						HttpDelayType: &api_v1alpha3.HTTPFaultInjection_Delay_FixedDelay{FixedDelay: &types.Duration{Seconds: 2}},
						Percentage:    &api_v1alpha3.Percent{Value: 50},
					},
				}
			}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate FaultInjection of type Delay of type Exponential", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.FaultInjection = &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
				FaultInjectionType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_{
					Delay: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay{
						HttpDelayType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_ExponentialDelay{
							ExponentialDelay: &types.Duration{Seconds: 2},
						},
					},
				},
				Percentage: 50,
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.Fault = &api_v1alpha3.HTTPFaultInjection{
					Delay: &api_v1alpha3.HTTPFaultInjection_Delay{
						HttpDelayType: &api_v1alpha3.HTTPFaultInjection_Delay_ExponentialDelay{ExponentialDelay: &types.Duration{Seconds: 2}},
						Percentage:    &api_v1alpha3.Percent{Value: 50},
					},
				}
			}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate Retries", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.Retries = &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
				Attempts:      5,
				PerTryTimeout: &types.Duration{Seconds: 2},
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.Retries = &api_v1alpha3.HTTPRetry{
					Attempts:      5,
					PerTryTimeout: &types.Duration{Seconds: 2},
				}
			}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HeaderMatchers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers[0] = &zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
				Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_GET},
				Headers: []*zephyr_networking_types.TrafficPolicySpec_HeaderMatcher{
					{
						Name:        "name1",
						Value:       "value1",
						Regex:       false,
						InvertMatch: false,
					},
					{
						Name:        "name2",
						Value:       "*",
						Regex:       true,
						InvertMatch: false,
					},
					{
						Name:        "name3",
						Value:       "[a-z]+",
						Regex:       true,
						InvertMatch: true,
					},
				},
			}
			expectedMatchRequest := *testContext.baseMatchRequest
			expectedMatchRequest.Method = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: zephyr_core_types.HttpMethodValue_GET.String()}}
			expectedMatchRequest.Headers = map[string]*api_v1alpha3.StringMatch{
				"name1": {MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "value1"}},
				"name2": {MatchType: &api_v1alpha3.StringMatch_Regex{Regex: "*"}},
			}
			expectedMatchRequest.WithoutHeaders = map[string]*api_v1alpha3.StringMatch{
				"name3": {MatchType: &api_v1alpha3.StringMatch_Regex{Regex: "[a-z]+"}},
			}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1alpha3.HTTPMatchRequest{&expectedMatchRequest}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HttpMatcher exact path specifiers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers[0] = &zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
				Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_GET},
				PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Regex{
					Regex: "*",
				},
			}
			expectedMatchRequest := *testContext.baseMatchRequest
			expectedMatchRequest.Method = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: zephyr_core_types.HttpMethodValue_GET.String()}}
			expectedMatchRequest.Uri = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Regex{Regex: "*"}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1alpha3.HTTPMatchRequest{&expectedMatchRequest}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HttpMatcher prefix path specifiers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers[0] = &zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
				Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_GET},
				PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
					Prefix: "prefix",
				},
			}
			expectedMatchRequest := *testContext.baseMatchRequest
			expectedMatchRequest.Method = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: zephyr_core_types.HttpMethodValue_GET.String()}}
			expectedMatchRequest.Uri = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Prefix{Prefix: "prefix"}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1alpha3.HTTPMatchRequest{&expectedMatchRequest}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate QueryParamMatchers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers[0] = &zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
				Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_GET},
				QueryParameters: []*zephyr_networking_types.TrafficPolicySpec_QueryParameterMatcher{
					{
						Name:  "qp1",
						Value: "qpv1",
						Regex: false,
					},
					{
						Name:  "qp2",
						Value: "qpv2",
						Regex: true,
					},
				},
			}
			expectedMatchRequest := *testContext.baseMatchRequest
			expectedMatchRequest.Method = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: zephyr_core_types.HttpMethodValue_GET.String()}}
			expectedMatchRequest.QueryParams = map[string]*api_v1alpha3.StringMatch{
				"qp1": {
					MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "qpv1"},
				},
				"qp2": {
					MatchType: &api_v1alpha3.StringMatch_Regex{Regex: "qpv2"},
				},
			}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1alpha3.HTTPMatchRequest{&expectedMatchRequest}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HttpMatcher regex path specifiers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers[0] = &zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
				Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_GET},
				PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Regex{
					Regex: "*",
				},
			}
			expectedMatchRequest := *testContext.baseMatchRequest
			expectedMatchRequest.Method = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: zephyr_core_types.HttpMethodValue_GET.String()}}
			expectedMatchRequest.Uri = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Regex{Regex: "*"}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1alpha3.HTTPMatchRequest{&expectedMatchRequest}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HttpMatcher prefix path specifiers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers[0] = &zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
				Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_GET},
				PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
					Prefix: "prefix",
				},
			}
			expectedMatchRequest := *testContext.baseMatchRequest
			expectedMatchRequest.Method = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: zephyr_core_types.HttpMethodValue_GET.String()}}
			expectedMatchRequest.Uri = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Prefix{Prefix: "prefix"}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1alpha3.HTTPMatchRequest{&expectedMatchRequest}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HttpMatcher exact path specifiers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers[0] = &zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
				Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_GET},
				PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
					Exact: "path",
				},
			}
			expectedMatchRequest := *testContext.baseMatchRequest
			expectedMatchRequest.Method = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: zephyr_core_types.HttpMethodValue_GET.String()}}
			expectedMatchRequest.Uri = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "path"}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1alpha3.HTTPMatchRequest{&expectedMatchRequest}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate TrafficShift without subsets", func() {
			testContext := setupTestContext()
			destName := "name"
			destNamespace := "namespace"
			multiClusterDnsName := "multicluster-dns-name"
			destCluster := "remote-cluster-1"
			testContext.trafficPolicy[0].Spec.TrafficShift = &zephyr_networking_types.TrafficPolicySpec_MultiDestination{
				Destinations: []*zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
					{
						Destination: &zephyr_core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
							Cluster:   destCluster,
						},
						Weight: 50,
					},
				},
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.Route = []*api_v1alpha3.HTTPRouteDestination{
					{
						Destination: &api_v1alpha3.Destination{
							Host: multiClusterDnsName,
						},
						Weight: 50,
					},
				}
			}
			backingMeshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
						},
					},
					Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{MulticlusterDnsName: multiClusterDnsName},
				},
			}
			mockResourceSelector.
				EXPECT().
				GetMeshServiceByRefSelector(ctx, destName, destNamespace, destCluster).
				Return(backingMeshService, nil)
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate TrafficShift with ports", func() {
			testContext := setupTestContext()
			destName := "name"
			destNamespace := "namespace"
			multiClusterDnsName := "multicluster-dns-name"
			port := uint32(9080)
			destCluster := "remote-cluster-1"
			testContext.trafficPolicy[0].Spec.TrafficShift = &zephyr_networking_types.TrafficPolicySpec_MultiDestination{
				Destinations: []*zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
					{
						Destination: &zephyr_core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
							Cluster:   destCluster,
						},
						Weight: 50,
						Port:   port,
					},
				},
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.Route = []*api_v1alpha3.HTTPRouteDestination{
					{
						Destination: &api_v1alpha3.Destination{
							Host: multiClusterDnsName,
							Port: &api_v1alpha3.PortSelector{
								Number: port,
							},
						},
						Weight: 50,
					},
				}
			}
			backingMeshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
						},
					},
					Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{MulticlusterDnsName: multiClusterDnsName},
				},
			}
			mockResourceSelector.
				EXPECT().
				GetMeshServiceByRefSelector(ctx, destName, destNamespace, destCluster).
				Return(backingMeshService, nil)
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate TrafficShift with subsets", func() {
			testContext := setupTestContext()
			destName := "name"
			destNamespace := "namespace"
			declaredSubset := map[string]string{"env": "dev", "version": "v1"}
			expectedSubsetName := "env-dev_version-v1"
			destination := &zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
				Destination: &zephyr_core_types.ResourceRef{
					Name:      destName,
					Namespace: destNamespace,
					Cluster:   testContext.clusterName,
				},
				Subset: declaredSubset,
				Weight: 50,
			}
			testContext.trafficPolicy[0].Spec.TrafficShift = &zephyr_networking_types.TrafficPolicySpec_MultiDestination{
				Destinations: []*zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
					destination,
				},
			}
			for _, httpRoute := range testContext.computedVirtualService.Spec.Http {
				httpRoute.Route = []*api_v1alpha3.HTTPRouteDestination{
					{
						Destination: &api_v1alpha3.Destination{
							Host:   destName,
							Subset: expectedSubsetName,
						},
						Weight: 50,
					},
				}
			}
			backingMeshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
						},
					},
				},
			}
			existingDestRule := &client_v1alpha3.DestinationRule{}
			computedDestRule := &client_v1alpha3.DestinationRule{
				Spec: api_v1alpha3.DestinationRule{
					Subsets: []*api_v1alpha3.Subset{
						{
							Name:   expectedSubsetName,
							Labels: declaredSubset,
						},
					},
				},
			}
			mockResourceSelector.
				EXPECT().
				GetMeshServiceByRefSelector(ctx, destName, destNamespace, testContext.clusterName).
				Return(backingMeshService, nil)

			mockDynamicClientGetter.
				EXPECT().
				GetClientForCluster(ctx, testContext.clusterName).
				Return(nil, nil)
			mockDestinationRuleClient.
				EXPECT().
				GetDestinationRule(ctx, client.ObjectKey{Name: destName, Namespace: destNamespace}).
				Return(existingDestRule, nil)
			mockDestinationRuleClient.
				EXPECT().
				UpdateDestinationRule(ctx, computedDestRule).
				Return(nil)
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should error translating multi cluster TrafficShift with subsets", func() {
			testContext := setupTestContext()
			destName := "name"
			destNamespace := "namespace"
			multiClusterDnsName := "multicluster-dns-name"
			destCluster := "remote-cluster-1"
			declaredSubset := map[string]string{"env": "dev", "version": "v1"}
			destination := &zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
				Destination: &zephyr_core_types.ResourceRef{
					Name:      destName,
					Namespace: destNamespace,
					Cluster:   destCluster,
				},
				Subset: declaredSubset,
				Weight: 50,
			}
			testContext.trafficPolicy[0].Spec.TrafficShift = &zephyr_networking_types.TrafficPolicySpec_MultiDestination{
				Destinations: []*zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
					destination,
				},
			}
			backingMeshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
						},
					},
					Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{MulticlusterDnsName: multiClusterDnsName},
				},
			}
			mockResourceSelector.
				EXPECT().
				GetMeshServiceByRefSelector(ctx, destName, destNamespace, destCluster).
				Return(backingMeshService, nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).NotTo(BeNil())
			Expect(translatorError.ErrorMessage).
				To(ContainSubstring(istio_translator.MultiClusterSubsetsNotSupported(destination).Error()))
		})

		It("should return error if multiple MeshServices found for name/namespace/cluster", func() {
			testContext := setupTestContext()
			destName := "name"
			destNamespace := "namespace"
			remoteClusterName := "remote-cluster"
			testContext.trafficPolicy[0].Spec.Mirror = &zephyr_networking_types.TrafficPolicySpec_Mirror{
				Destination: &zephyr_core_types.ResourceRef{
					Name:      destName,
					Namespace: destNamespace,
					Cluster:   remoteClusterName,
				},
				Percentage: 50,
			}
			err := eris.New("mesh-service-selector-error")
			mockResourceSelector.
				EXPECT().
				GetMeshServiceByRefSelector(ctx, destName, destNamespace, remoteClusterName).
				Return(nil, err)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError.ErrorMessage).To(ContainSubstring(err.Error()))
		})

		It("should translate HTTP RequestMatchers and order the resulting HTTPRoutes", func() {
			testContext := setupTestContext()
			labels := map[string]string{"env": "dev"}
			namespaces := []string{"n1", "n2"}
			testContext.trafficPolicy[0].Spec.SourceSelector = &zephyr_core_types.WorkloadSelector{
				Labels:     labels,
				Namespaces: namespaces,
			}
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers = []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
						Exact: "path",
					},
					Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_GET},
				},
				{
					Headers: []*zephyr_networking_types.TrafficPolicySpec_HeaderMatcher{
						{
							Name:        "name3",
							Value:       "[a-z]+",
							Regex:       true,
							InvertMatch: true,
						},
					},
					Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_POST},
				},
			}
			testContext.computedVirtualService.Spec.Http = []*api_v1alpha3.HTTPRoute{
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							Method:          &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "GET"}},
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "path"}},
							SourceLabels:    labels,
							SourceNamespace: namespaces[1],
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							Method: &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "POST"}},
							WithoutHeaders: map[string]*api_v1alpha3.StringMatch{
								"name3": {MatchType: &api_v1alpha3.StringMatch_Regex{Regex: "[a-z]+"}},
							},
							SourceLabels:    labels,
							SourceNamespace: namespaces[1],
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							Method:          &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "GET"}},
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "path"}},
							SourceLabels:    labels,
							SourceNamespace: namespaces[0],
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							Method: &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "POST"}},
							WithoutHeaders: map[string]*api_v1alpha3.StringMatch{
								"name3": {MatchType: &api_v1alpha3.StringMatch_Regex{Regex: "[a-z]+"}},
							},
							SourceLabels:    labels,
							SourceNamespace: namespaces[0],
						},
					},
					Route: testContext.defaultRoute,
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should deterministically order HTTPRoutes according to decreasing specificity", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers = []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
						Exact: "exact-path",
					},
				},
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
						Prefix: "/prefix",
					},
					Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{
						Method: zephyr_core_types.HttpMethodValue_GET,
					},
				},
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
						Exact: "exact-path",
					},
					Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{
						Method: zephyr_core_types.HttpMethodValue_GET,
					},
				},
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
						Exact: "exact-path",
					},
					Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{
						Method: zephyr_core_types.HttpMethodValue_PUT,
					},
				},
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Regex{
						Regex: "www*",
					},
				},
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
						Prefix: "/",
					},
					Headers: []*zephyr_networking_types.TrafficPolicySpec_HeaderMatcher{
						{
							Name:        "set-cookie",
							Value:       "foo=bar",
							InvertMatch: true,
						},
					},
				},
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
						Prefix: "/",
					},
					Headers: []*zephyr_networking_types.TrafficPolicySpec_HeaderMatcher{
						{
							Name:        "content-type",
							Value:       "text/html",
							InvertMatch: false,
						},
					},
				},
			}
			testContext.computedVirtualService.Spec.Http = []*api_v1alpha3.HTTPRoute{
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Headers: map[string]*api_v1alpha3.StringMatch{
								"content-type": {MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "text/html"}},
							},
							Uri: &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Prefix{Prefix: "/"}},
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "exact-path"}},
							Method:          &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "PUT"}},
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "exact-path"}},
							Method:          &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "GET"}},
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "exact-path"}},
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Regex{Regex: "www*"}},
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Prefix{Prefix: "/prefix"}},
							Method:          &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "GET"}},
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							WithoutHeaders: map[string]*api_v1alpha3.StringMatch{
								"set-cookie": {MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "foo=bar"}},
							},
							Uri: &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Prefix{Prefix: "/"}},
						},
					},
					Route: testContext.defaultRoute,
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should order longer prefixes, regexes, and exact URI matchers before shorter ones", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers = []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
						Exact: "short",
					},
				},
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
						Exact: "longer",
					},
				},
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
						Prefix: "/short",
					},
				},
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
						Prefix: "/longer",
					},
				},
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Regex{
						Regex: "short*",
					},
				},
				{
					PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Regex{
						Regex: "longer*",
					},
				},
			}
			testContext.computedVirtualService.Spec.Http = []*api_v1alpha3.HTTPRoute{
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "longer"}},
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: "short"}},
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Regex{Regex: "longer*"}},
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Regex{Regex: "short*"}},
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Prefix{Prefix: "/longer"}},
						},
					},
					Route: testContext.defaultRoute,
				},
				{
					Match: []*api_v1alpha3.HTTPMatchRequest{
						{
							SourceNamespace: testContext.baseMatchRequest.GetSourceNamespace(),
							Uri:             &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Prefix{Prefix: "/short"}},
						},
					},
					Route: testContext.defaultRoute,
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				UpsertVirtualServiceSpec(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})
	})
})
