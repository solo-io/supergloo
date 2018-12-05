package factory

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kubesecretplain"
	"reflect"
	"strings"

	"github.com/hashicorp/consul/api"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/configmap"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/consul"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/file"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kubesecret"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/vault"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type genericResourceClientFactory struct {
	opts ResourceClientFactory
}

type NewResourceClientParams struct {
	ResourceType resources.Resource
	Token        string
}

// TODO(ilackarms): more opts validation
func newResourceClient(factory ResourceClientFactory, params NewResourceClientParams) (clients.ResourceClient, error) {
	resourceType := params.ResourceType
	switch opts := factory.(type) {
	case *KubeResourceClientFactory:
		if params.Token != "" {
			opts.Cfg.BearerToken = strings.TrimPrefix(params.Token, "Bearer ")
		}
		inputResource, ok := params.ResourceType.(resources.InputResource)
		if !ok {
			return nil, errors.Errorf("the kubernetes crd client can only be used for input resources, received type %v", resources.Kind(resourceType))
		}
		if opts.Crd.Type == nil {
			return nil, errors.Errorf("must provide a crd for the kube resource client")
		}
		if opts.SharedCache == nil {
			return nil, errors.Errorf("must provide a shared cache for the kube resource client")
		}
		if opts.Cfg == nil {
			return nil, errors.Errorf("must provide a resclient.Config for the kube resource client")
		}
		return kube.NewResourceClient(opts.Crd, opts.Cfg, opts.SharedCache, inputResource)
	case *ConsulResourceClientFactory:
		return consul.NewResourceClient(opts.Consul, opts.RootKey, resourceType), nil
	case *FileResourceClientFactory:
		return file.NewResourceClient(opts.RootDir, resourceType), nil
	case *MemoryResourceClientFactory:
		return memory.NewResourceClient(opts.Cache, resourceType), nil
	case *KubeConfigMapClientFactory:
		return configmap.NewResourceClient(opts.Clientset, resourceType)
	case *KubeSecretClientFactory:
		if opts.PlainSecrets {
			return kubesecretplain.NewResourceClient(opts.Clientset, resourceType)
		}
		return kubesecret.NewResourceClient(opts.Clientset, resourceType)
	case *VaultSecretClientFactory:
		return vault.NewResourceClient(opts.Vault, opts.RootKey, resourceType), nil
	}
	panic("unsupported type " + reflect.TypeOf(factory).Name())
}

// https://golang.org/doc/faq#generics
type ResourceClientFactory interface {
	NewResourceClient(params NewResourceClientParams) (clients.ResourceClient, error)
}

type KubeResourceClientFactory struct {
	Crd         crd.Crd
	Cfg         *rest.Config
	SharedCache *kube.KubeCache
}

func (f *KubeResourceClientFactory) NewResourceClient(params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(f, params)
}

type ConsulResourceClientFactory struct {
	Consul  *api.Client
	RootKey string
}

func (f *ConsulResourceClientFactory) NewResourceClient(params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(f, params)
}

type FileResourceClientFactory struct {
	RootDir string
}

func (f *FileResourceClientFactory) NewResourceClient(params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(f, params)
}

type MemoryResourceClientFactory struct {
	Cache memory.InMemoryResourceCache
}

func (f *MemoryResourceClientFactory) NewResourceClient(params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(f, params)
}

type KubeConfigMapClientFactory struct {
	Clientset kubernetes.Interface
}

func (f *KubeConfigMapClientFactory) NewResourceClient(params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(f, params)
}

type KubeSecretClientFactory struct {
	Clientset kubernetes.Interface
	// set this  to true if resource fields are all strings
	// resources will be stored as plain kubernetes secrets without serializing/deserializing
	PlainSecrets bool
}

func (f *KubeSecretClientFactory) NewResourceClient(params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(f, params)
}

type VaultSecretClientFactory struct {
	Vault   *vaultapi.Client
	RootKey string
}

func (f *VaultSecretClientFactory) NewResourceClient(params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(f, params)
}
