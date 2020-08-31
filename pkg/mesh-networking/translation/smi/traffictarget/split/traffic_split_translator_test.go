package split_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smislpitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget/split"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
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
		in := input.NewInputSnapshotManualBuilder("").Build()
		trafficTarget := &discoveryv1alpha2.TrafficTarget{}

		ts := NewTranslator().Translate(ctx, in, trafficTarget, mockReporter)
		Expect(ts).To(BeNil())
	})

	It("can build a proper traffic shift", func() {
		ns := "default"
		in := input.NewInputSnapshotManualBuilder("").Build()
		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:      "service",
							Namespace: ns,
						},
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tt",
							Namespace: ns,
						},
						Spec: &v1alpha2.TrafficPolicySpec{
							TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
								Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
									{
										DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
											KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
												Name:      "one",
												Namespace: ns,
											},
										},
										Weight: 40,
									},
									{
										DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
											KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
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
		}

		expectedTT := &smislpitv1alpha2.TrafficSplit{
			ObjectMeta: metautils.TranslatedObjectMeta(
				trafficTarget.Spec.GetKubeService().Ref,
				trafficTarget.Annotations,
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

		ts := NewTranslator().Translate(ctx, in, trafficTarget, mockReporter)
		Expect(ts).To(Equal(expectedTT))
	})

})
