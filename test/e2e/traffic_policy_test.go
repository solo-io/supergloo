package e2e_test

import (
	"context"
	"time"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

	"github.com/solo-io/service-mesh-hub/test/data"

	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Must run `make generated-code` before running this test
var _ = Describe("TrafficPolicy E2e", func() {
	var (
		err      error
		manifest utils.Manifest
	)

	AfterEach(func() {
		manifest.Cleanup(BookinfoNamespace)
	})

	It("applies TrafficShift policies to local subsets", func() {
		manifest, err = utils.NewManifest("bookinfo-policies.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("initially curling reviews should return both reviews-v2 and reviews-v3", func() {
			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "black"`))
			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "red"`))
		})

		By("creating a TrafficPolicy with traffic shift to reviews-v2 should consistently shift traffic", func() {
			trafficShiftReviewsV2 := data.TrafficShiftPolicy(policyName, BookinfoNamespace, &v1.ClusterObjectRef{
				Name:        "reviews",
				Namespace:   BookinfoNamespace,
				ClusterName: masterClusterName,
			}, map[string]string{"version": "v2"}, 9080)

			err = manifest.AppendResource(trafficShiftReviewsV2)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			assertTrafficPolicyStatuses()
			// check we consistently hit the v2 subset
			Consistently(curlReviews, "10s", "0.1s").Should(ContainSubstring(`"color": "black"`))
		})

		By("delete TrafficPolicy should remove traffic shift", func() {
			err = manifest.KubeDelete(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "black"`))
			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "red"`))
		})
	})
})

//func assertCrdStatuses() {
//
//	err := testutils.Kubectl("apply", "-n="+BookinfoNamespace, "-f="+policyManifest)
//	Expect(err).NotTo(HaveOccurred())
//
//	assertTrafficPolicyStatuses()
//}

func assertTrafficPolicyStatuses() {
	ctx := context.Background()
	trafficPolicy := v1alpha2.NewTrafficPolicyClient(dynamicClient)

	EventuallyWithOffset(1, func() bool {
		list, err := trafficPolicy.ListTrafficPolicy(ctx, client.InNamespace(BookinfoNamespace))
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		ExpectWithOffset(1, list.Items).To(HaveLen(1))
		for _, policy := range list.Items {
			if policy.Status.ObservedGeneration == 0 {
				return false
			}
		}
		return true
	}, time.Second*20).Should(BeTrue())
}
