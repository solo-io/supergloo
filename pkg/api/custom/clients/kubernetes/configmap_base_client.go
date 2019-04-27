package kubernetes

import (
	"sort"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/common"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/supergloo/api/custom/kubernetes/configmap"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	kubev1 "k8s.io/api/core/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type configMapResourceClient struct {
	common.KubeCoreResourceClient
}

func newResourceClient(kube kubernetes.Interface, cache cache.KubeCoreCache) *configMapResourceClient {
	return &configMapResourceClient{
		KubeCoreResourceClient: common.KubeCoreResourceClient{
			Kube:         kube,
			Cache:        cache,
			ResourceType: &configmap.ConfigMap{},
		},
	}
}

func NewConfigMapClient(kube kubernetes.Interface, cache cache.KubeCoreCache) v1.ConfigMapClient {
	resourceClient := newResourceClient(kube, cache)
	return v1.NewConfigMapClientWithBase(resourceClient)
}

func FromKubeConfigMap(cm *kubev1.ConfigMap) *v1.ConfigMap {

	podCopy := cm.DeepCopy()
	kubeConfigMap := configmap.ConfigMap(*podCopy)
	resource := &v1.ConfigMap{
		ConfigMap: kubeConfigMap,
	}

	return resource
}

func ToKubeConfigMap(resource resources.Resource) (*kubev1.ConfigMap, error) {
	cmResource, ok := resource.(*v1.ConfigMap)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to config-map-only client", resources.Kind(resource))
	}

	cm := kubev1.ConfigMap(cmResource.ConfigMap)

	return &cm, nil
}

var _ clients.ResourceClient = &configMapResourceClient{}

func (rc *configMapResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	podObj, err := rc.Kube.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading podObj from kubernetes")
	}
	resource := FromKubeConfigMap(podObj)

	if resource == nil {
		return nil, errors.Errorf("podObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *configMapResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	cmObj, err := ToKubeConfigMap(resource)
	if err != nil {
		return nil, err
	}

	original, err := rc.Read(meta.Namespace, meta.Name, clients.ReadOpts{
		Ctx: opts.Ctx,
	})
	if original != nil && err == nil {
		if !opts.OverwriteExisting {
			return nil, errors.NewExistErr(meta)
		}
		if meta.ResourceVersion != original.GetMetadata().ResourceVersion {
			return nil, errors.NewResourceVersionErr(meta.Namespace, meta.Name, meta.ResourceVersion, original.GetMetadata().ResourceVersion)
		}
		if _, err := rc.Kube.CoreV1().ConfigMaps(meta.Namespace).Update(cmObj); err != nil {
			return nil, errors.Wrapf(err, "updating kube podObj %v", cmObj.Name)
		}
	} else {
		if _, err := rc.Kube.CoreV1().ConfigMaps(meta.Namespace).Create(cmObj); err != nil {
			return nil, errors.Wrapf(err, "creating kube podObj %v", cmObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(cmObj.Namespace, cmObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *configMapResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.Kube.CoreV1().ConfigMaps(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting cmObj %v", name)
	}
	return nil
}

func (rc *configMapResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	cmListObj, err := rc.Cache.ConfigMapLister().ConfigMaps(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing pods level")
	}
	var resourceList resources.ResourceList
	for _, cmObj := range cmListObj {
		resource := FromKubeConfigMap(cmObj)

		if resource == nil {
			continue
		}
		resourceList = append(resourceList, resource)
	}

	sort.SliceStable(resourceList, func(i, j int) bool {
		return resourceList[i].GetMetadata().Name < resourceList[j].GetMetadata().Name
	})

	return resourceList, nil
}

func (rc *configMapResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	return common.KubeResourceWatch(rc.Cache, rc.List, namespace, opts)
}

func (rc *configMapResourceClient) exist(namespace, name string) bool {
	_, err := rc.Kube.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
