// Code generated by protoc-gen-solo-kit. DO NOT EDIT.

package v1alpha1

import (
	"sort"

	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TODO: modify as needed to populate additional fields
func NewServiceRole(namespace, name string) *ServiceRole {
	return &ServiceRole{
		Metadata: core.Metadata{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func (r *ServiceRole) SetStatus(status core.Status) {
	r.Status = status
}

func (r *ServiceRole) SetMetadata(meta core.Metadata) {
	r.Metadata = meta
}

type ServiceRoleList []*ServiceRole
type ServiceRolesByNamespace map[string]ServiceRoleList

// namespace is optional, if left empty, names can collide if the list contains more than one with the same name
func (list ServiceRoleList) Find(namespace, name string) (*ServiceRole, error) {
	for _, serviceRole := range list {
		if serviceRole.Metadata.Name == name {
			if namespace == "" || serviceRole.Metadata.Namespace == namespace {
				return serviceRole, nil
			}
		}
	}
	return nil, errors.Errorf("list did not find serviceRole %v.%v", namespace, name)
}

func (list ServiceRoleList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, serviceRole := range list {
		ress = append(ress, serviceRole)
	}
	return ress
}

func (list ServiceRoleList) AsInputResources() resources.InputResourceList {
	var ress resources.InputResourceList
	for _, serviceRole := range list {
		ress = append(ress, serviceRole)
	}
	return ress
}

func (list ServiceRoleList) Names() []string {
	var names []string
	for _, serviceRole := range list {
		names = append(names, serviceRole.Metadata.Name)
	}
	return names
}

func (list ServiceRoleList) NamespacesDotNames() []string {
	var names []string
	for _, serviceRole := range list {
		names = append(names, serviceRole.Metadata.Namespace+"."+serviceRole.Metadata.Name)
	}
	return names
}

func (list ServiceRoleList) Sort() ServiceRoleList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Metadata.Less(list[j].Metadata)
	})
	return list
}

func (list ServiceRoleList) Clone() ServiceRoleList {
	var serviceRoleList ServiceRoleList
	for _, serviceRole := range list {
		serviceRoleList = append(serviceRoleList, proto.Clone(serviceRole).(*ServiceRole))
	}
	return serviceRoleList
}

func (list ServiceRoleList) ByNamespace() ServiceRolesByNamespace {
	byNamespace := make(ServiceRolesByNamespace)
	for _, serviceRole := range list {
		byNamespace.Add(serviceRole)
	}
	return byNamespace
}

func (byNamespace ServiceRolesByNamespace) Add(serviceRole ...*ServiceRole) {
	for _, item := range serviceRole {
		byNamespace[item.Metadata.Namespace] = append(byNamespace[item.Metadata.Namespace], item)
	}
}

func (byNamespace ServiceRolesByNamespace) Clear(namespace string) {
	delete(byNamespace, namespace)
}

func (byNamespace ServiceRolesByNamespace) List() ServiceRoleList {
	var list ServiceRoleList
	for _, serviceRoleList := range byNamespace {
		list = append(list, serviceRoleList...)
	}
	return list.Sort()
}

func (byNamespace ServiceRolesByNamespace) Clone() ServiceRolesByNamespace {
	return byNamespace.List().Clone().ByNamespace()
}

var _ resources.Resource = &ServiceRole{}

// Kubernetes Adapter for ServiceRole

func (o *ServiceRole) GetObjectKind() schema.ObjectKind {
	t := ServiceRoleCrd.TypeMeta()
	return &t
}

func (o *ServiceRole) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*ServiceRole)
}

var ServiceRoleCrd = crd.NewCrd("rbac.istio.io",
	"serviceroles",
	"rbac.istio.io",
	"v1alpha1",
	"ServiceRole",
	"svcrole",
	&ServiceRole{})
