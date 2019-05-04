// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"sort"

	github_com_solo_io_supergloo_api_external_linkerd "github.com/solo-io/supergloo/api/external/linkerd"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
)

func NewServiceProfile(namespace, name string) *ServiceProfile {
	serviceprofile := &ServiceProfile{}
	serviceprofile.ServiceProfile.SetMetadata(core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return serviceprofile
}

// require custom resource to implement Clone() as well as resources.Resource interface

type CloneableServiceProfile interface {
	resources.Resource
	Clone() *github_com_solo_io_supergloo_api_external_linkerd.ServiceProfile
}

var _ CloneableServiceProfile = &github_com_solo_io_supergloo_api_external_linkerd.ServiceProfile{}

type ServiceProfile struct {
	github_com_solo_io_supergloo_api_external_linkerd.ServiceProfile
}

func (r *ServiceProfile) Clone() resources.Resource {
	return &ServiceProfile{ServiceProfile: *r.ServiceProfile.Clone()}
}

func (r *ServiceProfile) Hash() uint64 {
	clone := r.ServiceProfile.Clone()

	resources.UpdateMetadata(clone, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})

	return hashutils.HashAll(clone)
}

type ServiceProfileList []*ServiceProfile
type ServiceprofilesByNamespace map[string]ServiceProfileList

// namespace is optional, if left empty, names can collide if the list contains more than one with the same name
func (list ServiceProfileList) Find(namespace, name string) (*ServiceProfile, error) {
	for _, serviceProfile := range list {
		if serviceProfile.GetMetadata().Name == name {
			if namespace == "" || serviceProfile.GetMetadata().Namespace == namespace {
				return serviceProfile, nil
			}
		}
	}
	return nil, errors.Errorf("list did not find serviceProfile %v.%v", namespace, name)
}

func (list ServiceProfileList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, serviceProfile := range list {
		ress = append(ress, serviceProfile)
	}
	return ress
}

func (list ServiceProfileList) Names() []string {
	var names []string
	for _, serviceProfile := range list {
		names = append(names, serviceProfile.GetMetadata().Name)
	}
	return names
}

func (list ServiceProfileList) NamespacesDotNames() []string {
	var names []string
	for _, serviceProfile := range list {
		names = append(names, serviceProfile.GetMetadata().Namespace+"."+serviceProfile.GetMetadata().Name)
	}
	return names
}

func (list ServiceProfileList) Sort() ServiceProfileList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list ServiceProfileList) Clone() ServiceProfileList {
	var serviceProfileList ServiceProfileList
	for _, serviceProfile := range list {
		serviceProfileList = append(serviceProfileList, resources.Clone(serviceProfile).(*ServiceProfile))
	}
	return serviceProfileList
}

func (list ServiceProfileList) Each(f func(element *ServiceProfile)) {
	for _, serviceProfile := range list {
		f(serviceProfile)
	}
}

func (list ServiceProfileList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *ServiceProfile) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

func (byNamespace ServiceprofilesByNamespace) Add(serviceProfile ...*ServiceProfile) {
	for _, item := range serviceProfile {
		byNamespace[item.GetMetadata().Namespace] = append(byNamespace[item.GetMetadata().Namespace], item)
	}
}

func (byNamespace ServiceprofilesByNamespace) Clear(namespace string) {
	delete(byNamespace, namespace)
}

func (byNamespace ServiceprofilesByNamespace) List() ServiceProfileList {
	var list ServiceProfileList
	for _, serviceProfileList := range byNamespace {
		list = append(list, serviceProfileList...)
	}
	return list.Sort()
}

func (byNamespace ServiceprofilesByNamespace) Clone() ServiceprofilesByNamespace {
	cloned := make(ServiceprofilesByNamespace)
	for ns, list := range byNamespace {
		cloned[ns] = list.Clone()
	}
	return cloned
}
