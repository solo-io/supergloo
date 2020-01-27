package config

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/clientfactory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"go.uber.org/zap"
	client_go "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CacheManager interface {
	SharedCache(ctx context.Context) clustercache.CacheGetter
	CoreCache(ctx context.Context) clustercache.CacheGetter
	handler.ClusterHandler
}

type cacheManager struct {
	sharedCache clustercache.CacheManager
	coreCache   clustercache.CacheManager
}

func NewCacheManager(ctx context.Context) (*cacheManager, error) {
	sharedCacheGetter, err := clustercache.NewCacheManager(ctx, kube.NewKubeSharedCacheForConfig)
	if err != nil {
		return nil, err
	}
	coreCacheGetter, err := clustercache.NewCacheManager(ctx, cache.NewCoreCacheForConfig)
	if err != nil {
		return nil, err
	}
	return &cacheManager{
		sharedCache: sharedCacheGetter,
		coreCache:   coreCacheGetter,
	}, nil
}

func (c *cacheManager) SharedCache(ctx context.Context) clustercache.CacheGetter {
	return c.sharedCache
}

func (c *cacheManager) CoreCache(ctx context.Context) clustercache.CacheGetter {
	return c.coreCache
}

func (c *cacheManager) ClusterAdded(cluster string, restConfig *rest.Config) {
	c.sharedCache.ClusterAdded(cluster, restConfig)
	c.coreCache.ClusterAdded(cluster, restConfig)
}

func (c *cacheManager) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.sharedCache.ClusterRemoved(cluster, restConfig)
	c.coreCache.ClusterRemoved(cluster, restConfig)
}

type MultiClusterGetters struct {
	upstream handler.ClusterHandler
}

func GetInitialSettings(installNamespace string, settings *OperatorConfig) *InitialSettings {
	return &InitialSettings{
		InstallNamespace: "",
		RefreshRate:      settings.RefreshRate,
	}
}

func GetWatchNamespaces(ctx context.Context, settings *InitialSettings) []string {
	return []string{settings.InstallNamespace}
}

type InitialSettings struct {
	InstallNamespace string
	RefreshRate      time.Duration
}

func GetWatchOpts(ctx context.Context, settings *InitialSettings) clients.WatchOpts {
	refreshRate := settings.RefreshRate
	if settings.RefreshRate <= 0 {
		refreshRate = time.Second
	}
	return clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: refreshRate,
	}
}

type ClientSet interface {
	Mesh() v1.MeshClient

	MultiClusterHandlers() []handler.ClusterHandler
	LocalClientGo() client_go.Interface
}
type MeshBridgeClientSet interface {
	ClientSet
	MeshBridge() v1.MeshBridgeClient
	MeshIngress() v1.MeshIngressClient
	ServiceEntry() v1alpha3.ServiceEntryClient
	Upstreams() gloov1.UpstreamClient
}

type MeshConfigClientSet interface {
	ClientSet
	// RbacConfig() v1alpha1.RbacConfigClient
	ClusterRbacConfig() v1alpha1.ClusterRbacConfigClient
}

type clientSet struct {
	meshBridge              v1.MeshBridgeClient
	mesh                    v1.MeshClient
	meshIngress             v1.MeshIngressClient
	serviceEntry            v1alpha3.ServiceEntryClient
	rbacConfigClient        v1alpha1.RbacConfigClient
	clusterRbacConfigClient v1alpha1.ClusterRbacConfigClient
	upstreams               gloov1.UpstreamClient
	services                kubernetes.ServiceClient
	pods                    kubernetes.PodClient
	mcHandlers              []handler.ClusterHandler

	localKube client_go.Interface

	// internal_watcher objects, used for lazy client loading
	ctx          context.Context
	cfg          *rest.Config
	watchHandler multicluster.ClientForClusterHandler
	settings     *InitialSettings
	cacheManger  CacheManager
}

func (c *clientSet) MeshBridge() v1.MeshBridgeClient {
	if c.meshBridge == nil {
		c.meshBridge = MustGetMeshBridgeClient(c.ctx, c.cfg, c.settings)
	}
	return c.meshBridge
}

func (c *clientSet) Mesh() v1.MeshClient {
	if c.mesh == nil {
		c.mesh = MustGetMeshClient(c.ctx, c.cfg, c.settings)
	}
	return c.mesh
}

