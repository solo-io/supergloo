package istio_translator_test

import (
	"context"
	"fmt"

	istio_security "github.com/solo-io/external-apis/pkg/api/istio/security.istio.io/v1beta1"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	mock_istio_security "github.com/solo-io/external-apis/pkg/api/istio/security.istio.io/v1beta1/mocks"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	mock_core "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/mocks"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	mock_multicluster "github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster/mocks"
	access_control_policy_translator "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-policy-translator"
	istio_translator "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-policy-translator/istio-translator"
	istio_security_types "istio.io/api/security/v1beta1"
	istio_types "istio.io/api/type/v1beta1"
	istio_security_client_types "istio.io/client-go/pkg/apis/security/v1beta1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("IstioTranslator", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 context.Context
		authPolicyClient    *mock_istio_security.MockAuthorizationPolicyClient
		meshClient          *mock_core.MockMeshClient
		istioTranslator     istio_translator.IstioTranslator
		dynamicClientGetter *mock_multicluster.MockDynamicClientGetter
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		authPolicyClient = mock_istio_security.NewMockAuthorizationPolicyClient(ctrl)
		meshClient = mock_core.NewMockMeshClient(ctrl)
		dynamicClientGetter = mock_multicluster.NewMockDynamicClientGetter(ctrl)
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
		accessControlPolicy *smh_networking.AccessControlPolicy
		targetServices      []access_control_policy_translator.TargetService
		clusterNames        []string
		acpClusterNames     []string
		trustDomains        []string
		namespaces          []string
		allowedPaths        []string
		allowedMethods      []smh_core_types.HttpMethodValue
		allowedPorts        []string
	}

	// convenience method for converting HttpMethod enum to string
	var methodsToString = func(methodEnums []smh_core_types.HttpMethodValue) []string {
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
		allowedMethods := []smh_core_types.HttpMethodValue{smh_core_types.HttpMethodValue_GET, smh_core_types.HttpMethodValue_POST}
		allowedPortsInts := []uint32{8080, 9080}
		allowedPortsString := []string{"8080", "9080"}
		istioMesh1 := &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
					Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
						Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
							CitadelInfo: &smh_discovery_types.MeshSpec_IstioMesh_CitadelInfo{TrustDomain: trustDomains[0]},
						},
					},
				},
				Cluster: &smh_core_types.ResourceRef{Name: clusterNames[0]},
			},
		}
		istioMesh2 := &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
					Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
						Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
							CitadelInfo: &smh_discovery_types.MeshSpec_IstioMesh_CitadelInfo{TrustDomain: trustDomains[1]},
						},
					},
				},
				Cluster: &smh_core_types.ResourceRef{Name: clusterNames[1]},
			},
		}
		acp := &smh_networking.AccessControlPolicy{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "acp-name",
				Namespace: "acp-namespace",
			},
			Spec: smh_networking_types.AccessControlPolicySpec{
				SourceSelector: &smh_core_types.IdentitySelector{
					IdentitySelectorType: &smh_core_types.IdentitySelector_Matcher_{
						Matcher: &smh_core_types.IdentitySelector_Matcher{
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
				MeshService: &smh_discovery.MeshService{
					Spec: smh_discovery_types.MeshServiceSpec{
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							WorkloadSelectorLabels: map[string]string{
								"k1a": "v1a", "k1b": "v1b",
							},
							Ref: &smh_core_types.ResourceRef{Namespace: "namespace1"},
						},
					},
				},
				Mesh: istioMesh1,
			},
			{
				MeshService: &smh_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Namespace: "namespace2",
					},
					Spec: smh_discovery_types.MeshServiceSpec{
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							WorkloadSelectorLabels: map[string]string{
								"k2a": "v2a", "k2b": "v2b",
							},
							Ref: &smh_core_types.ResourceRef{Namespace: "namespace2"},
						},
					},
				},
				Mesh: istioMesh2,
			},
		}
		meshClient.
			EXPECT().
			ListMesh(ctx).
			Return(&smh_discovery.MeshList{
				Items: []smh_discovery.Mesh{*istioMesh1, *istioMesh2},
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
		var expectedAuthPolicies []*istio_security_client_types.AuthorizationPolicy
		for _, trustDomain := range testData.trustDomains {
			for _, namespace := range testData.namespaces {
				expectedPrincipals = append(
					expectedPrincipals,
					fmt.Sprintf("%s/ns/%s/sa/*", trustDomain, namespace),
				)
			}
		}
		for _, targetService := range testData.targetServices {
			expectedAuthPolicy := &istio_security_client_types.AuthorizationPolicy{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      testData.accessControlPolicy.GetName() + "-" + targetService.MeshService.Name,
					Namespace: targetService.MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
				},
				Spec: istio_security_types.AuthorizationPolicy{
					Selector: &istio_types.WorkloadSelector{
						MatchLabels: targetService.MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
					},
					Rules: []*istio_security_types.Rule{
						{
							From: []*istio_security_types.Rule_From{
								{
									Source: &istio_security_types.Source{
										Principals: expectedPrincipals,
									},
								},
							},
							To: []*istio_security_types.Rule_To{
								{
									Operation: &istio_security_types.Operation{
										Ports:   testData.allowedPorts,
										Methods: methodsToString(testData.accessControlPolicy.Spec.GetAllowedMethods()),
										Paths:   testData.accessControlPolicy.Spec.GetAllowedPaths(),
									},
								},
							},
						},
					},
					Action: istio_security_types.AuthorizationPolicy_ALLOW,
				},
			}
			expectedAuthPolicies = append(expectedAuthPolicies, expectedAuthPolicy)
		}
		for i, expectedAuthPolicy := range expectedAuthPolicies {
			dynamicClientGetter.EXPECT().GetClientForCluster(ctx, testData.clusterNames[i]).Return(nil, nil)
			authPolicyClient.EXPECT().UpsertAuthorizationPolicy(ctx, expectedAuthPolicy)
		}
		translatorError := istioTranslator.Translate(ctx, testData.targetServices, testData.accessControlPolicy)
		Expect(translatorError).To(BeNil())
	})

	It("use suffix wildcard if cluster specified and namespace omitted", func() {
		testData := initTestData()
		testData.accessControlPolicy.Spec.SourceSelector.GetMatcher().Namespaces = nil
		var expectedPrincipals []string
		var expectedAuthPolicies []*istio_security_client_types.AuthorizationPolicy
		for _, trustDomain := range testData.trustDomains {
			expectedPrincipals = append(
				expectedPrincipals,
				fmt.Sprintf("%s/ns/*", trustDomain),
			)
		}
		for _, targetService := range testData.targetServices {
			expectedAuthPolicy := &istio_security_client_types.AuthorizationPolicy{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      testData.accessControlPolicy.GetName() + "-" + targetService.MeshService.Name,
					Namespace: targetService.MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
				},
				Spec: istio_security_types.AuthorizationPolicy{
					Selector: &istio_types.WorkloadSelector{
						MatchLabels: targetService.MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
					},
					Rules: []*istio_security_types.Rule{
						{
							From: []*istio_security_types.Rule_From{
								{
									Source: &istio_security_types.Source{
										Principals: expectedPrincipals,
									},
								},
							},
							To: []*istio_security_types.Rule_To{
								{
									Operation: &istio_security_types.Operation{
										Ports:   testData.allowedPorts,
										Methods: methodsToString(testData.accessControlPolicy.Spec.GetAllowedMethods()),
										Paths:   testData.accessControlPolicy.Spec.GetAllowedPaths(),
									},
								},
							},
						},
					},
					Action: istio_security_types.AuthorizationPolicy_ALLOW,
				},
			}
			expectedAuthPolicies = append(expectedAuthPolicies, expectedAuthPolicy)
		}
		for i, expectedAuthPolicy := range expectedAuthPolicies {
			dynamicClientGetter.EXPECT().GetClientForCluster(ctx, testData.clusterNames[i]).Return(nil, nil)
			authPolicyClient.EXPECT().UpsertAuthorizationPolicy(ctx, expectedAuthPolicy)
		}
		translatorError := istioTranslator.Translate(ctx, testData.targetServices, testData.accessControlPolicy)
		Expect(translatorError).To(BeNil())
	})

	It("should use principal wildcard if user omits source selector", func() {
		clusterNames := []string{"cluster-name1", "cluster-name2"}
		trustDomains := []string{"cluster.local1", "cluster.local2"}
		acp := &smh_networking.AccessControlPolicy{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "acp-name",
				Namespace: "acp-namespace",
			},
			Spec: smh_networking_types.AccessControlPolicySpec{},
		}
		targetServices := []access_control_policy_translator.TargetService{
			{
				MeshService: &smh_discovery.MeshService{
					Spec: smh_discovery_types.MeshServiceSpec{
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							WorkloadSelectorLabels: map[string]string{
								"k1a": "v1a", "k1b": "v1b",
							},
							Ref: &smh_core_types.ResourceRef{Namespace: "namespace1"},
						},
					},
				},
				Mesh: &smh_discovery.Mesh{
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
							Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
								Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
									CitadelInfo: &smh_discovery_types.MeshSpec_IstioMesh_CitadelInfo{TrustDomain: trustDomains[0]},
								},
							},
						},
						Cluster: &smh_core_types.ResourceRef{Name: clusterNames[0]},
					},
				},
			},
		}
		expectedAuthPolicy := &istio_security_client_types.AuthorizationPolicy{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      acp.GetName() + "-" + targetServices[0].MeshService.Name,
				Namespace: targetServices[0].MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
			},
			Spec: istio_security_types.AuthorizationPolicy{
				Selector: &istio_types.WorkloadSelector{
					MatchLabels: targetServices[0].MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
				},
				Rules: []*istio_security_types.Rule{
					{
						From: []*istio_security_types.Rule_From{
							{
								Source: &istio_security_types.Source{
									Principals: []string{"*"},
								},
							},
						},
						To: []*istio_security_types.Rule_To{
							{
								Operation: &istio_security_types.Operation{
									Methods: []string{"*"},
								},
							},
						},
					},
				},
				Action: istio_security_types.AuthorizationPolicy_ALLOW,
			},
		}
		dynamicClientGetter.EXPECT().GetClientForCluster(ctx, clusterNames[0]).Return(nil, nil)
		authPolicyClient.EXPECT().UpsertAuthorizationPolicy(ctx, expectedAuthPolicy).Return(nil)
		translatorError := istioTranslator.Translate(ctx, targetServices, acp)
		Expect(translatorError).To(BeNil())
	})

	It("should use From.Source.Namespaces if only Matcher.Namespaces specified (and cluster omitted)", func() {
		clusterNames := []string{"cluster-name1", "cluster-name2"}
		trustDomains := []string{"cluster.local1", "cluster.local2"}
		acp := &smh_networking.AccessControlPolicy{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "acp-name",
				Namespace: "acp-namespace",
			},
			Spec: smh_networking_types.AccessControlPolicySpec{
				SourceSelector: &smh_core_types.IdentitySelector{
					IdentitySelectorType: &smh_core_types.IdentitySelector_Matcher_{
						Matcher: &smh_core_types.IdentitySelector_Matcher{
							Namespaces: []string{"foo"},
						},
					},
				},
			},
		}
		targetServices := []access_control_policy_translator.TargetService{
			{
				MeshService: &smh_discovery.MeshService{
					Spec: smh_discovery_types.MeshServiceSpec{
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							WorkloadSelectorLabels: map[string]string{
								"k1a": "v1a", "k1b": "v1b",
							},
							Ref: &smh_core_types.ResourceRef{Namespace: "namespace1"},
						},
					},
				},
				Mesh: &smh_discovery.Mesh{
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
							Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
								Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
									CitadelInfo: &smh_discovery_types.MeshSpec_IstioMesh_CitadelInfo{TrustDomain: trustDomains[0]},
								},
							},
						},
						Cluster: &smh_core_types.ResourceRef{Name: clusterNames[0]},
					},
				},
			},
		}
		expectedAuthPolicy := &istio_security_client_types.AuthorizationPolicy{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      acp.GetName() + "-" + targetServices[0].MeshService.Name,
				Namespace: targetServices[0].MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
			},
			Spec: istio_security_types.AuthorizationPolicy{
				Selector: &istio_types.WorkloadSelector{
					MatchLabels: targetServices[0].MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
				},
				Rules: []*istio_security_types.Rule{
					{
						From: []*istio_security_types.Rule_From{
							{
								Source: &istio_security_types.Source{
									Namespaces: acp.Spec.GetSourceSelector().GetMatcher().GetNamespaces(),
								},
							},
						},
						To: []*istio_security_types.Rule_To{
							{
								Operation: &istio_security_types.Operation{
									Methods: []string{"*"},
								},
							},
						},
					},
				},
				Action: istio_security_types.AuthorizationPolicy_ALLOW,
			},
		}
		dynamicClientGetter.EXPECT().GetClientForCluster(ctx, clusterNames[0]).Return(nil, nil)
		authPolicyClient.EXPECT().UpsertAuthorizationPolicy(ctx, expectedAuthPolicy).Return(nil)
		translatorError := istioTranslator.Translate(ctx, targetServices, acp)
		Expect(translatorError).To(BeNil())
	})

	It("should lookup service accounts by reference if specified", func() {
		testData := initTestData()
		testData.accessControlPolicy.Spec.SourceSelector = &smh_core_types.IdentitySelector{
			IdentitySelectorType: &smh_core_types.IdentitySelector_ServiceAccountRefs_{
				ServiceAccountRefs: &smh_core_types.IdentitySelector_ServiceAccountRefs{
					ServiceAccounts: []*smh_core_types.ResourceRef{
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
		// a call to MeshClient.ListMesh is already included in initTestData, so we only need len(serviceAccounts) - 1
		meshClient.
			EXPECT().
			ListMesh(ctx).
			Return(&smh_discovery.MeshList{
				Items: []smh_discovery.Mesh{*testData.targetServices[0].Mesh, *testData.targetServices[1].Mesh},
			}, nil)
		var expectedPrincipals []string
		for i, serviceAccount := range testData.accessControlPolicy.Spec.SourceSelector.GetServiceAccountRefs().GetServiceAccounts() {
			expectedPrincipals = append(expectedPrincipals,
				fmt.Sprintf("%s/ns/%s/sa/%s", testData.trustDomains[i], serviceAccount.GetNamespace(), serviceAccount.GetName()))
		}
		var expectedAuthPolicies []*istio_security_client_types.AuthorizationPolicy
		for _, targetService := range testData.targetServices {
			expectedAuthPolicy := &istio_security_client_types.AuthorizationPolicy{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      testData.accessControlPolicy.GetName() + "-" + targetService.MeshService.Name,
					Namespace: targetService.MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
				},
				Spec: istio_security_types.AuthorizationPolicy{
					Selector: &istio_types.WorkloadSelector{
						MatchLabels: targetService.MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
					},
					Rules: []*istio_security_types.Rule{
						{
							From: []*istio_security_types.Rule_From{
								{
									Source: &istio_security_types.Source{
										Principals: expectedPrincipals,
									},
								},
							},
							To: []*istio_security_types.Rule_To{
								{
									Operation: &istio_security_types.Operation{
										Ports:   testData.allowedPorts,
										Methods: methodsToString(testData.accessControlPolicy.Spec.GetAllowedMethods()),
										Paths:   testData.accessControlPolicy.Spec.GetAllowedPaths(),
									},
								},
							},
						},
					},
					Action: istio_security_types.AuthorizationPolicy_ALLOW,
				},
			}
			expectedAuthPolicies = append(expectedAuthPolicies, expectedAuthPolicy)
		}
		for i, expectedAuthPolicy := range expectedAuthPolicies {
			dynamicClientGetter.EXPECT().GetClientForCluster(ctx, testData.clusterNames[i]).Return(nil, nil)
			authPolicyClient.EXPECT().UpsertAuthorizationPolicy(ctx, expectedAuthPolicy).Return(nil)
		}
		translatorError := istioTranslator.Translate(ctx, testData.targetServices, testData.accessControlPolicy)
		Expect(translatorError).To(BeNil())
	})

	It("should error if a service account ref doesn't match a real service account", func() {
		testData := initTestData()
		fakeRef := &smh_core_types.ResourceRef{
			Name:      "name1",
			Namespace: "namespace1",
			Cluster:   "fake-cluster-name",
		}
		testData.accessControlPolicy.Spec.SourceSelector = &smh_core_types.IdentitySelector{
			IdentitySelectorType: &smh_core_types.IdentitySelector_ServiceAccountRefs_{
				ServiceAccountRefs: &smh_core_types.IdentitySelector_ServiceAccountRefs{
					ServiceAccounts: []*smh_core_types.ResourceRef{fakeRef},
				},
			},
		}
		translatorError := istioTranslator.Translate(ctx, testData.targetServices, testData.accessControlPolicy)
		Expect(translatorError.ErrorMessage).To(ContainSubstring(istio_translator.ServiceAccountRefNonexistent(fakeRef).Error()))
	})

	It("should return ACP processing error", func() {
		acp := &smh_networking.AccessControlPolicy{
			Spec: smh_networking_types.AccessControlPolicySpec{
				SourceSelector: &smh_core_types.IdentitySelector{
					IdentitySelectorType: &smh_core_types.IdentitySelector_Matcher_{
						Matcher: &smh_core_types.IdentitySelector_Matcher{
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
			ListMesh(ctx).
			Return(nil, testErr)
		expectedTranslatorError := &smh_networking_types.AccessControlPolicyStatus_TranslatorError{
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
		expectedAuthPolicies := []*istio_security_client_types.AuthorizationPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      testData.accessControlPolicy.GetName() + "-" + testData.targetServices[0].MeshService.Name,
					Namespace: testData.targetServices[0].MeshService.Spec.GetKubeService().GetRef().GetNamespace(),
				},
				Spec: istio_security_types.AuthorizationPolicy{
					Selector: &istio_types.WorkloadSelector{
						MatchLabels: testData.targetServices[0].MeshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
					},
					Rules: []*istio_security_types.Rule{
						{
							From: []*istio_security_types.Rule_From{
								{
									Source: &istio_security_types.Source{
										Principals: expectedPrincipals,
									},
								},
							},
							To: []*istio_security_types.Rule_To{
								{
									Operation: &istio_security_types.Operation{
										Ports:   testData.allowedPorts,
										Methods: methodsToString(testData.accessControlPolicy.Spec.GetAllowedMethods()),
										Paths:   testData.accessControlPolicy.Spec.GetAllowedPaths(),
									},
								},
							},
						},
					},
					Action: istio_security_types.AuthorizationPolicy_ALLOW,
				},
			},
		}
		for i, _ := range testData.targetServices {
			dynamicClientGetter.EXPECT().GetClientForCluster(ctx, testData.clusterNames[i]).Return(nil, nil)
		}
		authPolicyClient.EXPECT().UpsertAuthorizationPolicy(ctx, expectedAuthPolicies[0]).Return(testErr)
		expectedTranslatorError := &smh_networking_types.AccessControlPolicyStatus_TranslatorError{
			TranslatorId: istio_translator.TranslatorId,
			ErrorMessage: istio_translator.AuthPolicyUpsertError(testErr, expectedAuthPolicies[0]).Error(),
		}
		translatorError := istioTranslator.Translate(ctx, testData.targetServices, testData.accessControlPolicy)
		Expect(translatorError).To(Equal(expectedTranslatorError))
	})
})
