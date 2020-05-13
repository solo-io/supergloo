package kubeconfig

import (
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig/internal"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

func NewFileBackedRESTClientGetter(kubeLoader KubeLoader, kubeConfigPath, kubeContext string) resource.RESTClientGetter {
	return &restClientGetter{
		kubeLoader:     kubeLoader,
		kubeConfigPath: kubeConfigPath,
		kubeContext:    kubeContext,
	}
}

type restClientGetter struct {
	kubeLoader     KubeLoader
	kubeConfigPath string
	kubeContext    string
}

func (r *restClientGetter) ToRESTConfig() (*rest.Config, error) {
	return r.kubeLoader.GetRestConfigForContext(r.kubeConfigPath, r.kubeContext)
}

func (i *restClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	cfg, err := i.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return internal.ToDiscoveryClient(cfg)
}

func (i *restClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	cfg, err := i.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return internal.ToRESTMapper(cfg)
}
