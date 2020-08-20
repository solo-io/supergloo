package e2e_test

import (
	"github.com/solo-io/service-mesh-hub/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

	"github.com/solo-io/service-mesh-hub/test/data"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TrafficPolicy", func() {
	var (
		err      error
		manifest utils.Manifest
	)

	AfterEach(func() {
		manifest.Cleanup(BookinfoNamespace)
	})

	It("applies traffic shift policies to local subsets", func() {
		manifest, err = utils.NewManifest("bookinfo-policies.yaml")
		Expect(err).NotTo(HaveOccurred())

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
})
