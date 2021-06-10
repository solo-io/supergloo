package main

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"istio.io/istio/pkg/test/echo/client"

	"istio.io/istio/pkg/test/framework/components/cluster"

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
func TestInMesh(t *testing.T) {
	flatNetworkingEnabled := ""
	if os.Getenv("FLAT_NETWORKING_ENABLED") != "true" {
		flatNetworkingEnabled = "flat networking not enabled, to enable set env FLAT_NETWORKING_ENABLED=true"
	}

	framework.
		NewTest(t).
		Run(func(ctx framework.TestContext) {

			tgs := []common.TestGroup{
				{
					Name: "virtual-destination",
					Cases: []common.TestCase{
						{
							Name:        "weighted-routing",
							Description: "Testing multi cluster weighted routing",
							Test:        testWeightedRouting,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "weighted-routing.yaml",
							Folder:      "gloo-mesh/in-mesh",
							Skip:        "Blocked https://github.com/solo-io/gloo-mesh-enterprise/issues/640 https://github.com/solo-io/gloo-mesh-enterprise/issues/589",
						},
						{
							Name:        "weighted-routing-single-cluster",
							Description: "Testing multi cluster weighted routing where only 1 cluster has apps",
							Test:        testWeightedRouting,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "weighted-routing-single-cluster.yaml",
							Folder:      "gloo-mesh/in-mesh",
							Skip:        "Blocked https://github.com/solo-io/gloo-mesh-enterprise/issues/640 https://github.com/solo-io/gloo-mesh-enterprise/issues/589",
						},
						{
							Name:        "same-cluster-http",
							Description: "Testing http routing in same cluster",
							Test:        testGlobalVirtualDestinationHTTP,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtual-destination-http.yaml",
							Folder:      "gloo-mesh/in-mesh",
						},
						{
							Name:        "same-cluster-https",
							Description: "Testing https routing in same cluster",
							Test:        testGlobalVirtualDestinationHTTPS,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtual-destination-https.yaml",
							Folder:      "gloo-mesh/in-mesh",
						},
						{
							Name:        "same-cluster-tcp",
							Description: "Testing tcp routing in same cluster",
							Test:        testGlobalVirtualDestinationTCP,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtual-destination-tcp.yaml",
							Folder:      "gloo-mesh/in-mesh",
						},
						{
							Name:        "different-cluster-http",
							Description: "Testing http routing from different cluster",
							Test:        testSingleClusterVirtualDestinationHTTP,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtual-destination-single-cluster-http.yaml",
							Folder:      "gloo-mesh/in-mesh",
						},
						{
							Name:        "different-cluster-https",
							Description: "Testing https routing in different cluster",
							Test:        testSingleClusterVirtualDestinationHTTPS,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtual-destination-single-cluster-https.yaml",
							Folder:      "gloo-mesh/in-mesh",
						},
						{
							Name:        "different-cluster-tcp",
							Description: "Testing tcp routing in different cluster",
							Test:        testSingleClusterVirtualDestinationTCP,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtual-destination-single-cluster-tcp.yaml",
							Folder:      "gloo-mesh/in-mesh",
						},
						{
							Name:        "failover-http",
							Description: "Testing http failover to different cluster",
							Test:        testFailoverHTTP,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtual-destination-http.yaml",
							Folder:      "gloo-mesh/in-mesh",
						},
						{
							Name:        "failover-https",
							Description: "Testing https failover to different cluster",
							Test:        testFailoverHTTPS,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtual-destination-https.yaml",
							Folder:      "gloo-mesh/in-mesh",
						},
						{
							Name:        "different-cluster-http-flat-network",
							Description: "Testing http routing from different cluster using a flat network",
							Test:        testSingleClusterVirtualDestinationHTTP,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtual-destination-single-cluster-http-flat-network.yaml",
							Folder:      "gloo-mesh/in-mesh",
							Skip:        flatNetworkingEnabled,
						},
						{
							Name:        "different-cluster-https-flat-network",
							Description: "Testing https routing in different cluster using a flat network",
							Test:        testSingleClusterVirtualDestinationHTTPS,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtual-destination-single-cluster-https-flat-network.yaml",
							Folder:      "gloo-mesh/in-mesh",
							Skip:        flatNetworkingEnabled,
						},
						{
							Name:        "different-cluster-tcp-flat-network",
							Description: "Testing tcp routing in different cluster using a flat network",
							Test:        testSingleClusterVirtualDestinationTCP,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtual-destination-single-cluster-tcp-flat-network.yaml",
							Folder:      "gloo-mesh/in-mesh",
							Skip:        flatNetworkingEnabled,
						},
					},
				},
			}
			for _, tg := range tgs {
				tg.Run(ctx, t, &deploymentCtx)
			}
		})

}

