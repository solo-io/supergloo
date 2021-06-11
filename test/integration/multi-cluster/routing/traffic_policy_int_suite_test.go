package main

import (
	"fmt"
	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"istio.io/istio/pkg/test/echo/common/scheme"
	"istio.io/istio/pkg/test/framework/components/echo"
	"istio.io/istio/pkg/test/framework/resource"
	"net/http"
	"testing"

	"github.com/solo-io/gloo-mesh/pkg/test/common"
	"istio.io/istio/pkg/test/framework"
)

func TestTrafficPolicies(t *testing.T) {
	framework.
		NewTest(t).
		Run(func(ctx framework.TestContext) {

			tgs := []common.TestGroup{
				{
					Name: "traffic-policies",
					Cases: []common.TestCase{
						{
							Name:        "request-timeout",
							Description: "Test 1s timeout on backend in cluster-1",
							Test:        testRequestTimeout,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "request-timeout.yaml",
							Folder:      "gloo-mesh/traffic-policy",
							Skip:        "https://github.com/solo-io/gloo-mesh-enterprise/issues/687",
						},
						{
							Name:        "request-timeout-multi-cluster",
							Description: "Test 1s timeout on backend in all clusters",
							Test:        testRequestTimeoutMultiCluster,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "request-timeout-multi-cluster.yaml",
							Folder:      "gloo-mesh/traffic-policy",
							Skip:        "https://github.com/solo-io/gloo-mesh-enterprise/issues/688",
						},
					},
				},
			}
			for _, tg := range tgs {
				tg.Run(ctx, t, &deploymentCtx)
			}
		})

}
func testRequestTimeout(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster := ctx.Clusters()[0]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backendHost := fmt.Sprintf("backend.%s.svc.cluster.local", deploymentCtx.EchoContext.AppNamespace.Name())

	// happy requests
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "info",
		Count:     5,
		Validator: echo.ExpectOK(),
	})

	// fail due to request timeout
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "info?delay=3s",
		Count:     5,
		Validator: echo.ExpectError(),
	})

	// cluster 2 test
	cluster = ctx.Clusters()[1]
	// frontend calling backend in cluster 2 with no request timeout
	src = deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "info?delay=3s",
		Count:     5,
		Validator: echo.ExpectOK(),
	})
}

func testRequestTimeoutMultiCluster(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster := ctx.Clusters()[0]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backendHost := fmt.Sprintf("backend.%s.svc.cluster.local", deploymentCtx.EchoContext.AppNamespace.Name())

	// happy requests
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "info",
		Count:     5,
		Validator: echo.ExpectOK(),
	})

	// fail due to request timeout
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "info?delay=3s",
		Count:     5,
		Validator: echo.ExpectError(),
	})

	// cluster 2 test
	cluster = ctx.Clusters()[1]
	// frontend calling backend in cluster 2 with no request timeout
	src = deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "/info?delay=3s",
		Count:     5,
		Validator: echo.ExpectError(),
	})
}
