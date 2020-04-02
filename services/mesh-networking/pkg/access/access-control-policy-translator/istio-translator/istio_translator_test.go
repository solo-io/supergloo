package istio_translator_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	istio_security "github.com/solo-io/mesh-projects/pkg/clients/istio/security"
	mock_istio_security "github.com/solo-io/mesh-projects/pkg/clients/istio/security/mock"
	mock_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	mock_mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager/mocks"
	access_control_policy_translator "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/access/access-control-policy-translator"
	istio_translator "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/access/access-control-policy-translator/istio-translator"
	security_v1beta1 "istio.io/api/security/v1beta1"
	"istio.io/api/type/v1beta1"
	client_security_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("IstioTranslator", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 context.Context
		authPolicyClient    *mock_istio_security.MockAuthorizationPolicyClient
		meshClient          *mock_core.MockMeshClient
		istioTranslator     istio_translator.IstioTranslator
		dynamicClientGetter *mock_mc_manager.MockDynamicClientGetter
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		authPolicyClient = mock_istio_security.NewMockAuthorizationPolicyClient(ctrl)
		meshClient = mock_core.NewMockMeshClient(ctrl)
		dynamicClientGetter = mock_mc_manager.NewMockDynamicClientGetter(ctrl)
		istioTranslator = istio_translator.NewIstioTranslator(
			meshClient,
			dynamicClientGetter,
			func(client client.Client) istio_security.AuthorizationPolicyClient {
				return authPolicyClient
			},
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	type testData struct {
		accessControlPolicy *networking_v1alpha1.AccessControlPolicy
		targetServices      []access_control_policy_translator.TargetService
		clusterNames        []string
		acpClusterNames     []string
		trustDomains        []string
		namespaces          []string
		allowedPaths        []string
		allowedMethods      []core_types.HttpMethodValue
		allowedPorts        []string
	}

	// convenience method for converting HttpMethod enum to string
	var methodsToString = func(methodEnums []core_types.HttpMethodValue) []string {
		methods := make([]string, 0, len(methodEnums))
		for _, methodEnum := range methodEnums {
			methods = append(methods, methodEnum.String())
		}
		return methods
	}

	var initTestData = func() testData {
		clusterNames := []string{"cluster-name1", "cluster-name2"}
		acpClusterNames := append(clusterNames[:0:0], clusterNames...)
		trustDomains := []string{"cluster.local1", "cluster.local2"}
		namespaces := []string{"source-namespace1", "source-namespace2", "source-namespace3"}
		allowedPaths := []string{"/path1", "/path2"}
		allowedMethods := []core_types.HttpMethodValue{core_types.HttpMethodValue_GET, core_types.HttpMethodValue_POST}
		allowedPortsInts := []uint32{8080, 9080}
		allowedPortsString := []string{"8080", "9080"}
		istioMesh1 := &discovery_v1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{
					Istio: &discovery_types.MeshSpec_IstioMesh{
						CitadelInfo: &discovery_types.MeshSpec_IstioMesh_CitadelInfo{TrustDomain: trustDomains[0]},
					},
				},
				Cluster: &core_types.ResourceRef{Name: clusterNames[0]},
			},
		}
		istioMesh2 := &discovery_v1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{
					Istio: &discovery_types.MeshSpec_IstioMesh{
						CitadelInfo: &discovery_types.MeshSpec_IstioMesh_CitadelInfo{TrustDomain: trustDomains[1]},
					},
				},
				Cluster: &core_types.ResourceRef{Name: clusterNames[1]},
			},
		}
		acp := &networking_v1alpha1.AccessControlPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name:      "acp-name",
				Namespace: "acp-namespace",
			},
			Spec: networking_types.AccessControlPolicySpec{
				SourceSelector: &core_types.IdentitySelector{
					IdentitySelectorType: &core_types.IdentitySelector_Matcher_{
						Matcher: &core_types.IdentitySelector_Matcher{
							Namespaces: namespaces,
							Clusters:   acpClusterNames,
						},
					},
				},
				AllowedPaths:   allowedPaths,
				AllowedMethods: allowedMethods,
				AllowedPorts:   allowedPortsInts,
			},
		}
		targetServices := []access_control_policy_translator.TargetService{
			{
				MeshService: &discovery_v1alpha1.MeshService{
					Spec: discovery_types.MeshServiceSpec{
						KubeService: &discovery_types.MeshServiceSpec_KubeService{
							WorkloadSelectorLabels: map[string]string{
								"k1a": "v1a", "k1b": "v1b",
							},
							Ref: &core_types.ResourceRef{Namespace: "namespace1"},
						},
					},
				},
				Mesh: istioMesh1,
			},
			{
				MeshService: &discovery_v1alpha1.MeshService{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "namespace2",
					},
					Spec: discovery_types.MeshServiceSpec{
						KubeService: &discovery_types.MeshServiceSpec_KubeService{
							WorkloadSelectorLabels: map[string]string{
								"k2a": "v2a", "k2b": "v2b",
							},
							Ref: &core_types.ResourceRef{Namespace: "namespace2"},
						},
					},
				},
				Mesh: istioMesh2,
			},
		}
		meshClient.
			EXPECT().
			List(ctx).
			Return(&discovery_v1alpha1.MeshList{
				Items: []discovery_v1alpha1.Mesh{*istioMesh1, *istioMesh2},
			}, nil)
		return testData{
			accessControlPolicy: acp,
			targetServices:      targetServices,
			clusterNames:        clusterNames,
			acpClusterNames:     acpClusterNames,
			trustDomains:        trustDomains,
			namespaces:          namespaces,
			allowedPaths:        allowedPaths,
			allowedMethods:      allowedMethods,
			allowedPorts:        allowedPortsString,
		}
	}

	It("should translate AccessControlPolicy to AuthorizationPolicies per target service", func() {
		testData := initTestData()
		var expectedPrincipals []string
		var expectedAuthPolicies []*client_security_v1beta1.AuthorizationPolicy
		for _, trustDomain := range testData.trustDomains {
			for _, namespace := range testData.namespaces {
				expectedPrincipals = append(
					expectedPrincipals,
					fmt.Sprintf("%s/ns/%s/sa/*", trustDomain, namespace),
				)
			}
		}
		for _, targetService := range testData.targetServices {
			expectedAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
				ObjectMeta: v1.ObjectMeta{
					Name:      testData.accessControlPolicy.GetName() + "-" + targetService.MeshService.Name,
					Namespace: targetService.MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
				},
				Spec: security_v1beta1.AuthorizationPolicy{
					Selector: &v1beta1.WorkloadSelector{
						MatchLabels: targetService.MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
					},
					Rules: []*security_v1beta1.Rule{
						{
							From: []*security_v1beta1.Rule_From{
								{
									Source: &security_v1beta1.Source{
										Principals: expectedPrincipals,
									},
								},
							},
							To: []*security_v1beta1.Rule_To{
								{
									Operation: &security_v1beta1.Operation{
										Ports:   testData.allowedPorts,
										Methods: methodsToString(testData.accessControlPolicy.Spec.GetAllowedMethods()),
										Paths:   testData.accessControlPolicy.Spec.GetAllowedPaths(),
									},
								},
							},
						},
					},
					Action: security_v1beta1.AuthorizationPolicy_ALLOW,
				},
			}
			expectedAuthPolicies = append(expectedAuthPolicies, expectedAuthPolicy)
		}
		for i, expectedAuthPolicy := range expectedAuthPolicies {
			dynamicClientGetter.EXPECT().GetClientForCluster(testData.clusterNames[i]).Return(nil, nil)
			authPolicyClient.EXPECT().UpsertSpec(ctx, expectedAuthPolicy)
		}
		translatorError := istioTranslator.Translate(ctx, testData.targetServices, testData.accessControlPolicy)
		Expect(translatorError).To(BeNil())
	})

	It("use suffix wildcard if cluster specified and namespace omitted", func() {
		testData := initTestData()
		testData.accessControlPolicy.Spec.SourceSelector.GetMatcher().Namespaces = nil
		var expectedPrincipals []string
		var expectedAuthPolicies []*client_security_v1beta1.AuthorizationPolicy
		for _, trustDomain := range testData.trustDomains {
			expectedPrincipals = append(
				expectedPrincipals,
				fmt.Sprintf("%s/ns/*", trustDomain),
			)
		}
		for _, targetService := range testData.targetServices {
			expectedAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
				ObjectMeta: v1.ObjectMeta{
					Name:      testData.accessControlPolicy.GetName() + "-" + targetService.MeshService.Name,
					Namespace: targetService.MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
				},
				Spec: security_v1beta1.AuthorizationPolicy{
					Selector: &v1beta1.WorkloadSelector{
						MatchLabels: targetService.MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
					},
					Rules: []*security_v1beta1.Rule{
						{
							From: []*security_v1beta1.Rule_From{
								{
									Source: &security_v1beta1.Source{
										Principals: expectedPrincipals,
									},
								},
							},
							To: []*security_v1beta1.Rule_To{
								{
									Operation: &security_v1beta1.Operation{
										Ports:   testData.allowedPorts,
										Methods: methodsToString(testData.accessControlPolicy.Spec.GetAllowedMethods()),
										Paths:   testData.accessControlPolicy.Spec.GetAllowedPaths(),
									},
								},
							},
						},
					},
					Action: security_v1beta1.AuthorizationPolicy_ALLOW,
				},
			}
			expectedAuthPolicies = append(expectedAuthPolicies, expectedAuthPolicy)
		}
		for i, expectedAuthPolicy := range expectedAuthPolicies {
			dynamicClientGetter.EXPECT().GetClientForCluster(testData.clusterNames[i]).Return(nil, nil)
			authPolicyClient.EXPECT().UpsertSpec(ctx, expectedAuthPolicy)
		}
		translatorError := istioTranslator.Translate(ctx, testData.targetServices, testData.accessControlPolicy)
		Expect(translatorError).To(BeNil())
	})

	It("should use principal wildcard if user omits source selector", func() {
		clusterNames := []string{"cluster-name1", "cluster-name2"}
		trustDomains := []string{"cluster.local1", "cluster.local2"}
		acp := &networking_v1alpha1.AccessControlPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name:      "acp-name",
				Namespace: "acp-namespace",
			},
			Spec: networking_types.AccessControlPolicySpec{},
		}
		targetServices := []access_control_policy_translator.TargetService{
			{
				MeshService: &discovery_v1alpha1.MeshService{
					Spec: discovery_types.MeshServiceSpec{
						KubeService: &discovery_types.MeshServiceSpec_KubeService{
							WorkloadSelectorLabels: map[string]string{
								"k1a": "v1a", "k1b": "v1b",
							},
							Ref: &core_types.ResourceRef{Namespace: "namespace1"},
						},
					},
				},
				Mesh: &discovery_v1alpha1.Mesh{
					Spec: discovery_types.MeshSpec{
						MeshType: &discovery_types.MeshSpec_Istio{
							Istio: &discovery_types.MeshSpec_IstioMesh{
								CitadelInfo: &discovery_types.MeshSpec_IstioMesh_CitadelInfo{TrustDomain: trustDomains[0]},
							},
						},
						Cluster: &core_types.ResourceRef{Name: clusterNames[0]},
					},
				},
			},
		}
		expectedAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name:      acp.GetName() + "-" + targetServices[0].MeshService.Name,
				Namespace: targetServices[0].MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
			},
			Spec: security_v1beta1.AuthorizationPolicy{
				Selector: &v1beta1.WorkloadSelector{
					MatchLabels: targetServices[0].MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
				},
				Rules: []*security_v1beta1.Rule{
					{
						From: []*security_v1beta1.Rule_From{
							{
								Source: &security_v1beta1.Source{
									Principals: []string{"*"},
								},
							},
						},
						To: []*security_v1beta1.Rule_To{
							{
								Operation: &security_v1beta1.Operation{
									Methods: []string{"*"},
								},
							},
						},
					},
				},
				Action: security_v1beta1.AuthorizationPolicy_ALLOW,
			},
		}
		dynamicClientGetter.EXPECT().GetClientForCluster(clusterNames[0]).Return(nil, nil)
		authPolicyClient.EXPECT().UpsertSpec(ctx, expectedAuthPolicy).Return(nil)
		translatorError := istioTranslator.Translate(ctx, targetServices, acp)
		Expect(translatorError).To(BeNil())
	})

	It("should use From.Source.Namespaces if only Matcher.Namespaces specified (and cluster omitted)", func() {
		clusterNames := []string{"cluster-name1", "cluster-name2"}
		trustDomains := []string{"cluster.local1", "cluster.local2"}
		acp := &networking_v1alpha1.AccessControlPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name:      "acp-name",
				Namespace: "acp-namespace",
			},
			Spec: networking_types.AccessControlPolicySpec{
				SourceSelector: &core_types.IdentitySelector{
					IdentitySelectorType: &core_types.IdentitySelector_Matcher_{
						Matcher: &core_types.IdentitySelector_Matcher{
							Namespaces: []string{"foo"},
						},
					},
				},
			},
		}
		targetServices := []access_control_policy_translator.TargetService{
			{
				MeshService: &discovery_v1alpha1.MeshService{
					Spec: discovery_types.MeshServiceSpec{
						KubeService: &discovery_types.MeshServiceSpec_KubeService{
							WorkloadSelectorLabels: map[string]string{
								"k1a": "v1a", "k1b": "v1b",
							},
							Ref: &core_types.ResourceRef{Namespace: "namespace1"},
						},
					},
				},
				Mesh: &discovery_v1alpha1.Mesh{
					Spec: discovery_types.MeshSpec{
						MeshType: &discovery_types.MeshSpec_Istio{
							Istio: &discovery_types.MeshSpec_IstioMesh{
								CitadelInfo: &discovery_types.MeshSpec_IstioMesh_CitadelInfo{TrustDomain: trustDomains[0]},
							},
						},
						Cluster: &core_types.ResourceRef{Name: clusterNames[0]},
					},
				},
			},
		}
		expectedAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name:      acp.GetName() + "-" + targetServices[0].MeshService.Name,
				Namespace: targetServices[0].MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
			},
			Spec: security_v1beta1.AuthorizationPolicy{
				Selector: &v1beta1.WorkloadSelector{
					MatchLabels: targetServices[0].MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
				},
				Rules: []*security_v1beta1.Rule{
					{
						From: []*security_v1beta1.Rule_From{
							{
								Source: &security_v1beta1.Source{
									Namespaces: acp.Spec.GetSourceSelector().GetMatcher().GetNamespaces(),
								},
							},
						},
						To: []*security_v1beta1.Rule_To{
							{
								Operation: &security_v1beta1.Operation{
									Methods: []string{"*"},
								},
							},
						},
					},
				},
				Action: security_v1beta1.AuthorizationPolicy_ALLOW,
			},
		}
		dynamicClientGetter.EXPECT().GetClientForCluster(clusterNames[0]).Return(nil, nil)
		authPolicyClient.EXPECT().UpsertSpec(ctx, expectedAuthPolicy).Return(nil)
		translatorError := istioTranslator.Translate(ctx, targetServices, acp)
		Expect(translatorError).To(BeNil())
	})

	It("should lookup service accounts by reference if specified", func() {
		testData := initTestData()
		testData.accessControlPolicy.Spec.SourceSelector = &core_types.IdentitySelector{
			IdentitySelectorType: &core_types.IdentitySelector_ServiceAccountRefs_{
				ServiceAccountRefs: &core_types.IdentitySelector_ServiceAccountRefs{
					ServiceAccounts: []*core_types.ResourceRef{
						{
							Name:      "name1",
							Namespace: "namespace1",
							Cluster:   testData.clusterNames[0],
						},
						{
							Name:      "name2",
							Namespace: "namespace2",
							Cluster:   testData.clusterNames[1],
						},
					},
				},
			},
		}
		// a call to meshClient.List is already included in initTestData, so we only need len(serviceAccounts) - 1
		meshClient.
			EXPECT().
			List(ctx).
			Return(&discovery_v1alpha1.MeshList{
				Items: []discovery_v1alpha1.Mesh{*testData.targetServices[0].Mesh, *testData.targetServices[1].Mesh},
			}, nil)
		var expectedPrincipals []string
		for i, serviceAccount := range testData.accessControlPolicy.Spec.SourceSelector.GetServiceAccountRefs().GetServiceAccounts() {
			expectedPrincipals = append(expectedPrincipals,
				fmt.Sprintf("%s/ns/%s/sa/%s", testData.trustDomains[i], serviceAccount.GetNamespace(), serviceAccount.GetName()))
		}
		var expectedAuthPolicies []*client_security_v1beta1.AuthorizationPolicy
		for _, targetService := range testData.targetServices {
			expectedAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
				ObjectMeta: v1.ObjectMeta{
					Name:      testData.accessControlPolicy.GetName() + "-" + targetService.MeshService.Name,
					Namespace: targetService.MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
				},
				Spec: security_v1beta1.AuthorizationPolicy{
					Selector: &v1beta1.WorkloadSelector{
						MatchLabels: targetService.MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
					},
					Rules: []*security_v1beta1.Rule{
						{
							From: []*security_v1beta1.Rule_From{
								{
									Source: &security_v1beta1.Source{
										Principals: expectedPrincipals,
									},
								},
							},
							To: []*security_v1beta1.Rule_To{
								{
									Operation: &security_v1beta1.Operation{
										Ports:   testData.allowedPorts,
										Methods: methodsToString(testData.accessControlPolicy.Spec.GetAllowedMethods()),
										Paths:   testData.accessControlPolicy.Spec.GetAllowedPaths(),
									},
								},
							},
						},
					},
					Action: security_v1beta1.AuthorizationPolicy_ALLOW,
				},
			}
			expectedAuthPolicies = append(expectedAuthPolicies, expectedAuthPolicy)
		}
		for i, expectedAuthPolicy := range expectedAuthPolicies {
			dynamicClientGetter.EXPECT().GetClientForCluster(testData.clusterNames[i]).Return(nil, nil)
			authPolicyClient.EXPECT().UpsertSpec(ctx, expectedAuthPolicy).Return(nil)
		}
		translatorError := istioTranslator.Translate(ctx, testData.targetServices, testData.accessControlPolicy)
		Expect(translatorError).To(BeNil())
	})

	It("should error if a service account ref doesn't match a real service account", func() {
		testData := initTestData()
		fakeRef := &core_types.ResourceRef{
			Name:      "name1",
			Namespace: "namespace1",
			Cluster:   "fake-cluster-name",
		}
		testData.accessControlPolicy.Spec.SourceSelector = &core_types.IdentitySelector{
			IdentitySelectorType: &core_types.IdentitySelector_ServiceAccountRefs_{
				ServiceAccountRefs: &core_types.IdentitySelector_ServiceAccountRefs{
					ServiceAccounts: []*core_types.ResourceRef{fakeRef},
				},
			},
		}
		translatorError := istioTranslator.Translate(ctx, testData.targetServices, testData.accessControlPolicy)
		Expect(translatorError.ErrorMessage).To(ContainSubstring(istio_translator.ServiceAccountRefNonexistent(fakeRef).Error()))
	})

	It("should return ACP processing error", func() {
		acp := &networking_v1alpha1.AccessControlPolicy{
			Spec: networking_types.AccessControlPolicySpec{
				SourceSelector: &core_types.IdentitySelector{
					IdentitySelectorType: &core_types.IdentitySelector_Matcher_{
						Matcher: &core_types.IdentitySelector_Matcher{
							Namespaces: []string{"foo"},
							Clusters:   []string{"cluster"},
						},
					},
				},
				AllowedPorts: []uint32{8080},
			},
		}
		testErr := eris.New("processing error")
		meshClient.
			EXPECT().
			List(ctx).
			Return(nil, testErr)
		expectedTranslatorError := &networking_types.AccessControlPolicyStatus_TranslatorError{
			TranslatorId: istio_translator.TranslatorId,
			ErrorMessage: istio_translator.ACPProcessingError(testErr, acp).Error(),
		}
		translatorError := istioTranslator.Translate(ctx, nil, acp)
		Expect(translatorError).To(Equal(expectedTranslatorError))
	})

	It("should return ACP upsert error", func() {
		testErr := eris.New("processing error")
		testData := initTestData()
		var expectedPrincipals []string
		for _, trustDomain := range testData.trustDomains {
			for _, namespace := range testData.namespaces {
				expectedPrincipals = append(
					expectedPrincipals,
					fmt.Sprintf("%s/ns/%s/sa/*", trustDomain, namespace),
				)
			}
		}
		expectedAuthPolicies := []*client_security_v1beta1.AuthorizationPolicy{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      testData.accessControlPolicy.GetName() + "-" + testData.targetServices[0].MeshService.Name,
					Namespace: testData.targetServices[0].MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
				},
				Spec: security_v1beta1.AuthorizationPolicy{
					Selector: &v1beta1.WorkloadSelector{
						MatchLabels: testData.targetServices[0].MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
					},
					Rules: []*security_v1beta1.Rule{
						{
							From: []*security_v1beta1.Rule_From{
								{
									Source: &security_v1beta1.Source{
										Principals: expectedPrincipals,
									},
								},
							},
							To: []*security_v1beta1.Rule_To{
								{
									Operation: &security_v1beta1.Operation{
										Ports:   testData.allowedPorts,
										Methods: methodsToString(testData.accessControlPolicy.Spec.GetAllowedMethods()),
										Paths:   testData.accessControlPolicy.Spec.GetAllowedPaths(),
									},
								},
							},
						},
					},
					Action: security_v1beta1.AuthorizationPolicy_ALLOW,
				},
			},
		}
		for i, _ := range testData.targetServices {
			dynamicClientGetter.EXPECT().GetClientForCluster(testData.clusterNames[i]).Return(nil, nil)
		}
		authPolicyClient.EXPECT().UpsertSpec(ctx, expectedAuthPolicies[0]).Return(testErr)
		expectedTranslatorError := &networking_types.AccessControlPolicyStatus_TranslatorError{
			TranslatorId: istio_translator.TranslatorId,
			ErrorMessage: istio_translator.AuthPolicyUpsertError(testErr, expectedAuthPolicies[0]).Error(),
		}
		translatorError := istioTranslator.Translate(ctx, testData.targetServices, testData.accessControlPolicy)
		Expect(translatorError).To(Equal(expectedTranslatorError))
	})
})