func (c *clientSet) MeshIngress() v1.MeshIngressClient {
	if c.meshIngress == nil {
		c.meshIngress = MustGetMeshIngressClient(c.ctx, c.cfg, c.settings)
	}
	return c.meshIngress
}

func (c *clientSet) ServiceEntry() v1alpha3.ServiceEntryClient {
	if c.serviceEntry == nil {
		c.serviceEntry = MustGetServiceEntryClient(c.ctx, c.cfg, c.settings)
	}
	return c.serviceEntry
}

// func (c *clientSet) RbacConfig() v1alpha1.RbacConfigClient {
// 	if c.rbacConfigClient == nil {
// 		c.rbacConfigClient = MustGetRbacConfigClient(c.ctx, c.cfg, c.settings)
// 	}
// 	return c.rbacConfigClient
// }

func (c *clientSet) ClusterRbacConfig() v1alpha1.ClusterRbacConfigClient {
	if c.clusterRbacConfigClient == nil {
		c.clusterRbacConfigClient = MustGetClusterRbacConfigClient(c.ctx, c.cfg, c.settings)
	}
	return c.clusterRbacConfigClient
}

func (c *clientSet) Upstreams() gloov1.UpstreamClient {
	if c.upstreams == nil {
		usClient, usHandler := MustGetUpstreamClient(c.ctx, c.cacheManger.SharedCache(c.ctx), c.settings)
		c.upstreams = usClient
		c.mcHandlers = append(c.mcHandlers, usHandler)
	}
	return c.upstreams
}

func (c *clientSet) LocalClientGo() client_go.Interface {
	if c.localKube == nil {
		c.localKube = client_go.NewForConfigOrDie(c.cfg)
	}
	return c.localKube
}

func (c *clientSet) MultiClusterHandlers() []handler.ClusterHandler {
	// TODO(EItanya): figure out a way to lazy load multi cluster clients
	return c.mcHandlers
}

func MustGetMeshBridgeClientSet(ctx context.Context, cm CacheManager, cfg *rest.Config,
	watchHandler multicluster.ClientForClusterHandler, settings *InitialSettings) MeshBridgeClientSet {

	upstreamClient, upstreamGetter := MustGetUpstreamClient(ctx, cm.SharedCache(ctx), settings)
	return &clientSet{
		ctx:          ctx,
		cfg:          cfg,
		watchHandler: watchHandler,
		settings:     settings,
		cacheManger:  cm,
		upstreams:    upstreamClient,
		mcHandlers:   []handler.ClusterHandler{upstreamGetter},
	}
}

func MustGetMeshConfigClientSet(ctx context.Context, cm CacheManager, cfg *rest.Config,
	watchHandler multicluster.ClientForClusterHandler, settings *InitialSettings) MeshConfigClientSet {

	return &clientSet{
		ctx:          ctx,
		cfg:          cfg,
		watchHandler: watchHandler,
		settings:     settings,
		cacheManger:  cm,
	}
}

func MustGetMeshBridgeClient(ctx context.Context, cfg *rest.Config, settings *InitialSettings) v1.MeshBridgeClient {
	client, err := GetMeshBridgeClient(ctx, cfg, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get Mesh bridge client")
	}
	return client
}

func GetMeshBridgeClient(ctx context.Context, cfg *rest.Config, settings *InitialSettings) (v1.MeshBridgeClient, error) {
	skipCrdCreation := true
	namespaceWhitelist := []string{settings.InstallNamespace}
	contextutils.LoggerFrom(ctx).Infow("Getting Mesh bridge client",
		zap.Bool("skipCrdCreation", skipCrdCreation),
		zap.Strings("namespaceWhitelist", namespaceWhitelist))
	client, err := v1.NewMeshBridgeClient(&factory.KubeResourceClientFactory{
		Crd:                v1.MeshBridgeCrd,
		Cfg:                cfg,
		SharedCache:        kube.NewKubeCache(ctx),
		SkipCrdCreation:    skipCrdCreation,
		NamespaceWhitelist: namespaceWhitelist,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get Mesh bridge client")
	}
	err = client.Register()
	if err != nil {
		return nil, eris.Wrapf(err, "failed to register Mesh bridge client")
	}
	return client, nil
}

func MustGetMeshClient(ctx context.Context, cfg *rest.Config, settings *InitialSettings) v1.MeshClient {
	client, err := GetMeshClient(ctx, cfg, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get Mesh bridge client")
	}
	return client
}

func GetMeshClient(ctx context.Context, cfg *rest.Config, settings *InitialSettings) (v1.MeshClient, error) {
	skipCrdCreation := true
	namespaceWhitelist := []string{settings.InstallNamespace}
	contextutils.LoggerFrom(ctx).Infow("Getting mesh client",
		zap.Bool("skipCrdCreation", skipCrdCreation),
		zap.Strings("namespaceWhitelist", namespaceWhitelist))
	client, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
		Crd:                v1.MeshCrd,
		Cfg:                cfg,
		SharedCache:        kube.NewKubeCache(ctx),
		SkipCrdCreation:    skipCrdCreation,
		NamespaceWhitelist: namespaceWhitelist,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get mesh client")
	}
	err = client.Register()
	if err != nil {
		return nil, eris.Wrapf(err, "failed to register mesh client")
	}
	return client, nil
}

