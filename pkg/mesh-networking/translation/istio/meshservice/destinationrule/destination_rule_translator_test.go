package destinationrule_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	mock_decorators "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/mocks"
	mock_trafficpolicy "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/destinationrule"
	mock_hostutils "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("DestinationRuleTranslator", func() {
	var (
		ctrl                      *gomock.Controller
		mockClusterDomainRegistry *mock_hostutils.MockClusterDomainRegistry
		mockDecoratorFactory      *mock_decorators.MockFactory
		meshServices              v1alpha2sets.MeshServiceSet
		mockReporter              *mock_reporting.MockReporter
		mockDecorator             *mock_trafficpolicy.MockTrafficPolicyDestinationRuleDecorator
		destinationRuleTranslator destinationrule.Translator
		in                        input.Snapshot
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClusterDomainRegistry = mock_hostutils.NewMockClusterDomainRegistry(ctrl)
		mockDecoratorFactory = mock_decorators.NewMockFactory(ctrl)
		meshServices = v1alpha2sets.NewMeshServiceSet()
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockDecorator = mock_trafficpolicy.NewMockTrafficPolicyDestinationRuleDecorator(ctrl)
		destinationRuleTranslator = destinationrule.NewTranslator(
			mockClusterDomainRegistry,
			mockDecoratorFactory,
			meshServices,
		)
		in = input.NewInputSnapshotManualBuilder("").Build()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate", func() {
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
					},
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-2",
							Namespace: "tp-namespace-2",
						},
					},
				},
			},
		}

		meshServices.Insert(&discoveryv1alpha2.MeshService{
			ObjectMeta: metav1.ObjectMeta{
				Name: "another-mesh-service",
			},
			Spec: discoveryv1alpha2.MeshServiceSpec{
				Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
					KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "another-mesh-service",
							Namespace:   "mesh-service-namespace",
							ClusterName: "mesh-service-cluster",
						},
					},
				},
			},
			Status: discoveryv1alpha2.MeshServiceStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "another-tp",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1alpha2.TrafficPolicySpec{
							TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
								Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
									{
										DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
											// original service
											Name:        meshService.Spec.GetKubeService().GetRef().Name,
											Namespace:   meshService.Spec.GetKubeService().GetRef().Namespace,
											ClusterName: meshService.Spec.GetKubeService().GetRef().ClusterName,

											Subset: map[string]string{"foo": "bar"},
										}},
									},
								},
							},
						},
					},
				},
			},
		})

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		mockClusterDomainRegistry.
			EXPECT().
			GetServiceLocalFQDN(meshService.Spec.GetKubeService().Ref).
			Return("local-hostname")

		initializedDestinatonRule := &networkingv1alpha3.DestinationRule{
			ObjectMeta: metautils.TranslatedObjectMeta(
				meshService.Spec.GetKubeService().Ref,
				meshService.Annotations,
			),
			Spec: networkingv1alpha3spec.DestinationRule{
				Host: "local-hostname",
				TrafficPolicy: &networkingv1alpha3spec.TrafficPolicy{
					Tls: &networkingv1alpha3spec.ClientTLSSettings{
						Mode: networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
				Subsets: []*networkingv1alpha3spec.Subset{
					{
						Name:   "foo-bar",
						Labels: map[string]string{"foo": "bar"},
					},
				},
			},
		}

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToDestinationRule(
				meshService.Status.AppliedTrafficPolicies[0],
				meshService,
				&initializedDestinatonRule.Spec,
				gomock.Any(),
			).
			Return(nil)
		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToDestinationRule(
				meshService.Status.AppliedTrafficPolicies[1],
				meshService,
				&initializedDestinatonRule.Spec,
				gomock.Any(),
			).
			Return(nil)

		destinationRule := destinationRuleTranslator.Translate(in, meshService, mockReporter)
		Expect(destinationRule).To(Equal(initializedDestinatonRule))
	})
})
