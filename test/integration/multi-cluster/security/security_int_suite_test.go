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
	licenseKey := os.Getenv("GLOO_MESH_LICENSE_KEY")
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
			licenseKey)).
		Setup(echo2.DeployEchos(&deploymentCtx)).
		Run()
}
func TestSecurity(t *testing.T) {
	// flatNetworkingEnabled := ""
	// if os.Getenv("FLAT_NETWORKING_ENABLED") != "true" {
	// 	flatNetworkingEnabled = "flat networking not enabled, to enable set env FLAT_NETWORKING_ENABLED=true"
	// }

	framework.
		NewTest(t).
		Run(func(ctx framework.TestContext) {

			tgs := []common.TestGroup{
				{
					Name: "virtual-destination",
					Cases: []common.TestCase{
						{
							Name:        "global-access-policy",
							Description: "Test that no app has access to any other in mesh",
							Test:        testGlobalAccessPolicy,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "global-access-policy.yaml",
							Folder:      "gloo-mesh/security",
						},
						{
							Name:        "single-cluster-policy",
							Description: "Only allow frontend pods from cluster-1",
							Test:        testSingleClusterPolicy,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "single-cluster-access.yaml",
							Folder:      "gloo-mesh/security",
						},
					},
				},
			}
			for _, tg := range tgs {
				tg.Run(ctx, t, &deploymentCtx)
			}
		})
}

// testing access to apps
func testGlobalAccessPolicy(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {

	cluster := ctx.Clusters()[0]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
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
		Validator: echo.ExpectCode("403"),
	})
	// cluster 2 test
	cluster = ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in same cluster
	src = deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
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
		Validator: echo.ExpectCode("403"),
	})

}

func testSingleClusterPolicy(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {

	cluster := ctx.Clusters()[0]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backendHost := "http-backend.solo.io"

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
		Validator: echo.ExpectCode("200"),
	})
	// only allowed traffic on /info
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "/not-info",
		Count:     5,
		Validator: echo.ExpectCode("403"),
	})
	// cluster 2 test
	cluster = ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in same cluster
	src = deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
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
		Validator: echo.ExpectCode("403"),
	})

}
