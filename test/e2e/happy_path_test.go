package e2e

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1alpha1types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("HappyPath", func() {

	AfterEach(func() {
		testLabels := map[string]string{"test": "true"}
		opts := &client.DeleteAllOfOptions{}
		opts.LabelSelector = labels.SelectorFromSet(labels.Set(testLabels))
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
		ParseYaml(tpYaml, &tp)
		err := env.Management.TrafficPolicyClient.CreateTrafficPolicy(context.Background(), &tp)
		Expect(err).NotTo(HaveOccurred())
		// see that it was accepted

		Eventually(StatusOf(tp, env.Management), "1m", "1s").Should(Equal(v1alpha1types.Status_ACCEPTED))
	}

	It("should work with traffic policy to local (v2) reviews", func() {
		env := GetEnv()

		var tp v1alpha1.TrafficPolicy

		const tpYaml = `
apiVersion: networking.zephyr.solo.io/v1alpha1
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
  trafficShift:
    destinations:
    - destination:
        cluster: management-plane-cluster
        name: reviews
        namespace: default
      weight: 100
      subset:
        version: v2
`
		applyTrafficPolicy(tpYaml)
		Consistently(func() string {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
			defer cancel()
			out := env.Management.GetPod("default", "productpage").Curl(ctx, "http://reviews:9080/reviews/1", "-v")
			GinkgoWriter.Write([]byte(out))
			return out
		}, "1m", "1s").Should(ContainSubstring(`"color": "black"`))
	})

	// This test assumes that ci script only deploy v3 to remote cluster
	FIt("should work with traffic policy to remove (v3) reviews", func() {
		env := GetEnv()

		var tp v1alpha1.TrafficPolicy

		const tpYaml = `
apiVersion: networking.zephyr.solo.io/v1alpha1
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
  trafficShift:
    destinations:
    - destination:
        cluster: target-cluster
        name: reviews
        namespace: default
      weight: 100
`
		applyTrafficPolicy(tpYaml)

		Consistently(func() string {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
			defer cancel()
			out := env.Management.GetPod("default", "productpage").Curl(ctx, "http://reviews:9080/reviews/1", "-v")
			GinkgoWriter.Write([]byte(out))
			return out
		}, "1m", "1s").Should(ContainSubstring(`"color": "red"`))
	})
})
