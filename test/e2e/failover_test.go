package e2e

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"

	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Failover Test", func() {
	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	AfterEach(func() {
		testLabels := map[string]string{"test": "true"}
		opts := &client.DeleteAllOfOptions{}
		opts.LabelSelector = labels.SelectorFromSet(testLabels)
		opts.Namespace = "service-mesh-hub"
		err := env.Management.TrafficPolicyClient.DeleteAllOfTrafficPolicy(ctx, opts)
		Expect(err).ToNot(HaveOccurred())
		err = env.Management.FailoverServiceClient.DeleteAllOfFailoverService(ctx, opts)
		Expect(err).ToNot(HaveOccurred())
	})

	applyFailoverService := func(env Env, failoverYaml string) {
		var failoverService smh_networking.FailoverService
		ParseYaml(failoverYaml, &failoverService)
		err := env.Management.FailoverServiceClient.CreateFailoverService(ctx, &failoverService)
		Expect(err).NotTo(HaveOccurred())
		// see that it was accepted
		Eventually(StatusOf(failoverService, env.Management), "1m", "1s").Should(Equal(smh_core_types.Status_ACCEPTED))
	}

	It("should get env", func() {
		env := GetEnv()
		Eventually(func() string {
			ctx, cancel := context.WithTimeout(ctx, time.Minute/3)
			defer cancel()
			out := env.Management.GetPod("default", "productpage").Curl(ctx, "http://reviews:9080/reviews/1")
			GinkgoWriter.Write([]byte(out))
			return out
		}, "1m", "1s").Should(ContainSubstring("The slapstick humour is refreshing!"))
	})

	It("should create failover config for the reviews service", func() {
		env := GetEnv()
		tpYaml := `
apiVersion: networking.smh.solo.io/v1alpha1
kind: TrafficPolicy
metadata:
  namespace: service-mesh-hub
  name: reviews-outlier-detection
  labels:
    test: true
spec:
  destinationSelector:
    serviceRefs:
      services:
      - name: reviews
        namespace: default
        cluster: management-plane-cluster
      - name: reviews
        namespace: default
        cluster: target-cluster
  outlierDetection:
    consecutiveErrors: 1
`
		failoverYaml := `
apiVersion: networking.smh.solo.io/v1alpha1
kind: FailoverService
metadata:
  name: reviews-failover
  namespace: service-mesh-hub
  labels:
    test: true
spec:
  targetService:
    name: reviews
    namespace: default
    cluster: management-plane-cluster
  failoverServices:
    - name: reviews
      namespace: default
      cluster: target-cluster
`
		ApplyTrafficPolicy(env, tpYaml)
		applyFailoverService(env, failoverYaml)

		// Make it failover to remote cluster with reviews-v3
		env.Management.DisableAppContainer(ctx, "default", "reviews-v1", "reviews")
		env.Management.DisableAppContainer(ctx, "default", "reviews-v2", "reviews")
		env.Management.WaitForRollout(ctx, "default", "reviews-v1")
		env.Management.WaitForRollout(ctx, "default", "reviews-v2")

		curlReviewsFunc := CurlReviews(env)
		// first check that we have a response to reduce flakiness
		Eventually(curlReviewsFunc, "5m", "2s").Should(ContainSubstring(`"color": "red"`))
		// now check that it is consistent 10 times in a row
		Eventually(func() bool {
			for i := 0; i < 5; i++ {
				if !strings.Contains(curlReviewsFunc(), `"color": "red"`) {
					return false
				}
				time.Sleep(2 * time.Second)
			}
			return true
		}, "5m", "2s").Should(BeTrue())
	})

	It("should re-enable management plane reviews deployments", func() {
		env.Management.EnableAppContainer(ctx, "default", "reviews-v1")
		env.Management.EnableAppContainer(ctx, "default", "reviews-v2")
		env.Management.WaitForRollout(ctx, "default", "reviews-v1")
		env.Management.WaitForRollout(ctx, "default", "reviews-v2")
		Eventually(CurlReviews(env), "2m", "2s").Should(ContainSubstring(`"color": "black"`))
	})
})
