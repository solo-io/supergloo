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
	policyName        = "bookinfo-policy"
	BookinfoNamespace = "bookinfo"

	masterClusterName = "master-cluster"
	remoteClusterName = "remote-cluster"

	// Initialize in BeforeSuite
	dynamicClient client.Client

	curlReviews = func() string {
		return curlFromProductpage("http://reviews:9080/reviews/1")
	}

	curlDetails = func() string {
		return curlFromProductpage("http://details:9080/details/1")
	}

	curlFromProductpage = func(url string) string {
		env := e2e.GetEnv()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
		defer cancel()
		out := env.Management.GetPod(BookinfoNamespace, "productpage").Curl(ctx, url, "-v")
		GinkgoWriter.Write([]byte(out))
		return out
	}
)
