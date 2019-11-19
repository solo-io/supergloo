package setup

import (
	"context"

	"github.com/solo-io/go-utils/configutils"
	"github.com/solo-io/go-utils/envutils"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/mesh-projects/services/internal/config"
	"github.com/solo-io/mesh-projects/services/internal/kube"
	"github.com/solo-io/mesh-projects/services/internal/mcutils"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/syncer"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/translator"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	sk_multicluster "github.com/solo-io/solo-kit/pkg/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
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
	clientForClusterHandler := mcutils.NewClientForClusterHandler(watchAggregator)
	cacheManager, err := config.NewCacheManager(ctx)
	if err != nil {
		return nil, err
	}
	clientSet := config.MustGetMeshBridgeClientSet(ctx, cacheManager, restConfig, clientForClusterHandler, initialSettings)
	networkBridgeEmitter := v1.NewNetworkBridgeEmitter(clientSet.MeshBridge(), clientSet.Mesh(), clientSet.MeshIngress())
	translatorTranslator := translator.NewMeshBridgeTranslator(clientSet)
	networkBridgeSyncer := syncer.NewMeshBridgeSyncer(clientSet, translatorTranslator)
	eventLoop := v1.NewNetworkBridgeEventLoop(networkBridgeEmitter, networkBridgeSyncer)
	watchOpts := config.GetWatchOpts(ctx, initialSettings)
	v := config.GetWatchNamespaces(ctx, initialSettings)

	handlers := []handler.ClusterHandler{cacheManager}
	handlers = append(handlers, clientSet.MultiClusterHandlers()...)

	restConfigHandler := sk_multicluster.NewRestConfigHandler(
		sk_multicluster.NewKubeConfigWatcher(),
		handlers...,
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