func MustGetMeshIngressClient(ctx context.Context, cfg *rest.Config, settings *InitialSettings) v1.MeshIngressClient {
	client, err := GetMeshIngressClient(ctx, cfg, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get Mesh bridge client")
	}
	return client
}

func GetMeshIngressClient(ctx context.Context, cfg *rest.Config, settings *InitialSettings) (v1.MeshIngressClient, error) {
	skipCrdCreation := true
	namespaceWhitelist := []string{settings.InstallNamespace}
	contextutils.LoggerFrom(ctx).Infow("Getting Mesh ingress client",
		zap.Bool("skipCrdCreation", skipCrdCreation),
		zap.Strings("namespaceWhitelist", namespaceWhitelist))
	meshClient, err := v1.NewMeshIngressClient(&factory.KubeResourceClientFactory{
		Crd:                v1.MeshIngressCrd,
		Cfg:                cfg,
		SharedCache:        kube.NewKubeCache(ctx),
		SkipCrdCreation:    skipCrdCreation,
		NamespaceWhitelist: namespaceWhitelist,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get mesh ingress client")
	}
	err = meshClient.Register()
	if err != nil {
		return nil, eris.Wrapf(err, "failed to register mesh ingress client")
	}
	return meshClient, nil
}

func MustGetServiceEntryClient(ctx context.Context, cfg *rest.Config, settings *InitialSettings) v1alpha3.ServiceEntryClient {
	serviceEntryClient, err := GetServiceEntryClient(ctx, cfg, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get service entry client",
			zap.Error(err))
	}
	return serviceEntryClient
}

func GetServiceEntryClient(ctx context.Context, cfg *rest.Config, settings *InitialSettings) (v1alpha3.ServiceEntryClient, error) {
	skipCrdCreation := true
	namespaceWhitelist := []string{settings.InstallNamespace}
	contextutils.LoggerFrom(ctx).Infow("Getting Mesh bridge client",
		zap.Bool("skipCrdCreation", skipCrdCreation),
		zap.Strings("namespaceWhitelist", namespaceWhitelist))
	serviceEntryClient, err := v1alpha3.NewServiceEntryClient(&factory.KubeResourceClientFactory{
		Crd:                v1alpha3.ServiceEntryCrd,
		Cfg:                cfg,
		SharedCache:        kube.NewKubeCache(ctx),
		SkipCrdCreation:    skipCrdCreation,
		NamespaceWhitelist: namespaceWhitelist,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get service entry client")
	}
	err = serviceEntryClient.Register()
	if err != nil {
		return nil, eris.Wrapf(err, "Failed to register service entry client")
	}
	return serviceEntryClient, nil
}

// func MustGetRbacConfigClient(ctx context.Context, cfg *rest.Config, settings *InitialSettings) v1alpha1.RbacConfigClient {
// 	rbacConfigClient, err := GetRbacConfigClient(ctx, cfg, settings)
// 	if err != nil {
// 		contextutils.LoggerFrom(ctx).Fatalw("unable to get rbac config client",
// 			zap.Error(err))
// 	}
// 	return rbacConfigClient
// }
//
// func GetRbacConfigClient(ctx context.Context, cfg *rest.Config, settings *InitialSettings) (v1alpha1.RbacConfigClient, error) {
// 	skipCrdCreation := true
// 	namespaceWhitelist := []string{settings.InstallNamespace}
// 	contextutils.LoggerFrom(ctx).Infow("Getting rbac config client",
// 		zap.Bool("skipCrdCreation", skipCrdCreation),
// 		zap.Strings("namespaceWhitelist", namespaceWhitelist))
// 	rbacConfigClient, err := v1alpha1.NewRbacConfigClient(&factory.KubeResourceClientFactory{
// 		Crd:                v1alpha1.RbacConfigCrd,
// 		Cfg:                cfg,
// 		SharedCache:        kube.NewKubeCache(ctx),
// 		SkipCrdCreation:    skipCrdCreation,
// 		NamespaceWhitelist: namespaceWhitelist,
// 	})
// 	if err != nil {
// 		return nil, eris.Wrapf(err, "failed to get rbac config client")
// 	}
// 	err = rbacConfigClient.Register()
// 	if err != nil {
// 		return nil, eris.Wrapf(err, "Failed to register rbac config client")
// 	}
// 	return rbacConfigClient, nil
// }

func MustGetClusterRbacConfigClient(ctx context.Context, cfg *rest.Config,
	settings *InitialSettings) v1alpha1.ClusterRbacConfigClient {
	clusterRbacConfigClient, err := GetRbacClusterConfigClient(ctx, cfg, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get cluster rbac config client",
			zap.Error(err))
	}
	return clusterRbacConfigClient
}

