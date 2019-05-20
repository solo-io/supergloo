package smi

import (
	"sort"

	"github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/apis/specs/v1alpha1"
	specsclient "github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/gen/client/specs/clientset/versioned"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/api/external/smi/specs"
	sgspecs "github.com/solo-io/supergloo/pkg/api/external/smi/specs/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type httpRouteGroupBaseClient struct {
	specsClient specsclient.Interface
	cache       Cache
}

func NewHTTPRouteGroupClient(specsClient specsclient.Interface, cache Cache) sgspecs.HTTPRouteGroupClient {
	return sgspecs.NewHTTPRouteGroupClientWithBase(&httpRouteGroupBaseClient{
		specsClient: specsClient,
		cache:       cache,
	})
}

func HTTPRouteGroupFromKube(sp *v1alpha1.HTTPRouteGroup) *sgspecs.HTTPRouteGroup {
	deepCopy := sp.DeepCopy()
	baseType := specs.HTTPRouteGroup(*deepCopy)
	resource := &sgspecs.HTTPRouteGroup{
		HTTPRouteGroup: baseType,
	}

	return resource
}

func HTTPRouteGroupToKube(resource resources.Resource) (*v1alpha1.HTTPRouteGroup, error) {
	httpRouteGroupResource, ok := resource.(*sgspecs.HTTPRouteGroup)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to service profile client", resources.Kind(resource))
	}

	httpRouteGroup := v1alpha1.HTTPRouteGroup(httpRouteGroupResource.HTTPRouteGroup)

	return &httpRouteGroup, nil
}

var _ clients.ResourceClient = &httpRouteGroupBaseClient{}

func (rc *httpRouteGroupBaseClient) Kind() string {
	return resources.Kind(&sgspecs.HTTPRouteGroup{})
}

func (rc *httpRouteGroupBaseClient) NewResource() resources.Resource {
	return resources.Clone(&sgspecs.HTTPRouteGroup{})
}

func (rc *httpRouteGroupBaseClient) Register() error {
	return nil
}

func (rc *httpRouteGroupBaseClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	httpRouteGroupObj, err := rc.specsClient.SpecsV1alpha1().HTTPRouteGroups(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading httpRouteGroupObj from kubernetes")
	}
	resource := HTTPRouteGroupFromKube(httpRouteGroupObj)

	if resource == nil {
		return nil, errors.Errorf("httpRouteGroupObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *httpRouteGroupBaseClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	httpRouteGroupObj, err := HTTPRouteGroupToKube(resource)
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
		if _, err := rc.specsClient.SpecsV1alpha1().HTTPRouteGroups(meta.Namespace).Update(httpRouteGroupObj); err != nil {
			return nil, errors.Wrapf(err, "updating kube httpRouteGroupObj %v", httpRouteGroupObj.Name)
		}
	} else {
		if _, err := rc.specsClient.SpecsV1alpha1().HTTPRouteGroups(meta.Namespace).Create(httpRouteGroupObj); err != nil {
			return nil, errors.Wrapf(err, "creating kube httpRouteGroupObj %v", httpRouteGroupObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(httpRouteGroupObj.Namespace, httpRouteGroupObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *httpRouteGroupBaseClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.specsClient.SpecsV1alpha1().HTTPRouteGroups(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting httpRouteGroupObj %v", name)
	}
	return nil
}

func (rc *httpRouteGroupBaseClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	httpRouteGroupObjList, err := rc.cache.HTTPRouteGroupLister().HTTPRouteGroups(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing httpRouteGroups level")
	}
	var resourceList resources.ResourceList
	for _, httpRouteGroupObj := range httpRouteGroupObjList {
		resource := HTTPRouteGroupFromKube(httpRouteGroupObj)

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

func (rc *httpRouteGroupBaseClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
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

func (rc *httpRouteGroupBaseClient) exist(namespace, name string) bool {
	_, err := rc.specsClient.SpecsV1alpha1().HTTPRouteGroups(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
