package config

import (
	"context"
	"time"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/clientfactory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

func NewSharedCacheManager(ctx context.Context) (clustercache.CacheGetter, error) {
	return clustercache.NewCacheManager(ctx, kube.NewKubeSharedCacheForConfig)
}

func GetInitialSettings(installNamespace string, settings *OperatorConfig) InitialSettings {
	return InitialSettings{
		InstallNamespace: "",
		RefreshRate:      settings.RefreshRate,
	}
}

func GetWatchNamespaces(ctx context.Context, settings InitialSettings) []string {
	return []string{settings.InstallNamespace}
}

type InitialSettings struct {
	InstallNamespace string
	RefreshRate      time.Duration
}

func GetWatchOpts(ctx context.Context, settings InitialSettings) clients.WatchOpts {
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
	MeshBridge() v1.MeshBridgeClient
	Mesh() v1.MeshClient
	MeshIngress() v1.MeshIngressClient
	ServiceEntry() v1alpha3.ServiceEntryClient
	Upstreams() gloov1.UpstreamClient
	MultiClusterHandlers() []handler.ClusterHandler
}

type clientSet struct {
	meshBridge   v1.MeshBridgeClient
	mesh         v1.MeshClient
	meshIngress  v1.MeshIngressClient
	serviceEntry v1alpha3.ServiceEntryClient
	upstreams    gloov1.UpstreamClient
	mcHandlers   []handler.ClusterHandler
}

func (c *clientSet) MeshBridge() v1.MeshBridgeClient {
	return c.meshBridge
}

func (c *clientSet) Mesh() v1.MeshClient {
	return c.mesh
}

func (c *clientSet) MeshIngress() v1.MeshIngressClient {
	return c.meshIngress
}

func (c *clientSet) ServiceEntry() v1alpha3.ServiceEntryClient {
	return c.serviceEntry
}

func (c *clientSet) Upstreams() gloov1.UpstreamClient {
	return c.upstreams
}

func (c *clientSet) MultiClusterHandlers() []handler.ClusterHandler {
	return c.mcHandlers
}

func NewClientSet(
	meshBridge v1.MeshBridgeClient,
	mesh v1.MeshClient,
	meshIngress v1.MeshIngressClient,
	serviceEntry v1alpha3.ServiceEntryClient,
	upstreams gloov1.UpstreamClient) ClientSet {
	return &clientSet{
		meshBridge:   meshBridge,
		mesh:         mesh,
		meshIngress:  meshIngress,
		serviceEntry: serviceEntry,
		upstreams:    upstreams,
	}
}

func MustGetClientSet(ctx context.Context, sharedCacheGetter clustercache.CacheGetter, cfg *rest.Config,
	watchHandler multicluster.ClientForClusterHandler, settings InitialSettings) ClientSet {

	upstreamClient, upstreamHandler := MustGetUpstreamClient(ctx, sharedCacheGetter, cfg, watchHandler, settings)
	return &clientSet{
		meshBridge:   MustGetMeshBridgeClient(ctx, cfg, settings),
		serviceEntry: MustGetServiceEntryClient(ctx, cfg, settings),
		mesh:         MustGetMeshClient(ctx, cfg, settings),
		meshIngress:  MustGetMeshIngressClient(ctx, cfg, settings),
		upstreams:    upstreamClient,
		mcHandlers:   []handler.ClusterHandler{upstreamHandler},
	}
}

func MustGetMeshBridgeClient(ctx context.Context, cfg *rest.Config, settings InitialSettings) v1.MeshBridgeClient {
	client, err := GetMeshBridgeClient(ctx, cfg, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get Mesh bridge client")
	}
	return client
}

func GetMeshBridgeClient(ctx context.Context, cfg *rest.Config, settings InitialSettings) (v1.MeshBridgeClient, error) {
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
		return nil, errors.Wrapf(err, "failed to get Mesh bridge client")
	}
	err = client.Register()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to register Mesh bridge client")
	}
	return client, nil
}

func MustGetMeshClient(ctx context.Context, cfg *rest.Config, settings InitialSettings) v1.MeshClient {
	client, err := GetMeshClient(ctx, cfg, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get Mesh bridge client")
	}
	return client
}

func GetMeshClient(ctx context.Context, cfg *rest.Config, settings InitialSettings) (v1.MeshClient, error) {
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
		return nil, errors.Wrapf(err, "failed to get mesh client")
	}
	err = client.Register()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to register mesh client")
	}
	return client, nil
}

func MustGetMeshIngressClient(ctx context.Context, cfg *rest.Config, settings InitialSettings) v1.MeshIngressClient {
	client, err := GetMeshIngressClient(ctx, cfg, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get Mesh bridge client")
	}
	return client
}

func GetMeshIngressClient(ctx context.Context, cfg *rest.Config, settings InitialSettings) (v1.MeshIngressClient, error) {
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
		return nil, errors.Wrapf(err, "failed to get mesh ingress client")
	}
	err = meshClient.Register()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to register mesh ingress client")
	}
	return meshClient, nil
}

func MustGetServiceEntryClient(ctx context.Context, cfg *rest.Config, settings InitialSettings) v1alpha3.ServiceEntryClient {
	serviceEntryClient, err := GetServiceEntryClient(ctx, cfg, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get service entry client")
	}
	return serviceEntryClient
}

func GetServiceEntryClient(ctx context.Context, cfg *rest.Config, settings InitialSettings) (v1alpha3.ServiceEntryClient, error) {
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
		return nil, errors.Wrapf(err, "failed to get service entry client")
	}
	err = serviceEntryClient.Register()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to register service entry client")
	}
	return serviceEntryClient, nil
}

func MustGetUpstreamClient(ctx context.Context, sharedCacheGetter clustercache.CacheGetter, cfg *rest.Config,
	watchHandler multicluster.ClientForClusterHandler, settings InitialSettings) (gloov1.UpstreamClient, handler.ClusterHandler) {
	upstreamClient, handler, err := GetUpstreamClient(ctx, sharedCacheGetter, cfg, watchHandler, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get upstream client")
	}
	return upstreamClient, handler
}

func GetUpstreamClient(ctx context.Context, sharedCacheGetter clustercache.CacheGetter, cfg *rest.Config,
	watchHandler multicluster.ClientForClusterHandler, settings InitialSettings) (gloov1.UpstreamClient, handler.ClusterHandler, error) {
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
		return nil, nil, errors.Wrapf(err, "Failed to register upstream client")
	}
	return upstreamClient, upstreamClientGetter, nil
}