// testGlobalVirtualDestinationHTTP making http requests for a virtual destination
// because of locality priority routing, we should see routing to local cluster first always
func testGlobalVirtualDestinationHTTP(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
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
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(cluster.Name())),
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
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(cluster.Name())),
	})
}

// testGlobalVirtualDestinationHTTPS making https requests for a virtual destination
// because of locality priority routing, we should see routing to local cluster first always
func testGlobalVirtualDestinationHTTPS(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster := ctx.Clusters()[0]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backendHost := "https-backend.solo.io"

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "https",
			ServicePort: 9443,
			TLS:         true,
		},
		Scheme:    scheme.HTTPS,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(cluster.Name())),
		CaCert:    echo2.GetEchoCACert(),
	})

	cluster = ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in same cluster
	src = deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "https",
			ServicePort: 9443,
			TLS:         true,
		},
		Scheme:    scheme.HTTPS,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(cluster.Name())),
		CaCert:    echo2.GetEchoCACert(),
	})
}

// testGlobalVirtualDestinationTCP making tcp requests for a virtual destination
// because of locality priority routing, we should see routing to local cluster first always
func testGlobalVirtualDestinationTCP(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster := ctx.Clusters()[0]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backendHost := "tcp-backend.solo.io"

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "tcp",
			ServicePort: 9000,
		},
		Scheme:    scheme.TCP,
		Address:   backendHost,
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(cluster.Name())),
	})

	cluster = ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in same cluster
	src = deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "tcp",
			ServicePort: 9000,
		},
		Scheme:    scheme.TCP,
		Address:   backendHost,
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(cluster.Name())),
	})
}

// testSingleClusterVirtualDestinationHTTP making http requests for a virtual destination that only exists in 1 cluster
func testSingleClusterVirtualDestinationHTTP(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	clientCluster := ctx.Clusters()[0]
	expectedCluster := ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in different cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(clientCluster)))
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
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(expectedCluster.Name())),
	})
	// cluster 2 test
	clientCluster = ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in same cluster
	src = deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(clientCluster)))
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
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(expectedCluster.Name())),
	})
}

// testSingleClusterVirtualDestinationHTTPS making https requests for a virtual destination that only exists in 1 cluster
func testSingleClusterVirtualDestinationHTTPS(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	clientCluster := ctx.Clusters()[0]
	expectedCluster := ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in different cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(clientCluster)))
	backendHost := "https-backend.solo.io"

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "https",
			ServicePort: 9443,
			TLS:         true,
		},
		Scheme:    scheme.HTTPS,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(expectedCluster.Name())),
		CaCert:    echo2.GetEchoCACert(),
	})

	clientCluster = ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in same cluster
	src = deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(clientCluster)))

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "https",
			ServicePort: 9443,
			TLS:         true,
		},
		Scheme:    scheme.HTTPS,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(expectedCluster.Name())),
		CaCert:    echo2.GetEchoCACert(),
	})
}

