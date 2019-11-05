package setup

import (
	"context"
	"os"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/mesh-projects/pkg/mcutils"
	"github.com/solo-io/mesh-projects/pkg/version"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/istio"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/linkerd"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/customresourcedefinition"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	sk_multicluster "github.com/solo-io/solo-kit/pkg/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/kubernetes"

	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/clientset"
)

// customCtx and customErrHandler are expected to be passed by tests
func Main(customCtx context.Context, errHandler func(error)) error {
	if os.Getenv("START_STATS_SERVER") != "" {
		stats.StartStatsServer()
	}

	writeNamespace := os.Getenv("POD_NAMESPACE")
	if writeNamespace == "" {
		writeNamespace = "sm-marketplace"
	}

	rootCtx := createRootContext(customCtx)

	if errHandler == nil {
		errHandler = func(err error) {
			if err == nil {
				return
			}
			contextutils.LoggerFrom(rootCtx).Errorf("error: %v", err)
		}
	}

	if err := runDiscoveryEventLoop(rootCtx, writeNamespace, errHandler); err != nil {
		return err
	}

	<-rootCtx.Done()
	return nil
}

func createRootContext(customCtx context.Context) context.Context {
	rootCtx := customCtx
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	rootCtx = contextutils.WithLogger(rootCtx, "meshdiscovery")
	loggingContext := []interface{}{"version", version.Version}
	rootCtx = contextutils.WithLoggerValues(rootCtx, loggingContext...)
	return rootCtx
}

func runDiscoveryEventLoopDeprecated(ctx context.Context, writeNamespace string, cs *clientset.Clientset, errHandler func(error)) error {

	meshReconciler := v1.NewMeshReconciler(cs.Discovery.Mesh)
	//
	// istioDiscovery := istio.NewIstioDiscoverySyncer(
	//	writeNamespace,
	//	meshReconciler,
	//	istioClients.MeshPolicies,
	//	cs.ApiExtensions.ApiextensionsV1beta1().CustomResourceDefinitions(),
	//	cs.Kube.BatchV1(),
	// )

	linkerdDiscovery := linkerd.NewLinkerdDiscoverySyncer(
		writeNamespace,
		meshReconciler,
	)

	// appmeshDiscovery := appmesh.NewAppmeshDiscoverySyncer(
	//	writeNamespace,
	//	meshReconciler,
	//	appmeshconfig.NewAppMeshClientBuilder(cs.Input.Secret),
	//	cs.Input.Secret,
	// )

	emitter := v1.NewDiscoverySimpleEmitter(wrapper.AggregatedWatchFromClients(
		wrapper.ClientWatchOpts{BaseClient: cs.Input.Deployment.BaseClient()},
		wrapper.ClientWatchOpts{BaseClient: cs.Input.Upstream.BaseClient()},
		wrapper.ClientWatchOpts{BaseClient: cs.Input.Pod.BaseClient()},
		// wrapper.ClientWatchOpts{BaseClient: cs.Input.TlsSecret.BaseClient()},
	))
	eventLoop := v1.NewDiscoverySimpleEventLoop(emitter,
		// istioDiscovery,
		linkerdDiscovery,
		// appmeshDiscovery,
	)

	errs, err := eventLoop.Run(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-errs:
				errHandler(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func runDiscoveryEventLoop(ctx context.Context, writeNamespace string, errHandler func(error)) error {
	sharedCacheGetter, err := clustercache.NewCacheManager(ctx, kube.NewKubeSharedCacheForConfig)
	if err != nil {
		return err
	}
	deploymentCacheGetter, err := clustercache.NewCacheManager(ctx, cache.NewDeploymentCacheFromConfig)
	if err != nil {
		return err
	}
	crdCacheGetter, err := clustercache.NewCacheManager(ctx, customresourcedefinition.NewCrdCacheForConfig)
	if err != nil {
		return err
	}
	coreCacheGetter, err := clustercache.NewCacheManager(ctx, cache.NewCoreCacheForConfig)
	if err != nil {
		return err
	}

	watchAggregator := wrapper.NewWatchAggregator()
	watchHandler := mcutils.NewAggregatedWatchClusterClientHandler(watchAggregator)

	// TODO does this Should Not be multicluster
	meshClient, _ := InitializeMeshClient(ctx, sharedCacheGetter)
	_, upstreamClientHandler := InitializeUpstreamClient(ctx, sharedCacheGetter, watchHandler)
	_, deploymentClientHandler := InitializeDeploymentClient(ctx, deploymentCacheGetter, watchHandler)
	_, podClientHandler := InitializePodClient(ctx, coreCacheGetter, watchHandler)
	crdClient, crdClientHandler := InitializeCustomResourceDefinitionClient(ctx, crdCacheGetter)
	meshPolicyClient, meshPolicyClientHandler := InitializeMeshPolicyClient(ctx, sharedCacheGetter)

	localRestConfig, err := kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return err
	}
	localKubeClient, err := kubernetes.NewForConfig(localRestConfig)
	if err != nil {
		return err
	}
	localCache, err := cache.NewKubeCoreCache(ctx, localKubeClient)
	if err != nil {
		return err
	}
	restConfigHandler := sk_multicluster.NewRestConfigHandler(
		sk_multicluster.NewKubeConfigWatcher(),
		sharedCacheGetter,
		deploymentCacheGetter,
		coreCacheGetter,
		upstreamClientHandler,
		deploymentClientHandler,
		podClientHandler,
		crdClientHandler,
		meshPolicyClientHandler,
		watchHandler,
	)

	errs, err := restConfigHandler.Run(ctx, localRestConfig, localKubeClient, localCache)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-errs:
				errHandler(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	meshReconciler := v1.NewMeshReconciler(meshClient)

	linkerdDiscovery := linkerd.NewLinkerdDiscoverySyncer(
		writeNamespace,
		meshReconciler,
	)

	istioDiscovery := istio.NewIstioDiscoverySyncer(
		writeNamespace,
		meshReconciler,
		meshPolicyClient,
		crdClient,
		nil,
	)

	emitter := v1.NewDiscoverySimpleEmitter(wrapper.ResourceWatch(watchAggregator, "", nil))
	// eventLoop := v1.NewDiscoverySimpleEventLoop(emitter, linkerdDiscovery)
	eventLoop := v1.NewDiscoverySimpleEventLoop(emitter, linkerdDiscovery, istioDiscovery)

	errs, err = eventLoop.Run(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-errs:
				errHandler(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
