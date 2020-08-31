package access_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	"github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget/access"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/service-mesh-hub/test/matchers"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TrafficTargetTranslator", func() {
	var (
		ctrl     *gomock.Controller
		ctx      context.Context
		reporter *mock_reporting.MockReporter
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		reporter = mock_reporting.NewMockReporter(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will report an error if no backing workloads exist", func() {
		in := input.NewInputSnapshotManualBuilder("").Build()

		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedAccessPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedAccessPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "hello",
							Namespace: "world",
						},
					},
				},
			},
		}

		reporter.
			EXPECT().
			ReportAccessPolicyToTrafficTarget(
				trafficTarget,
				trafficTarget.Status.AppliedAccessPolicies[0].Ref,
				NoServiceAccountError,
			)

		tt, hrg := NewTranslator().Translate(ctx, in, trafficTarget, reporter)
		Expect(tt).To(HaveLen(0))
		Expect(hrg).To(HaveLen(0))

	})

	It("will report an error if backing workloads belong to multiple service accounts", func() {
		ns := "default"
		podLabels := map[string]string{"we": "match"}
		in := input.NewInputSnapshotManualBuilder("").
			AddWorkloads([]*discoveryv1alpha2.Workload{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "one",
					},
					Spec: discoveryv1alpha2.WorkloadSpec{
						WorkloadType: &discoveryv1alpha2.WorkloadSpec_Kubernetes{
							Kubernetes: &discoveryv1alpha2.WorkloadSpec_KubernertesWorkload{
								Controller: &v1.ClusterObjectRef{
									Namespace: ns,
								},
								PodLabels:          podLabels,
								ServiceAccountName: "hello",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "two",
					},
					Spec: discoveryv1alpha2.WorkloadSpec{
						WorkloadType: &discoveryv1alpha2.WorkloadSpec_Kubernetes{
							Kubernetes: &discoveryv1alpha2.WorkloadSpec_KubernertesWorkload{
								Controller: &v1.ClusterObjectRef{
									Namespace: ns,
								},
								PodLabels:          podLabels,
								ServiceAccountName: "world",
							},
						},
					},
				},
			}).
			Build()

		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Namespace: ns,
						},
						WorkloadSelectorLabels: podLabels,
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedAccessPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedAccessPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "hello",
							Namespace: "world",
						},
					},
				},
			},
		}

		reporter.
			EXPECT().
			ReportAccessPolicyToTrafficTarget(
				trafficTarget,
				trafficTarget.Status.AppliedAccessPolicies[0].Ref,
				matchers.MatchesError(CouldNotDetermineServiceAccountError(2)),
			)

		tt, hrg := NewTranslator().Translate(ctx, in, trafficTarget, reporter)
		Expect(tt).To(HaveLen(0))
		Expect(hrg).To(HaveLen(0))
	})

	It("can create a valid traffictarget/httproutegroup pair", func() {
		ns := "default"
		podLabels := map[string]string{"we": "match"}
		in := input.NewInputSnapshotManualBuilder("").
			AddWorkloads([]*discoveryv1alpha2.Workload{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "one",
					},
					Spec: discoveryv1alpha2.WorkloadSpec{
						WorkloadType: &discoveryv1alpha2.WorkloadSpec_Kubernetes{
							Kubernetes: &discoveryv1alpha2.WorkloadSpec_KubernertesWorkload{
								Controller: &v1.ClusterObjectRef{
									Namespace: ns,
								},
								PodLabels:          podLabels,
								ServiceAccountName: "hello",
							},
						},
					},
				},
			}).
			Build()

		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: ns,
			},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:      "name",
							Namespace: ns,
						},
						WorkloadSelectorLabels: podLabels,
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedAccessPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedAccessPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "hello",
							Namespace: "world",
						},
						Spec: &networkingv1alpha2.AccessPolicySpec{
							SourceSelector: []*networkingv1alpha2.IdentitySelector{
								{
									KubeServiceAccountRefs: &networkingv1alpha2.IdentitySelector_KubeServiceAccountRefs{
										ServiceAccounts: []*v1.ClusterObjectRef{
											{
												Name:      "sa",
												Namespace: ns,
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

		apRef := trafficTarget.Status.GetAppliedAccessPolicies()[0].GetRef()
		refernceString := fmt.Sprintf("%s.%s", apRef.GetName(), apRef.GetNamespace())

		expectedHRG := &v1alpha3.HTTPRouteGroup{
			ObjectMeta: metautils.TranslatedObjectMeta(
				trafficTarget.Spec.GetKubeService().Ref,
				trafficTarget.Annotations,
			),
			Spec: v1alpha3.HTTPRouteGroupSpec{
				Matches: []v1alpha3.HTTPMatch{
					{
						Name:      refernceString,
						Methods:   []string{string(v1alpha3.HTTPRouteMethodAll)},
						PathRegex: constants.RegexMatchAll,
					},
				},
			},
		}
		expectedHRG.Name += "." + refernceString

		expectedTT := &v1alpha2.TrafficTarget{
			ObjectMeta: metautils.TranslatedObjectMeta(
				trafficTarget.Spec.GetKubeService().Ref,
				trafficTarget.Annotations,
			),
			Spec: v1alpha2.TrafficTargetSpec{
				Destination: v1alpha2.IdentityBindingSubject{
					Kind:      "ServiceAccount",
					Name:      "hello",
					Namespace: ns,
				},
				Sources: []v1alpha2.IdentityBindingSubject{
					{
						Kind:      "ServiceAccount",
						Name:      "sa",
						Namespace: ns,
					},
				},
				Rules: []v1alpha2.TrafficTargetRule{
					{
						Kind:    "HTTPRouteGroup",
						Name:    expectedHRG.GetName(),
						Matches: []string{expectedHRG.Spec.Matches[0].Name},
					},
				},
			},
		}
		expectedTT.Name += "." + refernceString

		tt, hrg := NewTranslator().Translate(ctx, in, trafficTarget, reporter)
		Expect(tt).To(HaveLen(1))
		Expect(tt[0]).To(Equal(expectedTT))
		Expect(hrg).To(HaveLen(1))
		Expect(hrg[0]).To(Equal(expectedHRG))
	})

})
