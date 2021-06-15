package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"istio.io/istio/pkg/test/echo/common/scheme"
	"istio.io/istio/pkg/test/framework/components/echo"
	"istio.io/istio/pkg/test/framework/resource"

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
						{
							Name:        "add-request-header",
							Description: "Add request header when calling backend service",
							Test:        testAddRequestHeader,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "add-request-header.yaml",
							Folder:      "gloo-mesh/traffic-policy",
							Skip:        "https://github.com/solo-io/gloo-mesh-enterprise/issues/688",
						},
						{
							Name:        "request-path-matcher",
							Description: "Checking request matching on a prefix route",
							Test:        testRequestPathMatcher,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "request-prefix-matcher.yaml",
							Folder:      "gloo-mesh/traffic-policy",
						},
						{
							Name:        "request-header-matcher",
							Description: "Checking request matching on a header",
							Test:        testRequestHeaderMatcher,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "request-header-matcher.yaml",
							Folder:      "gloo-mesh/traffic-policy",
						},
						{
							Name:        "request-prefix-header-matcher-and",
							Description: "Checking request matching on a header and prefix",
							Test:        testRequestPrefixAndHeaderMatcher,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "request-prefix-header-matcher-and.yaml",
							Folder:      "gloo-mesh/traffic-policy",
						},
						{
							Name:        "request-prefix-header-matcher-or",
							Description: "Checking request matching on a header or prefix",
							Test:        testRequestPrefixOrHeaderMatcher,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "request-prefix-header-matcher-or.yaml",
							Folder:      "gloo-mesh/traffic-policy",
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
		Path:      "/info",
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
		Path:      "/info?delay=3s",
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
		Path:      "/info",
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
		Path:      "/info?delay=3s",
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
func testAddRequestHeader(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
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
		Path:      "/info",
		Count:     1,
		Validator: echo.ExpectKey("who", "hoo"),
	})

}

func testRequestPathMatcher(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster := ctx.Clusters()[0]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backendHost := fmt.Sprintf("backend.%s.svc.cluster.local", deploymentCtx.EchoContext.AppNamespace.Name())

	// happy requests with correct path
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "/two",
		Count:     1,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(ctx.Clusters()[1].Name())),
	})

	// happy requests non /two should 404
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "/one",
		Count:     1,
		Validator: echo.ExpectCode("404"),
	})

}

func testRequestHeaderMatcher(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster := ctx.Clusters()[0]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backendHost := fmt.Sprintf("backend.%s.svc.cluster.local", deploymentCtx.EchoContext.AppNamespace.Name())
	header := http.Header{}
	header.Add("color-header", "blue")
	// happy requests with header color-header: blue
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Headers:   header,
		Path:      "/info",
		Count:     1,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(ctx.Clusters()[1].Name())),
	})
	header = http.Header{}
	header.Add("color-header", "red")
	// different header value
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Headers:   header,
		Path:      "/info",
		Count:     1,
		Validator: echo.ExpectCode("404"),
	})
	// no header
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "/info",
		Count:     1,
		Validator: echo.ExpectCode("404"),
	})

}

func testRequestPrefixAndHeaderMatcher(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster := ctx.Clusters()[0]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backendHost := fmt.Sprintf("backend.%s.svc.cluster.local", deploymentCtx.EchoContext.AppNamespace.Name())
	header := http.Header{}
	header.Add("color-header", "blue")
	// requests with header color-header: blue but wrong path
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Headers:   header,
		Path:      "/info",
		Count:     1,
		Validator: echo.ExpectCode("404"),
	})

	// right path but wrong header
	header = http.Header{}
	header.Add("color-header", "red")
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Headers:   header,
		Path:      "/two",
		Count:     1,
		Validator: echo.ExpectCode("404"),
	})
	// right path and right header
	header = http.Header{}
	header.Add("color-header", "blue")
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Headers:   header,
		Method:    http.MethodGet,
		Path:      "/two",
		Count:     1,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(ctx.Clusters()[1].Name())),
	})

}

func testRequestPrefixOrHeaderMatcher(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster := ctx.Clusters()[0]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backendHost := fmt.Sprintf("backend.%s.svc.cluster.local", deploymentCtx.EchoContext.AppNamespace.Name())
	header := http.Header{}
	header.Add("color-header", "blue")
	// requests with header color-header: blue but wrong path. still should match
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Headers:   header,
		Path:      "/info",
		Count:     1,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(ctx.Clusters()[1].Name())),
	})

	// right path but wrong header. still should match
	header = http.Header{}
	header.Add("color-header", "red")
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Headers:   header,
		Path:      "/two",
		Count:     1,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(ctx.Clusters()[1].Name())),
	})
	// right path and right header
	header = http.Header{}
	header.Add("color-header", "blue")
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Headers:   header,
		Method:    http.MethodGet,
		Path:      "/two",
		Count:     1,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(ctx.Clusters()[1].Name())),
	})

	// neither path nor header
	header = http.Header{}
	header.Add("color-header", "red")
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Headers:   header,
		Method:    http.MethodGet,
		Path:      "/info",
		Count:     1,
		Validator: echo.ExpectCode("404"),
	})
}
