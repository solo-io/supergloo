package tests

import (
	"context"
	"time"

	"github.com/solo-io/gloo-mesh/test/data"
	"github.com/solo-io/gloo-mesh/test/e2e"
	"github.com/solo-io/gloo-mesh/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TrafficPolicyTest() {
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
			Eventually(CurlReviews, "1m", "1s").Should(ContainSubstring(`"color": "black"`))
			Eventually(CurlReviews, "1m", "1s").ShouldNot(ContainSubstring(`"color": "black"`))
		})

		By("creating a TrafficPolicy with traffic shift to reviews-v2 should consistently shift traffic", func() {
			trafficShiftReviewsV2 := data.LocalTrafficShiftPolicy("bookinfo-policy", BookinfoNamespace, &v1.ClusterObjectRef{
				Name:        "reviews",
				Namespace:   BookinfoNamespace,
				ClusterName: MgmtClusterName,
			}, map[string]string{"version": "v2"}, 9080)

			err = manifest.AppendResources(trafficShiftReviewsV2)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			utils.AssertTrafficPolicyStatuses(ctx, e2e.GetEnv().Management.TrafficPolicyClient, BookinfoNamespace)

			// insert a sleep because we can't effectively use an Eventually here to ensure config has propagated to envoy
			time.Sleep(time.Second * 5)

			// check we can eventually (consistently) hit the v2 subset
			Eventually(CurlReviews, "30s", "0.1s").Should(ContainSubstring(`"color": "black"`))
			Consistently(CurlReviews, "10s", "0.1s").Should(ContainSubstring(`"color": "black"`))
		})

		By("deleting TrafficPolicy should remove traffic shift", func() {
			err = manifest.KubeDelete(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			Eventually(CurlReviews, "1m", "1s").Should(ContainSubstring(`"color": "black"`))
			Eventually(CurlReviews, "1m", "1s").ShouldNot(ContainSubstring(`"color": "black"`))
		})
	})
}
