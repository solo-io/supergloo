package sets

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"k8s.io/apimachinery/pkg/util/sets"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type MeshServiceSet interface {
	Set() sets.String
	List() []*zephyr_discovery.MeshService
	Map() map[string]*zephyr_discovery.MeshService
	Insert(meshService ...*zephyr_discovery.MeshService)
	Equal(meshServiceSet MeshServiceSet) bool
	Has(meshService *zephyr_discovery.MeshService) bool
	Delete(meshService *zephyr_discovery.MeshService)
	Union(set MeshServiceSet) MeshServiceSet
	Difference(set MeshServiceSet) MeshServiceSet
	Intersection(set MeshServiceSet) MeshServiceSet
}

type MeshWorkloadSet interface {
	Set() sets.String
	List() []*zephyr_discovery.MeshWorkload
	Map() map[string]*zephyr_discovery.MeshWorkload
	Insert(meshWorkload ...*zephyr_discovery.MeshWorkload)
	Equal(meshWorkloadSet MeshWorkloadSet) bool
	Has(meshWorkload *zephyr_discovery.MeshWorkload) bool
	Delete(meshWorkload *zephyr_discovery.MeshWorkload)
	Union(set MeshWorkloadSet) MeshWorkloadSet
	Difference(set MeshWorkloadSet) MeshWorkloadSet
	Intersection(set MeshWorkloadSet) MeshWorkloadSet
}
