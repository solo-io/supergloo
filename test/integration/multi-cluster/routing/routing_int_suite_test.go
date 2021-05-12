package main

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"istio.io/istio/pkg/test/echo/common/scheme"

	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	echo2 "github.com/solo-io/gloo-mesh/pkg/test/apps/echo"
	gloo_mesh "github.com/solo-io/gloo-mesh/pkg/test/apps/gloo-mesh"
	"github.com/solo-io/gloo-mesh/pkg/test/common"
	"istio.io/istio/pkg/test/framework/components/echo"
	"istio.io/istio/pkg/test/framework/components/environment/kube"
	"istio.io/istio/pkg/test/framework/config"
	"istio.io/istio/pkg/test/framework/resource"

	"istio.io/istio/pkg/test/framework"
	"istio.io/istio/pkg/test/framework/components/istio"
)

var (
	i             istio.Instance
	deploymentCtx context.DeploymentContext
)

func TestMain(m *testing.M) {
	if os.Getenv("RUN_INTEGRATION") == "" {
		fmt.Println("skipping E2E Integration tests")
		return
	}
	licenceKey := os.Getenv("GLOO_MESH_LICENSE_KEY")
	// get kube settings from command line
	config.Parse()
	kubeSettings, _ := kube.NewSettingsFromCommandLine()
	clusterKubeConfigs := make(map[string]string)

	// this is a hack to match the kube configs with the cluster names so we can match them when deploy happens
	for i, k := range kubeSettings.KubeConfig {
		clusterKubeConfigs[fmt.Sprintf("cluster-%d", i)] = k
	}

	framework.
		NewSuite(m).
		RequireMinClusters(2).
		Setup(istio.Setup(&i, common.IstioSetupFunc("gloo-mesh-istio.yaml"))).
		Setup(gloo_mesh.Deploy(&deploymentCtx, &gloo_mesh.Config{
			ClusterKubeConfigs:                  clusterKubeConfigs,
			DeployControlPlaneToManagementPlane: true,
		},
			licenceKey)).
		Setup(echo2.DeployEchos(&deploymentCtx)).
		Run()
}
func TestInMesh(t *testing.T) {
	framework.
		NewTest(t).
		Run(func(ctx framework.TestContext) {

			tgs := []common.TestGroup{
				{
					Name: "virtual-destination",
					Cases: []common.TestCase{
						{
							Name:        "same-cluster-http",
							Description: "Testing http routing in same cluster",
							Test:        testVirtualDestinationHTTP,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "same-cluster-http.yaml",
							Folder:      "gloo-mesh/in-mesh",
						},
						{
							Name:        "same-cluster-https",
							Description: "Testing https routing in same cluster",
							Test:        testVirtualDestinationHTTPS,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "same-cluster-https.yaml",
							Folder:      "gloo-mesh/in-mesh",
						},
						{
							Name:        "same-cluster-tcp",
							Description: "Testing tcp routing in same cluster",
							Test:        testVirtualDestinationTCP,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "same-cluster-tcp.yaml",
							Folder:      "gloo-mesh/in-mesh",
						},
					},
				},
			}
			for _, tg := range tgs {
				tg.Run(ctx, t, &deploymentCtx)
			}
		})

}

// Run the API tests
func TestRouting(t *testing.T) {
	framework.
		NewTest(t).
		Run(func(ctx framework.TestContext) {

			tgs := []common.TestGroup{
				{
					Name: "routing",
					Cases: []common.TestCase{
						{
							Name:        "prefix-1",
							Description: "HTTP/HTTPS prefix based routing",
							Test:        testPrefixMatch,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "prefix-1.yaml",
							Folder:      "traffic",
						},
						{
							Name:        "traffic-policy-timeout",
							Description: "Timeout of 1s using Gloo Mesh traffic policy",
							Test:        testTimeoutTrafficPolicy,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "timeout.yaml",
							Folder:      "gloo-mesh",
						},
					},
				},
			}
			for _, tg := range tgs {
				tg.Run(ctx, t, &deploymentCtx)
			}
		})
}

// testPrefixMatch makes a call from frontend to backend application
func testPrefixMatch(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend"))

	backendHost := "backend." + deploymentCtx.EchoContext.AppNamespace.Name() + ".svc.cluster.local"

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    "http",
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "/route1",
		Count:     1,
		Validator: echo.ExpectOK(),
	})

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    "http",
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "/route2",
		Count:     1,
		Validator: echo.ExpectOK(),
	})

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    "http",
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "/bad-route",
		Count:     1,
		Validator: echo.ExpectCode("404"),
	})
}

// testVirtualDestinationHTTP making http requests for a virtual destination
func testVirtualDestinationHTTP(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {

	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend"))
	backendHost := "http-backend.solo.io"

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "",
		Count:     5,
		Validator: echo.ExpectOK(),
	})

}

// testVirtualDestinationHTTPS making https requests for a virtual destination
func testVirtualDestinationHTTPS(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend"))
	backendHost := "http-backend.solo.io"

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "https",
			ServicePort: 8090,
			TLS:         true,
		},
		Scheme:    scheme.HTTPS,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "",
		Count:     5,
		Validator: echo.ExpectOK(),
		CaCert:    echo2.GetEchoCACert(),
	})
}

// testVirtualDestinationTCP making tcp requests for a virtual destination
func testVirtualDestinationTCP(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend"))
	backendHost := "http-backend.solo.io"

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "tcp",
			ServicePort: 9000,
		},
		Scheme:    scheme.TCP,
		Address:   backendHost,
		Count:     5,
		Validator: echo.ExpectOK(),
	})
}

// testTimeoutTrafficPolicy calling frontend applications to test timeout
func testTimeoutTrafficPolicy(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("backend"))
	frontendHost := "frontend." + deploymentCtx.EchoContext.AppNamespace.Name() + ".svc.cluster.local"

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   frontendHost,
		Method:    http.MethodGet,
		Path:      "",
		Count:     5,
		Validator: echo.ExpectOK(),
	})

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   frontendHost,
		Method:    http.MethodGet,
		Path:      "/?delay=4s",
		Count:     5,
		Validator: echo.ExpectCode("408"),
	})
}