func GetRbacClusterConfigClient(ctx context.Context, cfg *rest.Config,
	settings *InitialSettings) (v1alpha1.ClusterRbacConfigClient, error) {
	skipCrdCreation := true
	namespaceWhitelist := []string{settings.InstallNamespace}
	contextutils.LoggerFrom(ctx).Infow("Getting rbac config client",
		zap.Bool("skipCrdCreation", skipCrdCreation),
		zap.Strings("namespaceWhitelist", namespaceWhitelist))
	clusterRbacConfigClient, err := v1alpha1.NewClusterRbacConfigClient(&factory.KubeResourceClientFactory{
		Crd:                v1alpha1.ClusterRbacConfigCrd,
		Cfg:                cfg,
		SharedCache:        kube.NewKubeCache(ctx),
		SkipCrdCreation:    skipCrdCreation,
		NamespaceWhitelist: namespaceWhitelist,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get cluster rbac config client")
	}
	err = clusterRbacConfigClient.Register()
	if err != nil {
		return nil, eris.Wrapf(err, "Failed to register cluster rbac config client")
	}
	return clusterRbacConfigClient, nil
}

func MustGetUpstreamClient(ctx context.Context, sharedCacheGetter clustercache.CacheGetter,
	settings *InitialSettings) (gloov1.UpstreamClient, handler.ClusterHandler) {
	upstreamClient, handler, err := GetUpstreamClient(ctx, sharedCacheGetter, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get upstream client")
	}
	return upstreamClient, handler
}

func GetUpstreamClient(ctx context.Context, sharedCacheGetter clustercache.CacheGetter,
	settings *InitialSettings) (gloov1.UpstreamClient, handler.ClusterHandler, error) {
	skipCrdCreation := true
	namespaceWhitelist := []string{settings.InstallNamespace}
	upstreamClientFactory := clientfactory.NewKubeResourceClientFactory(sharedCacheGetter,
		gloov1.UpstreamCrd,
		skipCrdCreation,
		namespaceWhitelist,
		0,
		factory.NewResourceClientParams{ResourceType: &gloov1.Upstream{}})

	contextutils.LoggerFrom(ctx).Infow("Getting upstream client",
		zap.Bool("skipCrdCreation", skipCrdCreation),
		zap.Strings("namespaceWhitelist", namespaceWhitelist))

	upstreamClientGetter := multicluster.NewClusterClientManager(ctx, upstreamClientFactory)
	upstreamBaseClient := multicluster.NewMultiClusterResourceClient(&gloov1.Upstream{}, upstreamClientGetter)
	upstreamClient := gloov1.NewUpstreamClientWithBase(upstreamBaseClient)
	err := upstreamClient.Register()
	if err != nil {
		return nil, nil, eris.Wrapf(err, "Failed to register upstream client")
	}
	return upstreamClient, upstreamClientGetter, nil
}
