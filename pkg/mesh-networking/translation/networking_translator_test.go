package translation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation"
	mock_appmesh "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/appmesh/mocks"
	mock_istio "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mocks"
	mock_osm "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/osm/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("NetworkingTranslator", func() {
	var (
		ctrl                  *gomock.Controller
		ctx                   context.Context
		mockReporter          *mock_reporting.MockReporter
		mockIstioTranslator   *mock_istio.MockTranslator
		mockAppMeshTranslator *mock_appmesh.MockTranslator
		mockOsmTranslator     *mock_osm.MockTranslator
		networkingTranslator  translation.Translator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockIstioTranslator = mock_istio.NewMockTranslator(ctrl)
		mockAppMeshTranslator = mock_appmesh.NewMockTranslator(ctrl)
		mockOsmTranslator = mock_osm.NewMockTranslator(ctrl)
		networkingTranslator = translation.NewTranslator(mockIstioTranslator, mockAppMeshTranslator, mockOsmTranslator)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("first translation should initialize outputs", func() {
		var eventObjs []ezkube.ResourceId
		in := input.NewInputLocalSnapshotManualBuilder("").Build()
		userSupplied := input.NewInputRemoteSnapshotManualBuilder("").Build()

		mockIstioTranslator.
			EXPECT().
			Translate(gomock.Any(), eventObjs, in, userSupplied, gomock.Any(), gomock.Any(), mockReporter)
		mockAppMeshTranslator.
			EXPECT().
			Translate(gomock.Any(), in, gomock.Any(), mockReporter)
		mockOsmTranslator.
			EXPECT().
			Translate(gomock.Any(), in, gomock.Any(), mockReporter)

		_, err := networkingTranslator.Translate(ctx, eventObjs, in, userSupplied, mockReporter)
		Expect(err).To(Not(HaveOccurred()))
	})

	It("subsequent translations should update outputs with new objects and garbage collect orphaned objects", func() {
		destination1 := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "destination1",
				Namespace: "ns",
			},
		}
		trafficPolicy1 := &networkingv1.TrafficPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "trafficpolicy1",
				Namespace: "ns",
			},
		}
		trafficPolicy2 := &networkingv1.TrafficPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "trafficpolicy2",
				Namespace: "ns",
			},
		}

		vs1 := virtualService(
			"vs1",
			[]string{"host1"},
			map[schema.GroupVersionKind][]ezkube.ResourceId{
				discoveryv1.DestinationGVK: {destination1},
			},
		)
		vs1updated := virtualService(
			"vs1",
			[]string{"host1", "host2"},
			map[schema.GroupVersionKind][]ezkube.ResourceId{
				discoveryv1.DestinationGVK: {destination1},
			},
		)
		dr1 := destinationRule(
			"dr1",
			map[schema.GroupVersionKind][]ezkube.ResourceId{
				discoveryv1.DestinationGVK:    {destination1},
				networkingv1.TrafficPolicyGVK: {trafficPolicy1},
			},
		)
		dr2 := destinationRule(
			"dr2",
			map[schema.GroupVersionKind][]ezkube.ResourceId{
				discoveryv1.DestinationGVK:    {destination1},
				networkingv1.TrafficPolicyGVK: {trafficPolicy1, trafficPolicy2},
			},
		)

		var eventObjs []ezkube.ResourceId
		in1 := input.NewInputLocalSnapshotManualBuilder("").
			AddDestinations([]*discoveryv1.Destination{destination1}).
			AddTrafficPolicies([]*networkingv1.TrafficPolicy{trafficPolicy1}).
			Build()
		userSupplied := input.NewInputRemoteSnapshotManualBuilder("").Build()

		mockIstioTranslator.
			EXPECT().
			Translate(gomock.Any(), eventObjs, in1, userSupplied, gomock.Any(), gomock.Any(), mockReporter).
			Do(func(
				ctx context.Context,
				eventObjs []ezkube.ResourceId,
				in input.LocalSnapshot,
				userSupplied input.RemoteSnapshot,
				istioOutputs istio.Builder,
				localOutputs local.Builder,
				reporter reporting.Reporter,
			) {
				istioOutputs.AddVirtualServices(vs1)
				istioOutputs.AddDestinationRules(dr1, dr2)
			})
		mockAppMeshTranslator.
			EXPECT().
			Translate(gomock.Any(), in1, gomock.Any(), mockReporter)
		mockOsmTranslator.
			EXPECT().
			Translate(gomock.Any(), in1, gomock.Any(), mockReporter)

		// first translation should initialize all outputs
		outputs, err := networkingTranslator.Translate(ctx, eventObjs, in1, userSupplied, mockReporter)
		Expect(err).To(Not(HaveOccurred()))

		Expect(outputs.Istio.GetVirtualServices().List()).To(ConsistOf([]*networkingv1alpha3.VirtualService{vs1}))
		Expect(outputs.Istio.GetDestinationRules().List()).To(ConsistOf([]*networkingv1alpha3.DestinationRule{dr1, dr2}))

		// remove the traffic policy, should cause dr1 to be garbage collected
		in2 := input.NewInputLocalSnapshotManualBuilder("").
			AddDestinations([]*discoveryv1.Destination{destination1}).
			AddTrafficPolicies([]*networkingv1.TrafficPolicy{trafficPolicy2}).
			Build()

		mockIstioTranslator.
			EXPECT().
			Translate(gomock.Any(), eventObjs, in2, userSupplied, gomock.Any(), gomock.Any(), mockReporter).
			Do(func(
				ctx context.Context,
				eventObjs []ezkube.ResourceId,
				in input.LocalSnapshot,
				userSupplied input.RemoteSnapshot,
				istioOutputs istio.Builder,
				localOutputs local.Builder,
				reporter reporting.Reporter,
			) {
				istioOutputs.AddVirtualServices(vs1updated)
			})
		mockAppMeshTranslator.
			EXPECT().
			Translate(gomock.Any(), in2, gomock.Any(), mockReporter)
		mockOsmTranslator.
			EXPECT().
			Translate(gomock.Any(), in2, gomock.Any(), mockReporter)

		// second translation should update existing outputs and garbage collect orphaned outputs
		outputs, err = networkingTranslator.Translate(ctx, eventObjs, in2, userSupplied, mockReporter)
		Expect(err).To(Not(HaveOccurred()))

		Expect(outputs.Istio.GetVirtualServices().List()).To(ConsistOf([]*networkingv1alpha3.VirtualService{vs1updated}))
		Expect(outputs.Istio.GetDestinationRules().List()).To(ConsistOf([]*networkingv1alpha3.DestinationRule{dr2}))
	})
})

func virtualService(
	name string,
	hosts []string,
	parents map[schema.GroupVersionKind][]ezkube.ResourceId,
) *networkingv1alpha3.VirtualService {
	vs := &networkingv1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "namespace",
		},
		Spec: networkingv1alpha3spec.VirtualService{
			Hosts: hosts,
		},
	}

	metautils.AnnotateParents(context.TODO(), vs, parents)

	return vs
}

func destinationRule(
	name string,
	parents map[schema.GroupVersionKind][]ezkube.ResourceId,
) *networkingv1alpha3.DestinationRule {
	dr := &networkingv1alpha3.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "namespace",
		},
	}

	metautils.AnnotateParents(context.TODO(), dr, parents)

	return dr
}
