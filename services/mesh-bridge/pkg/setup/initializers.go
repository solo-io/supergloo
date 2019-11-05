package setup

import (
	"context"

	"github.com/solo-io/go-utils/configutils"
	"github.com/solo-io/go-utils/envutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/mesh-projects/pkg/mcutils"
	"github.com/solo-io/mesh-projects/services/internal/kube"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/setup/config"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/syncer"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/translator"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	sk_multicluster "github.com/solo-io/solo-kit/pkg/multicluster"
	"k8s.io/client-go/kubernetes"
)

type RestConfigHandlerRunFunc func() (<-chan error, error)

type RunOpts struct {
	OperatorLoop        eventloop.EventLoop
	MultiClusterRunFunc RestConfigHandlerRunFunc
	WatchOpts           clients.WatchOpts
	WatchNamespaces     []string
}

func NewLoopSet(operatorLoop eventloop.EventLoop, multiClusterRunFunc RestConfigHandlerRunFunc,
	opts clients.WatchOpts, watchNamespaces []string) RunOpts {
	return RunOpts{
		OperatorLoop:        operatorLoop,
		MultiClusterRunFunc: multiClusterRunFunc,
		WatchOpts:           opts,
		WatchNamespaces:     watchNamespaces,
	}
}

func MustInitializeMeshBridge(ctx context.Context) (*RunOpts, error) {
	restConfig := kube.MustGetKubeConfig(ctx)
	podNamespace := envutils.MustGetPodNamespace(ctx)
	kubernetesInterface := kube.MustGetClient(ctx, restConfig)
	configMapClient := configutils.NewConfigMapClient(kubernetesInterface)
	operatorConfig, err := config.GetOperatorConfig(ctx, configMapClient, podNamespace)
	if err != nil {
		return nil, err
	}
	initialSettings := config.GetInitialSettings(podNamespace, operatorConfig)
	watchAggregator := wrapper.NewWatchAggregator()
	cacheGetter, err := config.NewSharedCacheManager(ctx)
	if err != nil {
		return nil, err
	}
	clientForClusterHandler := mcutils.NewClientForClusterHandler(watchAggregator)
	clientSet := config.MustGetClientSet(ctx, cacheGetter, restConfig, clientForClusterHandler, initialSettings)
	networkBridgeEmitter := v1.NewNetworkBridgeEmitter(clientSet.MeshBridge())
	translatorTranslator := translator.NewMeshBridgeTranslator(clientSet)
	serviceEntryReconciler := v1alpha3.NewServiceEntryReconciler(clientSet.ServiceEntry())
	networkBridgeSyncer := syncer.NewMeshBridgeSyncer(clientSet, translatorTranslator, serviceEntryReconciler)
	eventLoop := v1.NewNetworkBridgeEventLoop(networkBridgeEmitter, networkBridgeSyncer)
	watchOpts := config.GetWatchOpts(ctx, initialSettings)
	v := config.GetWatchNamespaces(ctx, initialSettings)

	restConfigHandler := sk_multicluster.NewRestConfigHandler(
		sk_multicluster.NewKubeConfigWatcher(),
		clientSet.MultiClusterHandlers()...,
	)

	localKubeClient := kubernetes.NewForConfigOrDie(restConfig)
	localCache, err := cache.NewKubeCoreCache(ctx, localKubeClient)
	if err != nil {
		return nil, err
	}

	restConfigHandlerFunc := func() (<-chan error, error) {
		return restConfigHandler.Run(ctx, restConfig, localKubeClient, localCache)
	}

	loopSet := NewLoopSet(eventLoop, restConfigHandlerFunc, watchOpts, v)

	return &loopSet, nil
}
