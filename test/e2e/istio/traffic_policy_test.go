package istio_test

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/service-mesh-hub/test/e2e"
	"github.com/solo-io/service-mesh-hub/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/solo-io/service-mesh-hub/test/data"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TrafficPolicy", func() {
	var (
		err      error
		manifest utils.Manifest
		ctx      = context.Background()
	)

	BeforeEach(func() {
		manifest, err = utils.NewManifest("bookinfo-policies.yaml")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		manifest.Cleanup(BookinfoNamespace)
	})

	It("applies traffic shift policies to local subsets", func() {

		By("initially curling reviews should return both reviews-v1 and reviews-v2", func() {
			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "black"`))
			Eventually(curlReviews, "1m", "1s").ShouldNot(ContainSubstring(`"color": "black"`))
		})

		By("creating a TrafficPolicy with traffic shift to reviews-v2 should consistently shift traffic", func() {
			trafficShiftReviewsV2 := data.LocalTrafficShiftPolicy("bookinfo-policy", BookinfoNamespace, &v1.ClusterObjectRef{
				Name:        "reviews",
				Namespace:   BookinfoNamespace,
				ClusterName: mgmtClusterName,
			}, map[string]string{"version": "v2"}, 9080)

			err = manifest.AppendResources(trafficShiftReviewsV2)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			utils.AssertTrafficPolicyStatuses(dynamicClient, BookinfoNamespace)
			// check we consistently hit the v2 subset
			Consistently(curlReviews, "10s", "0.1s").Should(ContainSubstring(`"color": "black"`))
		})

		By("delete TrafficPolicy should remove traffic shift", func() {
			err = manifest.KubeDelete(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "black"`))
			Eventually(curlReviews, "1m", "1s").ShouldNot(ContainSubstring(`"color": "black"`))
		})
	})

	It("disables mTLS for traffic target", func() {
		var getReviewsDestinationRule = func() (*istionetworkingv1alpha3.DestinationRule, error) {
			env := e2e.GetEnv()
			destRuleClient := env.Management.DestinationRuleClient
			meta := metautils.TranslatedObjectMeta(
				&v1.ClusterObjectRef{
					Name:        "reviews",
					Namespace:   BookinfoNamespace,
					ClusterName: mgmtClusterName,
				},
				nil,
			)
			return destRuleClient.GetDestinationRule(ctx, client.ObjectKey{Name: meta.Name, Namespace: meta.Namespace})
		}

		By("initially ensure that DestinationRule exists for mgmt reviews traffic target", func() {
			Eventually(func() *istionetworkingv1alpha3.DestinationRule {
				destRule, err := getReviewsDestinationRule()
				if err != nil {
					return nil
				}
				return destRule
			}, "30s", "1s").ShouldNot(BeNil())
		})

		By("creating TrafficPolicy that overrides default mTLS settings for reviews traffic target", func() {
			tp := &v1alpha2.TrafficPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TrafficPolicy",
					APIVersion: v1alpha2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "mtls-disable",
					Namespace:   BookinfoNamespace,
					ClusterName: mgmtClusterName,
				},
				Spec: v1alpha2.TrafficPolicySpec{
					Mtls: &v1alpha2.TrafficPolicySpec_MTLS{
						Istio: &v1alpha2.TrafficPolicySpec_MTLS_Istio{
							TlsMode: v1alpha2.TrafficPolicySpec_MTLS_Istio_DISABLE,
						},
					},
				},
			}
			err = manifest.AppendResources(tp)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			utils.AssertTrafficPolicyStatuses(dynamicClient, BookinfoNamespace)

			// Check that DestinationRule for reviews no longer exists
			Eventually(func() bool {
				_, err := getReviewsDestinationRule()
				return errors.IsNotFound(err)
			}, "30s", "1s").Should(BeTrue())
		})

		By("first ensure that DestinationRule for mgmt reviews traffic target exists", func() {
			err = manifest.KubeDelete(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() *istionetworkingv1alpha3.DestinationRule {
				destRule, _ := getReviewsDestinationRule()
				return destRule
			}, "30s", "1s").ShouldNot(BeNil())
		})
	})
})
