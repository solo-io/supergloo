package authorizationpolicy_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/input"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/types"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/authorizationpolicy"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	securityv1beta1spec "istio.io/api/security/v1beta1"
	"istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("AuthorizationPolicyTranslator", func() {
	var (
		ctrl         *gomock.Controller
		translator   authorizationpolicy.Translator
		mockReporter *mock_reporting.MockReporter
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		translator = authorizationpolicy.NewTranslator()
	})

	It("should translate a rule for each AccessPolicy applied to a MeshService", func() {
		meshService := &discoveryv1alpha2.MeshService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ms",
				Namespace: "ms-namespace",
			},
			Spec: discoveryv1alpha2.MeshServiceSpec{
				Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
					KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "kube-service",
							Namespace:   "kube-service-namespace",
							ClusterName: "cluster",
						},
						WorkloadSelectorLabels: map[string]string{
							"app": "kube-service",
						},
					},
				},
			},
			Status: discoveryv1alpha2.MeshServiceStatus{
				AppliedAccessPolicies: []*discoveryv1alpha2.MeshServiceStatus_AppliedAccessPolicy{
					{
						Spec: &networkingv1alpha2.AccessPolicySpec{
							SourceSelector: []*networkingv1alpha2.IdentitySelector{
								{
									KubeIdentityMatcher: &networkingv1alpha2.IdentitySelector_KubeIdentityMatcher{
										Namespaces: []string{"source-namespace1", "source-namespace2", "source-namespace3"},
										Clusters:   []string{"cluster1", "cluster2"},
									},
								},
							},
							AllowedPaths:   []string{"/path1", "/path2"},
							AllowedMethods: []types.HttpMethodValue{types.HttpMethodValue_GET, types.HttpMethodValue_POST},
							AllowedPorts:   []uint32{8080, 9080},
						},
					},
					{
						Spec: &networkingv1alpha2.AccessPolicySpec{
							SourceSelector: []*networkingv1alpha2.IdentitySelector{
								{
									KubeServiceAccountRefs: &networkingv1alpha2.IdentitySelector_KubeServiceAccountRefs{
										ServiceAccounts: []*v1.ClusterObjectRef{
											{
												Name:        "sa",
												Namespace:   "sa-namespace",
												ClusterName: "cluster1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		meshes := []*discoveryv1alpha2.Mesh{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mesh1",
				},
				Spec: discoveryv1alpha2.MeshSpec{
					MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
						Istio: &discoveryv1alpha2.MeshSpec_Istio{
							Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
								Cluster: "cluster1",
							},
							CitadelInfo: &discoveryv1alpha2.MeshSpec_Istio_CitadelInfo{
								TrustDomain: "cluster1.local",
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mesh2",
				},
				Spec: discoveryv1alpha2.MeshSpec{
					MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
						Istio: &discoveryv1alpha2.MeshSpec_Istio{
							Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
								Cluster: "cluster2",
							},
							CitadelInfo: &discoveryv1alpha2.MeshSpec_Istio_CitadelInfo{
								TrustDomain: "cluster2.local",
							},
						},
					},
				},
			},
		}
		expectedAuthPolicy := &securityv1beta1.AuthorizationPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:        meshService.Spec.GetKubeService().Ref.Name,
				Namespace:   meshService.Spec.GetKubeService().Ref.Namespace,
				ClusterName: meshService.Spec.GetKubeService().Ref.ClusterName,
				Labels: map[string]string{
					"owner.networking.smh.solo.io": "service-mesh-hub",
				},
			},
			Spec: securityv1beta1spec.AuthorizationPolicy{
				Selector: &v1beta1.WorkloadSelector{
					MatchLabels: meshService.Spec.GetKubeService().WorkloadSelectorLabels,
				},
				Rules: []*securityv1beta1spec.Rule{
					{
						From: []*securityv1beta1spec.Rule_From{
							{
								Source: &securityv1beta1spec.Source{
									Principals: []string{
										"cluster1.local/ns/source-namespace1/sa/*",
										"cluster1.local/ns/source-namespace2/sa/*",
										"cluster1.local/ns/source-namespace3/sa/*",
										"cluster2.local/ns/source-namespace1/sa/*",
										"cluster2.local/ns/source-namespace2/sa/*",
										"cluster2.local/ns/source-namespace3/sa/*",
									},
								},
							},
						},
						To: []*securityv1beta1spec.Rule_To{
							{
								Operation: &securityv1beta1spec.Operation{
									Ports:   []string{"8080", "9080"},
									Methods: []string{"GET", "POST"},
									Paths:   []string{"/path1", "/path2"},
								},
							},
						},
					},
					{
						From: []*securityv1beta1spec.Rule_From{
							{
								Source: &securityv1beta1spec.Source{
									Principals: []string{
										"cluster1.local/ns/sa-namespace/sa/sa",
									},
								},
							},
						},
					},
				},
				Action: securityv1beta1spec.AuthorizationPolicy_ALLOW,
			},
		}
		inputSnapshot := input.NewInputSnapshotManualBuilder("").AddMeshes(meshes).Build()
		authPolicy := translator.Translate(inputSnapshot, meshService, mockReporter)
		Expect(authPolicy).To(Equal(expectedAuthPolicy))
	})
})
