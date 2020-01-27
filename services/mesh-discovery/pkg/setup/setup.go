package setup

import (
	"context"
	"os"

	"github.com/solo-io/mesh-projects/pkg/common/docker"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/mesh-projects/pkg/version"
	"github.com/solo-io/mesh-projects/services/internal/mcutils"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/consul"
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
)

// customCtx and customErrHandler are expected to be passed by tests
func Main(customCtx context.Context, errHandler func(error)) error {
	stats.ConditionallyStartStatsServer()

	writeNamespace := os.Getenv("POD_NAMESPACE")
	if writeNamespace == "" {
		writeNamespace = "sm-marketplace"
	}
	if errHandler == nil {
		errHandler = func(err error) {
			if err == nil {
				return
			}
			contextutils.LoggerFrom(customCtx).Errorf("error: %v", err)
		}
	}

	if err := runDiscoveryEventLoop(customCtx, writeNamespace, errHandler); err != nil {
		return err
	}

	<-customCtx.Done()
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
	meshIngressClient, _ := InitializeMeshIngressClient(ctx, sharedCacheGetter)
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
	meshIngressReconciler := v1.NewMeshIngressReconciler(meshIngressClient)

	linkerdDiscovery := linkerd.NewLinkerdDiscoverySyncer(
		writeNamespace,
		meshReconciler,
		meshIngressReconciler,
	)

	istioDiscovery := istio.NewIstioDiscoverySyncer(
		writeNamespace,
		meshReconciler,
		meshPolicyClient,
		crdClient,
		meshIngressReconciler,
	)

	dockerImageNameParser := docker.NewImageNameParser()
	connectInstallationFinder := consul.NewConsulConnectInstallationFinder(dockerImageNameParser)

	consulDiscovery := consul.NewConsulDiscoveryPlugin(
		writeNamespace,
		meshReconciler,
		meshIngressReconciler,
		connectInstallationFinder,
		dockerImageNameParser,
	)

	// appMeshClientBuilder := appmesh.NewAppMeshClientBuilder()
	// appmeshDiscovery := appmesh.NewAppmeshDiscoverySyncer(writeNamespace, meshReconciler, appMeshClientBuilder)

	emitter := v1.NewDiscoverySimpleEmitter(wrapper.ResourceWatch(watchAggregator, "", nil))
	eventLoop := v1.NewDiscoverySimpleEventLoop(emitter,
		linkerdDiscovery,
		istioDiscovery,
		consulDiscovery,
		// appmeshDiscovery,
	)

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
