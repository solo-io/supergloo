package destination

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	mock_output "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio/mocks"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	mock_authorizationpolicy "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/authorizationpolicy/mocks"
	mock_destinationrule "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/destinationrule/mocks"
	mock_federation "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/federation/mocks"
	mock_virtualservice "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/virtualservice/mocks"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("IstioDestinationTranslator", func() {
	var (
		ctrl                              *gomock.Controller
		mockDestinationRuleTranslator     *mock_destinationrule.MockTranslator
		mockVirtualServiceTranslator      *mock_virtualservice.MockTranslator
		mockAuthorizationPolicyTranslator *mock_authorizationpolicy.MockTranslator
		mockFederationTranslator          *mock_federation.MockTranslator
		mockOutputs                       *mock_output.MockBuilder
		mockReporter                      *mock_reporting.MockReporter
		istioDestinationTranslator        Translator
		ctx                               = context.TODO()
		eventObjs                         = map[schema.GroupVersionKind][]ezkube.ResourceId{}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDestinationRuleTranslator = mock_destinationrule.NewMockTranslator(ctrl)
		mockVirtualServiceTranslator = mock_virtualservice.NewMockTranslator(ctrl)
		mockAuthorizationPolicyTranslator = mock_authorizationpolicy.NewMockTranslator(ctrl)
		mockFederationTranslator = mock_federation.NewMockTranslator(ctrl)
		mockOutputs = mock_output.NewMockBuilder(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		istioDestinationTranslator = &translator{
			ctx:                   ctx,
			destinationRules:      mockDestinationRuleTranslator,
			virtualServices:       mockVirtualServiceTranslator,
			authorizationPolicies: mockAuthorizationPolicyTranslator,
			federation:            mockFederationTranslator,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate", func() {
		destination := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Mesh: &skv2corev1.ObjectRef{
					Name:      "hello",
					Namespace: "world",
				},
			},
		}

		in := input.NewInputLocalSnapshotManualBuilder("").
			AddMeshes([]*discoveryv1.Mesh{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      destination.Spec.GetMesh().GetName(),
						Namespace: destination.Spec.GetMesh().GetNamespace(),
					},
					Spec: discoveryv1.MeshSpec{
						Type: &discoveryv1.MeshSpec_Istio_{
							Istio: &discoveryv1.MeshSpec_Istio{},
						},
					},
				},
			}).
			Build()

		vs := &v1alpha3.VirtualService{}
		dr := &v1alpha3.DestinationRule{}
		ap := &v1beta1.AuthorizationPolicy{}
		federatedSe := []*v1alpha3.ServiceEntry{}
		federatedVs := []*v1alpha3.VirtualService{}
		federatedDr := []*v1alpha3.DestinationRule{}

		mockDestinationRuleTranslator.
			EXPECT().
			ShouldTranslate(destination, eventObjs).
			Return(true)
		mockDestinationRuleTranslator.
			EXPECT().
			Translate(ctx, in, destination, nil, mockReporter).
			Return(dr)

		mockVirtualServiceTranslator.
			EXPECT().
			ShouldTranslate(destination, eventObjs).
			Return(true)
		mockVirtualServiceTranslator.
			EXPECT().
			Translate(ctx, in, destination, nil, mockReporter).
			Return(vs)

		mockAuthorizationPolicyTranslator.
			EXPECT().
			ShouldTranslate(destination, eventObjs).
			Return(true)
		mockAuthorizationPolicyTranslator.
			EXPECT().
			Translate(ctx, in, destination, mockReporter).
			Return(ap)

		mockFederationTranslator.
			EXPECT().
			ShouldTranslate(destination, eventObjs).
			Return(true)
		mockFederationTranslator.
			EXPECT().
			Translate(ctx, in, destination, mockReporter).
			Return(federatedSe, federatedVs, federatedDr)

		mockOutputs.
			EXPECT().
			AddVirtualServices(vs)
		mockOutputs.
			EXPECT().
			AddDestinationRules(dr)
		mockOutputs.
			EXPECT().
			AddAuthorizationPolicies(ap)
		mockOutputs.
			EXPECT().
			AddServiceEntries(federatedSe)
		mockOutputs.
			EXPECT().
			AddVirtualServices(federatedVs)
		mockOutputs.
			EXPECT().
			AddDestinationRules(federatedDr)

		istioDestinationTranslator.Translate(eventObjs, in, destination, mockOutputs, mockReporter)
	})
})
