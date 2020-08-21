package osm_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/service-mesh-hub/test/e2e"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
)

// Shared test vars
var (
	BookThiefNamespace = "bookthief"

	masterClusterName = "master-cluster"

	mainMesh = &v1.ObjectRef{
		Name:      "istiod-istio-system-master-cluster",
		Namespace: "service-mesh-hub",
	}

	curlFromProductpage = func(url string) string {
		env := e2e.GetEnv()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
		defer cancel()
		out := env.Management.GetPod(ctx, BookThiefNamespace, BookThiefNamespace).Curl(ctx, url, "-v")
		GinkgoWriter.Write([]byte(out))
		return out
	}
)

var _ = Describe("OsmE2e", func() {
	//
	// var (
	// 	ctx context.Context
	// )
	//
	// BeforeEach(func() {
	// 	ctx = context.Background()
	// })

	It("is good", func() {

	})

})
