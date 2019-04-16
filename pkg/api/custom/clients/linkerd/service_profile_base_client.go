package linkerd

import (
	"sort"

	"github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha1"
	linkerdclient "github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/api/custom/linkerd"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type ResourceClient struct {
	linkerdClient linkerdclient.Interface
	cache         Cache
}

func NewResourceClient(linkerdClient linkerdclient.Interface, cache Cache) *ResourceClient {
	return &ResourceClient{
		linkerdClient: linkerdClient,
		cache:         cache,
	}
}

func FromKube(sp *v1alpha1.ServiceProfile) *v1.ServiceProfile {
	deepCopy := sp.DeepCopy()
	baseType := linkerd.ServiceProfile(*deepCopy)
	resource := &v1.ServiceProfile{
		ServiceProfile: baseType,
	}

	return resource
}

func ToKube(resource resources.Resource) (*v1alpha1.ServiceProfile, error) {
	serviceProfileResource, ok := resource.(*v1.ServiceProfile)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to service profile client", resources.Kind(resource))
	}

	serviceProfile := v1alpha1.ServiceProfile(serviceProfileResource.ServiceProfile)

	return &serviceProfile, nil
}

var _ clients.ResourceClient = &ResourceClient{}

func (rc *ResourceClient) Kind() string {
	return resources.Kind(&v1.ServiceProfile{})
}

func (rc *ResourceClient) NewResource() resources.Resource {
	return resources.Clone(&v1.ServiceProfile{})
}

func (rc *ResourceClient) Register() error {
	return nil
}

func (rc *ResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	serviceProfileObj, err := rc.linkerdClient.LinkerdV1alpha1().ServiceProfiles(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading serviceProfileObj from kubernetes")
	}
	resource := FromKube(serviceProfileObj)

	if resource == nil {
		return nil, errors.Errorf("serviceProfileObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *ResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	serviceProfileObj, err := ToKube(resource)
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
		if _, err := rc.linkerdClient.LinkerdV1alpha1().ServiceProfiles(meta.Namespace).Update(serviceProfileObj); err != nil {
			return nil, errors.Wrapf(err, "updating kube serviceProfileObj %v", serviceProfileObj.Name)
		}
	} else {
		if _, err := rc.linkerdClient.LinkerdV1alpha1().ServiceProfiles(meta.Namespace).Create(serviceProfileObj); err != nil {
			return nil, errors.Wrapf(err, "creating kube serviceProfileObj %v", serviceProfileObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(serviceProfileObj.Namespace, serviceProfileObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *ResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.linkerdClient.LinkerdV1alpha1().ServiceProfiles(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting serviceProfileObj %v", name)
	}
	return nil
}

func (rc *ResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	serviceProfileObjList, err := rc.cache.ServiceProfileLister().ServiceProfiles(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing serviceProfiles level")
	}
	var resourceList resources.ResourceList
	for _, serviceProfileObj := range serviceProfileObjList {
		resource := FromKube(serviceProfileObj)

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

func (rc *ResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	opts = opts.WithDefaults()
	watch := rc.cache.Subscribe()

	resourcesChan := make(chan resources.ResourceList)
	errs := make(chan error)
	// prevent flooding the channel with duplicates
	var previous *resources.ResourceList
	updateResourceList := func() {
		list, err := rc.List(namespace, clients.ListOpts{
			Ctx:      opts.Ctx,
			Selector: opts.Selector,
		})
		if err != nil {
			errs <- err
			return
		}
		if previous != nil {
			if list.Equal(*previous) {
				return
			}
		}
		previous = &list
		resourcesChan <- list
	}

	go func() {
		defer rc.cache.Unsubscribe(watch)
		defer close(resourcesChan)
		defer close(errs)

		// watch should open up with an initial read
		updateResourceList()
		for {
			select {
			case _, ok := <-watch:
				if !ok {
					return
				}
				updateResourceList()
			case <-opts.Ctx.Done():
				return
			}
		}
	}()

	return resourcesChan, errs, nil
}

func (rc *ResourceClient) exist(namespace, name string) bool {
	_, err := rc.linkerdClient.LinkerdV1alpha1().ServiceProfiles(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
