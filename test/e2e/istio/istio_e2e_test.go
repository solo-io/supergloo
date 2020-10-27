package istio_test

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/solo-io/service-mesh-hub/test/extensions"

	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/test/e2e"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Shared test vars
var (
	BookinfoNamespace = "bookinfo"

	mgmtClusterName   = "mgmt-cluster"
	remoteClusterName = "remote-cluster"

	masterMesh = &v1.ObjectRef{
		Name:      "istiod-istio-system-mgmt-cluster",
		Namespace: "service-mesh-hub",
	}

	remoteMesh = &v1.ObjectRef{
		Name:      "istiod-istio-system-remote-cluster",
		Namespace: "service-mesh-hub",
	}

	// Initialize in BeforeSuite
	dynamicClient client.Client

	curlReviews = func() string {
		return curlFromProductpage("http://reviews:9080/reviews/1")
	}

	curlHelloServer = func() string {
		return curlFromProductpage(fmt.Sprintf("http://%v:%v/", extensions.HelloServerHostname, extensions.HelloServerPort))
	}

	curlRemoteReviews = func() string {
		return curlFromProductpage(fmt.Sprintf("http://reviews.%v.svc.%v.global:9080/reviews/1", BookinfoNamespace, remoteClusterName))
	}

	curlRatings = func() string {
		return curlFromProductpage("http://ratings:9080/ratings/1")
	}

	curlFromProductpage = func(url string) string {
		env := e2e.GetEnv()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
		defer cancel()
		out := env.Management.GetPod(ctx, BookinfoNamespace, "productpage").Curl(ctx, url, "-v")
		GinkgoWriter.Write([]byte(out))
		return out
	}

	curlGateway = func(hostname, path, body, method string) string {
		out, err := exec.Command("curl", "--connect-timeout", "1", "--max-time", "5", "-H", hostname, "http://localhost:32000"+path, "-v", "-d", body, "-X", method).CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		GinkgoWriter.Write(out)

		return string(out)
	}
)
