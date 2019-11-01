package kube

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/kubeutils"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func MustGetClient(ctx context.Context, cfg *rest.Config) kubernetes.Interface {
	contextutils.LoggerFrom(ctx).Debugw("Getting kube client")
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Could not get kube client", zap.Error(err))
	}
	return client
}

func MustGetKubeConfig(ctx context.Context) *rest.Config {
	contextutils.LoggerFrom(ctx).Debugw("Getting kube client config.")
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Failed to get kubernetes config.", zap.Error(err))
	}
	return cfg
}

func NewKubeCoreCache(ctx context.Context, iface kubernetes.Interface) (cache.KubeCoreCache, error) {
	cache, err := cache.NewKubeCoreCache(ctx, iface)
	if err != nil {
		return nil, err
	}
	return cache, nil
}
