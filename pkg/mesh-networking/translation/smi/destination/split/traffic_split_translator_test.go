package split_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smislpitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	. "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/smi/destination/split"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TrafficSplitTranslator", func() {
	var (
		ctx          context.Context
		ctrl         *gomock.Controller
		mockReporter *mock_reporting.MockReporter
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockReporter = mock_reporting.NewMockReporter(ctrl)
	})

	It("will return nothing if no traffic policies are applied", func() {
		in := input.NewInputLocalSnapshotManualBuilder("").Build()
		destination := &discoveryv1alpha2.Destination{}

		ts := NewTranslator().Translate(ctx, in, destination, mockReporter)
		Expect(ts).To(BeNil())
	})

	It("can build a proper traffic shift", func() {
		ns := "default"
		in := input.NewInputLocalSnapshotManualBuilder("").Build()
		destination := &discoveryv1alpha2.Destination{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: discoveryv1alpha2.DestinationSpec{
				Type: &discoveryv1alpha2.DestinationSpec_KubeService_{
					KubeService: &discoveryv1alpha2.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:      "service",
							Namespace: ns,
						},
					},
				},
			},
			Status: discoveryv1alpha2.DestinationStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.DestinationStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tt",
							Namespace: ns,
						},
						Spec: &v1alpha2.TrafficPolicySpec{
							Policy: &v1alpha2.TrafficPolicySpec_Policy{
								TrafficShift: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination{
									Destinations: []*v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination{
										{
											DestinationType: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService{
												KubeService: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination{
													Name:      "one",
													Namespace: ns,
												},
											},
											Weight: 40,
										},
										{
											DestinationType: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService{
												KubeService: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination{
													Name:      "two",
													Namespace: ns,
												},
											},
											Weight: 60,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		expectedTT := &smislpitv1alpha2.TrafficSplit{
			ObjectMeta: metautils.TranslatedObjectMeta(
				destination.Spec.GetKubeService().Ref,
				destination.Annotations,
			),
			Spec: smislpitv1alpha2.TrafficSplitSpec{
				Service: "service.default",
				Backends: []smislpitv1alpha2.TrafficSplitBackend{
					{
						Service: "one",
						Weight:  40,
					},
					{
						Service: "two",
						Weight:  60,
					},
				},
			},
		}
		expectedTT.Annotations = map[string]string{
			metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1alpha2, Kind=TrafficPolicy":[{"name":"tt","namespace":"default"}]}`,
		}

		ts := NewTranslator().Translate(ctx, in, destination, mockReporter)
		Expect(ts).To(Equal(expectedTT))
	})

})
