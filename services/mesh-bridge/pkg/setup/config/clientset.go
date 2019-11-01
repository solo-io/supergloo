package config

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

func GetInitialSettings(installNamespace string, settings *OperatorConfig) InitialSettings {
	return InitialSettings{
		InstallNamespace: installNamespace,
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

type ClientSet struct {
	meshBridge   v1.MeshBridgeClient
	serviceEntry v1alpha3.ServiceEntryClient
}

func MustGetClientSet(ctx context.Context, cfg *rest.Config, settings InitialSettings) *ClientSet {
	return &ClientSet{
		meshBridge:   MustGetMeshBridgeClient(ctx, cfg, settings),
		serviceEntry: MustGetServiceEntryClient(ctx, cfg, settings),
	}
}

func MustGetMeshBridgeClient(ctx context.Context, cfg *rest.Config, settings InitialSettings) v1.MeshBridgeClient {
	meshBridgeClient, err := GetMeshBridgeClient(ctx, cfg, settings)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to get mesh bridge client")
	}
	return meshBridgeClient
}

func GetMeshBridgeClient(ctx context.Context, cfg *rest.Config, settings InitialSettings) (v1.MeshBridgeClient, error) {
	skipCrdCreation := true
	namespaceWhitelist := []string{settings.InstallNamespace}
	contextutils.LoggerFrom(ctx).Infow("Getting mesh bridge client",
		zap.Bool("skipCrdCreation", skipCrdCreation),
		zap.Strings("namespaceWhitelist", namespaceWhitelist))
	meshBridgeClient, err := v1.NewMeshBridgeClient(&factory.KubeResourceClientFactory{
		Crd:                v1.MeshBridgeCrd,
		Cfg:                cfg,
		SharedCache:        kube.NewKubeCache(ctx),
		SkipCrdCreation:    skipCrdCreation,
		NamespaceWhitelist: namespaceWhitelist,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get mesh bridge client")
	}
	err = meshBridgeClient.Register()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to register mesh bridge client")
	}
	return meshBridgeClient, nil
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
	contextutils.LoggerFrom(ctx).Infow("Getting mesh bridge client",
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
