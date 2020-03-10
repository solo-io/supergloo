package istio_translator_test

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	istio_networking "github.com/solo-io/mesh-projects/pkg/clients/istio/networking"
	mock_istio_networking "github.com/solo-io/mesh-projects/pkg/clients/istio/networking/mocks"
	mock_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	"github.com/solo-io/mesh-projects/services/common"
	mock_mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager/mocks"
	istio_translator "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/istio-translator"
	mock_preprocess "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess/mocks"
	api_v1beta1 "istio.io/api/networking/v1beta1"
	client_v1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type testContext struct {
	clusterName            string
	meshObjKey             client.ObjectKey
	meshServiceObjKey      client.ObjectKey
	kubeServiceObjKey      client.ObjectKey
	mesh                   *discovery_v1alpha1.Mesh
	meshService            *discovery_v1alpha1.MeshService
	trafficPolicy          []*v1alpha1.TrafficPolicy
	computedVirtualService *client_v1beta1.VirtualService
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
		mockMeshServiceSelector      *mock_preprocess.MockMeshServiceSelector
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockDynamicClientGetter = mock_mc_manager.NewMockDynamicClientGetter(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockMeshServiceClient = mock_core.NewMockMeshServiceClient(ctrl)
		mockVirtualServiceClient = mock_istio_networking.NewMockVirtualServiceClient(ctrl)
		mockMeshServiceSelector = mock_preprocess.NewMockMeshServiceSelector(ctrl)
		istioTrafficPolicyTranslator = istio_translator.NewIstioTrafficPolicyTranslator(
			mockDynamicClientGetter,
			mockMeshClient,
			mockMeshServiceClient,
			mockMeshServiceSelector,
			func(client client.Client) istio_networking.VirtualServiceClient {
				return mockVirtualServiceClient
			},
		)
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Context("should translate TrafficPolicies into VirtualService and upsert", func() {
		setupTestContext := func() *testContext {
			clusterName := "clusterName"
			meshObjKey := client.ObjectKey{Name: "mesh-name", Namespace: "mesh-namespace"}
			meshServiceObjKey := client.ObjectKey{Name: "mesh-service-name", Namespace: "mesh-service-namespace"}
			kubeServiceObjKey := client.ObjectKey{Name: "kube-service-name", Namespace: "kube-service-namespace"}
			meshServiceFederationMCDnsName := "multiclusterDNSname"
			meshService := &discovery_v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{
					Name:        meshServiceObjKey.Name,
					Namespace:   meshServiceObjKey.Namespace,
					ClusterName: clusterName,
				},
				Spec: discovery_types.MeshServiceSpec{
					Mesh: &core_types.ResourceRef{
						Name:      meshObjKey.Name,
						Namespace: meshObjKey.Namespace,
					},
					KubeService: &discovery_types.KubeService{
						Ref: &core_types.ResourceRef{
							Name:      kubeServiceObjKey.Name,
							Namespace: kubeServiceObjKey.Namespace,
							Cluster:   &types.StringValue{Value: clusterName},
						},
					},
					Federation: &discovery_types.Federation{
						MulticlusterDnsName: meshServiceFederationMCDnsName,
					},
				},
			}
			mesh := &discovery_v1alpha1.Mesh{
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &discovery_types.MeshSpec_Istio{
						Istio: &discovery_types.IstioMesh{},
					},
				},
			}
			trafficPolicy := []*v1alpha1.TrafficPolicy{{
				Spec: networking_types.TrafficPolicySpec{
					HttpRequestMatchers: []*networking_types.HttpMatcher{},
				}},
			}
			computedVirtualService := &client_v1beta1.VirtualService{
				ObjectMeta: v1.ObjectMeta{
					Name:      meshService.Spec.GetKubeService().GetRef().GetName(),
					Namespace: meshService.Spec.GetKubeService().GetRef().GetNamespace(),
				},
				Spec: api_v1beta1.VirtualService{
					Hosts: []string{meshServiceObjKey.Name + "." + meshServiceObjKey.Namespace},
					Http: []*api_v1beta1.HTTPRoute{
						{
							Match: []*api_v1beta1.HTTPMatchRequest{},
						},
					},
				},
			}
			mockMeshClient.EXPECT().Get(ctx, meshObjKey).Return(mesh, nil)
			mockDynamicClientGetter.EXPECT().GetClientForCluster(clusterName).Return(nil, true)
			return &testContext{
				clusterName:            clusterName,
				meshObjKey:             meshObjKey,
				meshServiceObjKey:      meshServiceObjKey,
				kubeServiceObjKey:      kubeServiceObjKey,
				mesh:                   mesh,
				meshService:            meshService,
				trafficPolicy:          trafficPolicy,
				computedVirtualService: computedVirtualService,
			}
		}

		It("should upsert VirtualService", func() {
			testContext := setupTestContext()
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				testContext.meshService,
				testContext.mesh,
				testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate CorsPolicy", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.CorsPolicy = &networking_types.CorsPolicy{
				AllowOrigins: []*networking_types.StringMatch{
					{MatchType: &networking_types.StringMatch_Exact{Exact: "exact"}},
					{MatchType: &networking_types.StringMatch_Prefix{Prefix: "prefix"}},
					{MatchType: &networking_types.StringMatch_Regex{Regex: "regex"}},
				},
				AllowMethods:     []string{"GET", "POST"},
				AllowHeaders:     []string{"Header1", "Header2"},
				ExposeHeaders:    []string{"ExposedHeader1", "ExposedHeader2"},
				MaxAge:           &types.Duration{Seconds: 1},
				AllowCredentials: &types.BoolValue{Value: false},
			}
			testContext.computedVirtualService.Spec.Http[0].CorsPolicy = &api_v1beta1.CorsPolicy{
				AllowOrigins: []*api_v1beta1.StringMatch{
					{MatchType: &api_v1beta1.StringMatch_Exact{Exact: "exact"}},
					{MatchType: &api_v1beta1.StringMatch_Prefix{Prefix: "prefix"}},
					{MatchType: &api_v1beta1.StringMatch_Regex{Regex: "regex"}},
				},
				AllowMethods:     []string{"GET", "POST"},
				AllowHeaders:     []string{"Header1", "Header2"},
				ExposeHeaders:    []string{"ExposedHeader1", "ExposedHeader2"},
				MaxAge:           &types.Duration{Seconds: 1},
				AllowCredentials: &types.BoolValue{Value: false},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HeaderManipulation", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HeaderManipulation = &networking_types.HeaderManipulation{
				AppendRequestHeaders:  map[string]string{"a": "b"},
				RemoveRequestHeaders:  []string{"3", "4"},
				AppendResponseHeaders: map[string]string{"foo": "bar"},
				RemoveResponseHeaders: []string{"1", "2"},
			}
			testContext.computedVirtualService.Spec.Http[0].Headers = &api_v1beta1.Headers{
				Request: &api_v1beta1.Headers_HeaderOperations{
					Add:    map[string]string{"a": "b"},
					Remove: []string{"3", "4"},
				},
				Response: &api_v1beta1.Headers_HeaderOperations{
					Add:    map[string]string{"foo": "bar"},
					Remove: []string{"1", "2"},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate Mirror destination on same cluster", func() {
			testContext := setupTestContext()
			destName := "name"
			destNamespace := "namespace"
			destCluster := &types.StringValue{Value: testContext.clusterName}
			testContext.trafficPolicy[0].Spec.Mirror = &networking_types.Mirror{
				Destination: &core_types.ResourceRef{
					Name:      destName,
					Namespace: destNamespace,
					Cluster:   destCluster,
				},
				Percentage: 50,
			}
			testContext.computedVirtualService.Spec.Http[0].Mirror = &api_v1beta1.Destination{
				Host: destName + "." + destNamespace,
			}
			testContext.computedVirtualService.Spec.Http[0].MirrorPercentage = &api_v1beta1.Percent{Value: 50.0}
			backingMeshService := &discovery_v1alpha1.MeshService{
				Spec: discovery_types.MeshServiceSpec{
					KubeService: &discovery_types.KubeService{
						Ref: &core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
						},
					},
				},
			}
			mockMeshServiceSelector.
				EXPECT().
				GetBackingMeshService(ctx, destName, destNamespace, destCluster.GetValue()).
				Return(backingMeshService, nil)
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate Mirror destination on same *local* cluster", func() {
			testContext := setupTestContext()
			destName := "name"
			destNamespace := "namespace"
			testContext.meshService.Spec.GetKubeService().GetRef().GetCluster().Value = common.LocalClusterName
			testContext.trafficPolicy[0].Spec.Mirror = &networking_types.Mirror{
				Destination: &core_types.ResourceRef{
					Name:      destName,
					Namespace: destNamespace,
					// omit cluster to specify local cluster
				},
				Percentage: 50,
			}
			testContext.computedVirtualService.Spec.Http[0].Mirror = &api_v1beta1.Destination{
				Host: destName + "." + destNamespace,
			}
			testContext.computedVirtualService.Spec.Http[0].MirrorPercentage = &api_v1beta1.Percent{Value: 50.0}
			backingMeshService := &discovery_v1alpha1.MeshService{
				Spec: discovery_types.MeshServiceSpec{
					KubeService: &discovery_types.KubeService{
						Ref: &core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
						},
					},
				},
			}
			mockMeshServiceSelector.
				EXPECT().
				GetBackingMeshService(ctx, destName, destNamespace, "").
				Return(backingMeshService, nil)
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
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
			testContext.trafficPolicy[0].Spec.Mirror = &networking_types.Mirror{
				Destination: &core_types.ResourceRef{
					Name:      destName,
					Namespace: destNamespace,
					Cluster:   &types.StringValue{Value: remoteClusterName},
				},
				Percentage: 50,
			}
			testContext.computedVirtualService.Spec.Http[0].Mirror = &api_v1beta1.Destination{
				Host: multiClusterDnsName,
			}
			testContext.computedVirtualService.Spec.Http[0].MirrorPercentage = &api_v1beta1.Percent{Value: 50.0}
			backingMeshService := &discovery_v1alpha1.MeshService{
				Spec: discovery_types.MeshServiceSpec{
					KubeService: &discovery_types.KubeService{
						Ref: &core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
						},
					},
					Federation: &discovery_types.Federation{MulticlusterDnsName: multiClusterDnsName},
				},
			}
			mockMeshServiceSelector.
				EXPECT().
				GetBackingMeshService(ctx, destName, destNamespace, remoteClusterName).
				Return(backingMeshService, nil)
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate FaultInjection of type Abort", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.FaultInjection = &networking_types.FaultInjection{
				FaultInjectionType: &networking_types.FaultInjection_Abort_{
					Abort: &networking_types.FaultInjection_Abort{
						ErrorType: &networking_types.FaultInjection_Abort_HttpStatus{HttpStatus: 404},
					},
				},
				Percentage: 50,
			}
			testContext.computedVirtualService.Spec.Http[0].Fault = &api_v1beta1.HTTPFaultInjection{
				Abort: &api_v1beta1.HTTPFaultInjection_Abort{
					ErrorType:  &api_v1beta1.HTTPFaultInjection_Abort_HttpStatus{HttpStatus: 404},
					Percentage: &api_v1beta1.Percent{Value: 50},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate FaultInjection of type Delay of type Fixed", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.FaultInjection = &networking_types.FaultInjection{
				FaultInjectionType: &networking_types.FaultInjection_Delay_{
					Delay: &networking_types.FaultInjection_Delay{
						HttpDelayType: &networking_types.FaultInjection_Delay_FixedDelay{
							FixedDelay: &types.Duration{Seconds: 2},
						},
					},
				},
				Percentage: 50,
			}
			testContext.computedVirtualService.Spec.Http[0].Fault = &api_v1beta1.HTTPFaultInjection{
				Delay: &api_v1beta1.HTTPFaultInjection_Delay{
					HttpDelayType: &api_v1beta1.HTTPFaultInjection_Delay_FixedDelay{FixedDelay: &types.Duration{Seconds: 2}},
					Percentage:    &api_v1beta1.Percent{Value: 50},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate FaultInjection of type Delay of type Exponential", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.FaultInjection = &networking_types.FaultInjection{
				FaultInjectionType: &networking_types.FaultInjection_Delay_{
					Delay: &networking_types.FaultInjection_Delay{
						HttpDelayType: &networking_types.FaultInjection_Delay_ExponentialDelay{
							ExponentialDelay: &types.Duration{Seconds: 2},
						},
					},
				},
				Percentage: 50,
			}
			testContext.computedVirtualService.Spec.Http[0].Fault = &api_v1beta1.HTTPFaultInjection{
				Delay: &api_v1beta1.HTTPFaultInjection_Delay{
					HttpDelayType: &api_v1beta1.HTTPFaultInjection_Delay_ExponentialDelay{ExponentialDelay: &types.Duration{Seconds: 2}},
					Percentage:    &api_v1beta1.Percent{Value: 50},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate Retries", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.Retries = &networking_types.RetryPolicy{
				Attempts:      5,
				PerTryTimeout: &types.Duration{Seconds: 2},
			}
			testContext.computedVirtualService.Spec.Http[0].Retries = &api_v1beta1.HTTPRetry{
				Attempts:      5,
				PerTryTimeout: &types.Duration{Seconds: 2},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HeaderMatchers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers = []*networking_types.HttpMatcher{{
				Method: networking_types.HttpMethod_GET,
				Headers: []*networking_types.HeaderMatcher{
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
			}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1beta1.HTTPMatchRequest{
				{
					Method: &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: networking_types.HttpMethod_GET.String()}},
					Headers: map[string]*api_v1beta1.StringMatch{
						"name1": {MatchType: &api_v1beta1.StringMatch_Exact{Exact: "value1"}},
						"name2": {MatchType: &api_v1beta1.StringMatch_Regex{Regex: "*"}},
					},
					WithoutHeaders: map[string]*api_v1beta1.StringMatch{
						"name3": {MatchType: &api_v1beta1.StringMatch_Regex{Regex: "[a-z]+"}},
					},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HttpMatcher exact path specifiers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers = []*networking_types.HttpMatcher{{
				Method: networking_types.HttpMethod_GET,
				PathSpecifier: &networking_types.HttpMatcher_Regex{
					Regex: "*",
				},
			}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1beta1.HTTPMatchRequest{
				{
					Method: &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: networking_types.HttpMethod_GET.String()}},
					Uri:    &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Regex{Regex: "*"}},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HttpMatcher prefix path specifiers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers = []*networking_types.HttpMatcher{{
				Method: networking_types.HttpMethod_GET,
				PathSpecifier: &networking_types.HttpMatcher_Prefix{
					Prefix: "prefix",
				},
			}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1beta1.HTTPMatchRequest{
				{
					Method: &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: networking_types.HttpMethod_GET.String()}},
					Uri:    &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Prefix{Prefix: "prefix"}},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate QueryParamMatchers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers = []*networking_types.HttpMatcher{{
				Method: networking_types.HttpMethod_GET,
				QueryParameters: []*networking_types.QueryParameterMatcher{
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
			}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1beta1.HTTPMatchRequest{
				{
					Method: &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: networking_types.HttpMethod_GET.String()}},
					QueryParams: map[string]*api_v1beta1.StringMatch{
						"qp1": {
							MatchType: &api_v1beta1.StringMatch_Exact{Exact: "qpv1"},
						},
						"qp2": {
							MatchType: &api_v1beta1.StringMatch_Regex{Regex: "qpv2"},
						},
					},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HttpMatcher regex path specifiers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers = []*networking_types.HttpMatcher{{
				Method: networking_types.HttpMethod_GET,
				PathSpecifier: &networking_types.HttpMatcher_Regex{
					Regex: "*",
				},
			}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1beta1.HTTPMatchRequest{
				{
					Method: &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: networking_types.HttpMethod_GET.String()}},
					Uri:    &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Regex{Regex: "*"}},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HttpMatcher prefix path specifiers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers = []*networking_types.HttpMatcher{{
				Method: networking_types.HttpMethod_GET,
				PathSpecifier: &networking_types.HttpMatcher_Prefix{
					Prefix: "prefix",
				},
			}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1beta1.HTTPMatchRequest{
				{
					Method: &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: networking_types.HttpMethod_GET.String()}},
					Uri:    &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Prefix{Prefix: "prefix"}},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HttpMatcher exact path specifiers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers = []*networking_types.HttpMatcher{{
				Method: networking_types.HttpMethod_GET,
				PathSpecifier: &networking_types.HttpMatcher_Exact{
					Exact: "path",
				},
			}}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1beta1.HTTPMatchRequest{
				{
					Method: &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: networking_types.HttpMethod_GET.String()}},
					Uri:    &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: "path"}},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})

		It("should translate HTTP RequestMatchers", func() {
			testContext := setupTestContext()
			testContext.trafficPolicy[0].Spec.HttpRequestMatchers = []*networking_types.HttpMatcher{
				{
					PathSpecifier: &networking_types.HttpMatcher_Exact{
						Exact: "path",
					},
					Method: networking_types.HttpMethod_GET,
				},
				{
					Headers: []*networking_types.HeaderMatcher{
						{
							Name:        "name3",
							Value:       "[a-z]+",
							Regex:       true,
							InvertMatch: true,
						},
					},
					Method: networking_types.HttpMethod_POST,
				},
			}
			testContext.computedVirtualService.Spec.Http[0].Match = []*api_v1beta1.HTTPMatchRequest{
				{
					Method: &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: "GET"}},
					Uri:    &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: "path"}},
				},
				{
					Method: &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: "POST"}},
					WithoutHeaders: map[string]*api_v1beta1.StringMatch{
						"name3": {MatchType: &api_v1beta1.StringMatch_Regex{Regex: "[a-z]+"}},
					},
				},
			}
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
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
			destCluster := &types.StringValue{Value: "remote-cluster-1"}
			testContext.trafficPolicy[0].Spec.TrafficShift = &networking_types.MultiDestination{
				Destinations: []*networking_types.MultiDestination_WeightedDestination{
					{
						Destination: &core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
							Cluster:   destCluster,
						},
						Weight: 50,
					},
				},
			}
			testContext.computedVirtualService.Spec.Http[0].Route = []*api_v1beta1.HTTPRouteDestination{
				{
					Destination: &api_v1beta1.Destination{
						Host: multiClusterDnsName,
					},
					Weight: 50,
				},
			}
			backingMeshService := &discovery_v1alpha1.MeshService{
				Spec: discovery_types.MeshServiceSpec{
					KubeService: &discovery_types.KubeService{
						Ref: &core_types.ResourceRef{
							Name:      destName,
							Namespace: destNamespace,
						},
					},
					Federation: &discovery_types.Federation{MulticlusterDnsName: multiClusterDnsName},
				},
			}
			mockMeshServiceSelector.
				EXPECT().
				GetBackingMeshService(ctx, destName, destNamespace, destCluster.GetValue()).
				Return(backingMeshService, nil)
			mockVirtualServiceClient.
				EXPECT().
				Upsert(ctx, testContext.computedVirtualService).
				Return(nil)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError).To(BeNil())
		})
	})

	Describe("should return translator errors", func() {
		setupTestContext := func() *testContext {
			clusterName := "clusterName"
			meshObjKey := client.ObjectKey{Name: "mesh-name", Namespace: "mesh-namespace"}
			meshServiceObjKey := client.ObjectKey{Name: "mesh-service-name", Namespace: "mesh-service-namespace"}
			kubeServiceObjKey := client.ObjectKey{Name: "kube-service-name", Namespace: "kube-service-namespace"}
			meshServiceFederationMCDnsName := "multiclusterDNSname"
			meshService := &discovery_v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{
					Name:        meshServiceObjKey.Name,
					Namespace:   meshServiceObjKey.Namespace,
					ClusterName: clusterName,
				},
				Spec: discovery_types.MeshServiceSpec{
					Mesh: &core_types.ResourceRef{
						Name:      meshObjKey.Name,
						Namespace: meshObjKey.Namespace,
					},
					KubeService: &discovery_types.KubeService{
						Ref: &core_types.ResourceRef{
							Name:      kubeServiceObjKey.Name,
							Namespace: kubeServiceObjKey.Namespace,
							Cluster:   &types.StringValue{Value: clusterName},
						},
					},
					Federation: &discovery_types.Federation{
						MulticlusterDnsName: meshServiceFederationMCDnsName,
					},
				},
			}
			mesh := &discovery_v1alpha1.Mesh{
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &discovery_types.MeshSpec_Istio{
						Istio: &discovery_types.IstioMesh{},
					},
				},
			}
			trafficPolicy := []*v1alpha1.TrafficPolicy{{
				Spec: networking_types.TrafficPolicySpec{
					HttpRequestMatchers: []*networking_types.HttpMatcher{},
				}},
			}
			return &testContext{
				clusterName:       clusterName,
				meshObjKey:        meshObjKey,
				meshServiceObjKey: meshServiceObjKey,
				kubeServiceObjKey: kubeServiceObjKey,
				mesh:              mesh,
				meshService:       meshService,
				trafficPolicy:     trafficPolicy,
			}
		}

		It("should return error if multiple MeshServices found for name/namespace/cluster", func() {
			testContext := setupTestContext()
			destName := "name"
			destNamespace := "namespace"
			remoteClusterName := "remote-cluster"
			testContext.trafficPolicy[0].Spec.Mirror = &networking_types.Mirror{
				Destination: &core_types.ResourceRef{
					Name:      destName,
					Namespace: destNamespace,
					Cluster:   &types.StringValue{Value: remoteClusterName},
				},
				Percentage: 50,
			}
			err := eris.New("mesh-service-selector-error")
			mockMeshServiceSelector.
				EXPECT().
				GetBackingMeshService(ctx, destName, destNamespace, remoteClusterName).
				Return(nil, err)
			translatorError := istioTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx, testContext.meshService, testContext.mesh, testContext.trafficPolicy)
			Expect(translatorError.ErrorMessage).To(ContainSubstring(err.Error()))
		})
	})
})
