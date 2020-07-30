package e2e_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/service-mesh-hub/test/e2e"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Shared test vars
var (
	policyName     = "bookinfo-policy"
	appNamespace   = "bookinfo"
	policyManifest = "test/e2e/bookinfo-policies.yaml"

	dynamicClient client.Client

	curlReviews = func() string {
		env := e2e.GetEnv()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
		defer cancel()
		out := env.Management.GetPod(appNamespace, "productpage").Curl(ctx, "http://reviews:9080/reviews/1", "-v")
		GinkgoWriter.Write([]byte(out))
		return out
	}
)
