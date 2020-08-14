package e2e_test

import (
	"github.com/solo-io/service-mesh-hub/test/data"
	"github.com/solo-io/service-mesh-hub/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Federation", func() {
	var (
		err      error
		manifest utils.Manifest
	)

	AfterEach(func() {
		manifest.Cleanup(BookinfoNamespace)
	})

	It("enables communication across clusters using global dns names", func() {
		Expect(err).NotTo(HaveOccurred())
		manifest, err = utils.NewManifest("federation-trafficpolicies.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("with federation enabled, TrafficShifts can be used for subsets across meshes ", func() {
			// create cross cluster traffic shift
			trafficShiftReviewsV3 := data.RemoteTrafficShiftPolicy("bookinfo-policy", BookinfoNamespace, &v1.ClusterObjectRef{
				Name:        "reviews",
				Namespace:   BookinfoNamespace,
				ClusterName: masterClusterName,
			}, remoteClusterName, map[string]string{"version": "v3"}, 9080)

			err = manifest.AppendResources(trafficShiftReviewsV3)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			utils.AssertTrafficPolicyStatuses(dynamicClient, BookinfoNamespace)

			// check we can eventually hit the v3 subset
			Eventually(curlReviews, "30s", "1s").Should(ContainSubstring(`"color": "red"`))
		})
	})
})
