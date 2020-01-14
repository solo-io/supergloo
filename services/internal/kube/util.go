package kube

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/multicluster/secretconverter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

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

func GetKubeConfigForCluster(ctx context.Context, kube kubernetes.Interface, clusterName string) (*rest.Config, error) {
	secrets, err := kube.CoreV1().Secrets("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, secret := range secrets.Items {
		if secret.Type != secretconverter.KubeCfgType {
			continue
		}
		kubeCfg, err := secretconverter.KubeCfgFromSecret(&secret)
		if err != nil {
			return nil, err
		}
		if kubeCfg.Cluster != clusterName {
			continue
		}
		raw, err := clientcmd.Write(kubeCfg.Config)
		if err != nil {
			return nil, err
		}
		restCfg, err := clientcmd.RESTConfigFromKubeConfig(raw)
		if err != nil {
			return nil, eris.Wrapf(err, "failed to construct *rest.Config from "+
				"kubeconfig %v", kubeCfg.Metadata.Ref())
		}
		return restCfg, nil
	}
	return nil, eris.Errorf("could not find a rest config for the given cluster name")
}
