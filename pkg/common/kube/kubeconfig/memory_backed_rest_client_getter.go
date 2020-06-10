package kubeconfig

import (
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig/internal"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type InMemoryRESTClientGetterFactory func(cfg *rest.Config) resource.RESTClientGetter

func NewInMemoryRESTClientGetterFactory() InMemoryRESTClientGetterFactory {
	return NewInMemoryRESTClientGetter
}

func NewInMemoryRESTClientGetter(cfg *rest.Config) resource.RESTClientGetter {
	return &inMemoryRESTClientGetter{cfg: cfg}
}

type inMemoryRESTClientGetter struct {
	cfg *rest.Config
}

func (i *inMemoryRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return i.cfg, nil
}

func (i *inMemoryRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	cfg, err := i.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return internal.ToDiscoveryClient(cfg)
}

func (i *inMemoryRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	cfg, err := i.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return internal.ToRESTMapper(cfg)
}