// testSingleClusterVirtualDestinationTCP making tcp requests for a virtual destination that only exists in 1 cluster
func testSingleClusterVirtualDestinationTCP(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	clientCluster := ctx.Clusters()[0]
	expectedCluster := ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in different cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(clientCluster)))
	backendHost := "tcp-backend.solo.io"

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "tcp",
			ServicePort: 9000,
		},
		Scheme:    scheme.TCP,
		Address:   backendHost,
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(expectedCluster.Name())),
	})

	clientCluster = ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in same cluster
	src = deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(clientCluster)))
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "tcp",
			ServicePort: 9000,
		},
		Scheme:    scheme.TCP,
		Address:   backendHost,
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(expectedCluster.Name())),
	})
}

// testFailoverHTTP testing failover incase of error
// because of locality priority routing, we should see routing to local cluster first always
// TODO there is a bug where if someone creates a standalone pod in mesh and tries to make http calls. the calls are successful but do not respect regionality
// its like istio does not know what region they are in even though that comes from the node.
func testFailoverHTTP(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	westCluster := ctx.Clusters()[0]
	eastCluster := ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(westCluster)))
	backendHost := "http-backend.solo.io"

	// submit a 500 error to kick the west cluster
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "?codes=500:1", // returns 500
		Count:     15,
		Validator: echo.And(echo.ExpectCode("500"), echo.ExpectCluster(westCluster.Name())),
	})
	// should only get east cluster calls for 30s
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "?codes=200:1", // returns 200
		Count:     5,
		Validator: echo.ExpectReachedClusters(cluster.Clusters{eastCluster}),
	})
}

// testFailoverHTTPS testing failover incase of error
// because of locality priority routing, we should see routing to local cluster first always
// TODO there is a bug where if someone creates a standalone pod in mesh and tries to make http calls. the calls are successful but do not respect regionality
// its like istio does not know what region they are in even though that comes from the node.
func testFailoverHTTPS(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	westCluster := ctx.Clusters()[0]
	eastCluster := ctx.Clusters()[1]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(westCluster)))
	backendHost := "https-backend.solo.io"

	// submit a 500 error to kick the west cluster
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "https",
			ServicePort: 9443,
		},
		Scheme:    scheme.HTTPS,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "?codes=500:1", // returns 500
		Count:     15,
		Validator: echo.And(echo.ExpectCode("500"), echo.ExpectCluster(westCluster.Name())),
	})
	// should only get east cluster calls for 30s
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "https",
			ServicePort: 9443,
		},
		Scheme:    scheme.HTTPS,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "?codes=200:1", // returns 200
		Count:     5,
		Validator: echo.ExpectReachedClusters(cluster.Clusters{eastCluster}),
	})
}

// testWeightedRouting testing multi cluster weighted routing
func testWeightedRouting(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	westCluster := ctx.Clusters()[0]
	// frontend calling subset in mesh using virtual destination in same cluster and different clusters
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(westCluster)))
	backendHost := "http-subset.solo.io"

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   backendHost,
		Method:    http.MethodGet,
		Path:      "",
		Count:     100,
		Validator: echo.And(validateWeightedRouting(map[string]int{"v1": 90, "v2": 10}, 2)),
	})
}

func validateWeightedRouting(expected map[string]int, variance int) echo.Validator {
	return echo.ValidatorFunc(func(resp client.ParsedResponses, err error) error {
		// calculate the number of apps reached
		actual := make(map[string]int)
		for _, r := range resp {
			actual[r.Version] = actual[r.Version] + 1
		}
		if len(expected) != len(actual) {
			return fmt.Errorf("did not recieve the correct amount of subset actual got %d expectedf %d", len(actual), len(expected))
		}
		for v, e := range expected {
			a, exists := actual[v]
			if e != 0 && !exists {
				return fmt.Errorf("no requests for subset %s", v)
			}
			// make sure a is in the variance window
			if a < e-variance || a > e+variance {
				return fmt.Errorf("actaul amount of responses from subset %s was %d but expected to be between %d and %d ", v, a, e-variance, e+variance)
			}
		}

		return nil
	})

}
