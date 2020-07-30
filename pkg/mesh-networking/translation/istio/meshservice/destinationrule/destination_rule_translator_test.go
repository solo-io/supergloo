package destinationrule_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	mock_decorators "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators/mocks"
	mock_trafficpolicy "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/mocks"
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
		mockReporter              *mock_reporting.MockReporter
		mockAggregatingDecorator  *mock_trafficpolicy.MockAggregatingDestinationRuleDecorator
		mockDecorator             *mock_trafficpolicy.MockDestinationRuleDecorator
		destinationRuleTranslator destinationrule.Translator
		in                        input.Snapshot
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClusterDomainRegistry = mock_hostutils.NewMockClusterDomainRegistry(ctrl)
		mockDecoratorFactory = mock_decorators.NewMockFactory(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockAggregatingDecorator = mock_trafficpolicy.NewMockAggregatingDestinationRuleDecorator(ctrl)
		mockDecorator = mock_trafficpolicy.NewMockDestinationRuleDecorator(ctrl)
		destinationRuleTranslator = destinationrule.NewTranslator(mockClusterDomainRegistry, mockDecoratorFactory)
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

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockAggregatingDecorator, mockDecorator})

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
			},
		}

		mockAggregatingDecorator.
			EXPECT().
			ApplyAllToDestinationRule(
				meshService.Status.AppliedTrafficPolicies,
				&initializedDestinatonRule.Spec,
				gomock.Any(),
			).
			Return(nil)

		mockDecorator.
			EXPECT().
			ApplyToDestinationRule(
				meshService.Status.AppliedTrafficPolicies[0],
				meshService,
				&initializedDestinatonRule.Spec,
				gomock.Any(),
			).
			Return(nil)
		mockDecorator.
			EXPECT().
			ApplyToDestinationRule(
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
