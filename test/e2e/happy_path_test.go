package e2e

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("HappyPath", func() {

	AfterEach(func() {
		testLabels := map[string]string{"test": "true"}
		opts := &client.DeleteAllOfOptions{}
		opts.LabelSelector = labels.SelectorFromSet(testLabels)
		opts.Namespace = "service-mesh-hub"
		env.Management.TrafficPolicyClient.DeleteAllOfTrafficPolicy(context.Background(), opts)
	})

	It("should get env", func() {
		env := GetEnv()
		Eventually(func() string {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
			defer cancel()
			out := env.Management.GetPod("default", "productpage").Curl(ctx, "http://reviews:9080/reviews/1", "-v")
			GinkgoWriter.Write([]byte(out))
			return out
		}, "1m", "1s").Should(ContainSubstring("The slapstick humour is refreshing!"))
	})

	applyTrafficPolicy := func(tpYaml string) {
		var tp smh_networking.TrafficPolicy
		ParseYaml(tpYaml, &tp)
		err := env.Management.TrafficPolicyClient.CreateTrafficPolicy(context.Background(), &tp)
		Expect(err).NotTo(HaveOccurred())
		// see that it was accepted

		Eventually(StatusOf(tp, env.Management), "1m", "1s").Should(Equal(smh_core_types.Status_ACCEPTED))
	}

	curlReviews := func() string {
		env := GetEnv()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
		defer cancel()
		out := env.Management.GetPod("default", "productpage").Curl(ctx, "http://reviews:9080/reviews/1", "-v")
		GinkgoWriter.Write([]byte(out))
		return out
	}

	FIt("should work with traffic policy to local (v2) reviews", func() {
		const tpYaml = `
apiVersion: networking.smh.solo.io/v1alpha1
kind: TrafficPolicy
metadata:
  namespace: service-mesh-hub
  name: simple
  labels:
    test: true
spec:
  sourceSelector:
    labels:
      app: productpage
  destinationSelector:
    serviceRefs:
      services:
        name: ratings-default-management-plane-cluster
        namespace: default
  trafficShift:
    destinations:
    - destination:
        cluster: management-plane-cluster
        name: reviews
        namespace: default
        cluster: management-plane-cluster
      weight: 100
      subset:
        version: v2
`
		applyTrafficPolicy(tpYaml)
		// first check that we have a response to reduce flakiness
		Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "black"`))
		// now check that it is consistent 10 times in a row
		for i := 0; i < 10; i++ {
			Expect(curlReviews()).Should(ContainSubstring(`"color": "black"`))
		}
	})

	// This test assumes that ci script only deploy v3 to remote cluster
	It("should work with traffic policy to remote (v3) reviews", func() {
		const tpYaml = `
apiVersion: networking.smh.solo.io/v1alpha1
kind: TrafficPolicy
metadata:
  namespace: service-mesh-hub
  name: simplev3
  labels:
    test: true
spec:
  sourceSelector:
    labels:
      app: productpage
  destinationSelector:
    serviceRefs:
      services:
        name: ratings
        namespace: default
        cluster: management-plane-cluster
  trafficShift:
    destinations:
    - destination:
        cluster: target-cluster
        name: reviews
        namespace: default
      weight: 100
`
		applyTrafficPolicy(tpYaml)

		// first check that we have a response to reduce flakiness
		Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "red"`))
		// now check that it is consistent 10 times in a row
		for i := 0; i < 10; i++ {
			Expect(curlReviews()).Should(ContainSubstring(`"color": "red"`))
		}
	})
})
