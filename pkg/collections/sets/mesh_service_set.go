package sets

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"k8s.io/apimachinery/pkg/util/sets"
)

type meshServiceSet struct {
	set     sets.String
	mapping map[string]*zephyr_discovery.MeshService
}

func NewMeshServiceSet(meshServices ...*zephyr_discovery.MeshService) MeshServiceSet {
	set := sets.NewString()
	mapping := map[string]*zephyr_discovery.MeshService{}
	for _, meshService := range meshServices {
		key := clients.ToUniqueSingleClusterString(meshService.ObjectMeta)
		set.Insert(key)
		mapping[key] = meshService
	}
	return &meshServiceSet{set: set}
}

func (m *meshServiceSet) Set() sets.String {
	return m.set
}

func (m *meshServiceSet) List() []*zephyr_discovery.MeshService {
	var meshServices []*zephyr_discovery.MeshService
	for _, key := range m.set.List() {
		meshServices = append(meshServices, m.mapping[key])
	}
	return meshServices
}

func (m *meshServiceSet) Map() map[string]*zephyr_discovery.MeshService {
	return m.mapping
}

func (m *meshServiceSet) Insert(
	meshServices ...*zephyr_discovery.MeshService,
) {
	for _, meshService := range meshServices {
		key := clients.ToUniqueSingleClusterString(meshService.ObjectMeta)
		m.mapping[key] = meshService
		m.set.Insert(key)
	}
}

func (m *meshServiceSet) Has(meshService *zephyr_discovery.MeshService) bool {
	return m.set.Has(clients.ToUniqueSingleClusterString(meshService.ObjectMeta))
}

func (m *meshServiceSet) Equal(
	meshServiceSet MeshServiceSet,
) bool {
	return m.set.Equal(meshServiceSet.Set())
}

func (m *meshServiceSet) Delete(meshService *zephyr_discovery.MeshService) {
	key := clients.ToUniqueSingleClusterString(meshService.ObjectMeta)
	delete(m.mapping, key)
	m.set.Delete(key)
}

func (m *meshServiceSet) Union(set MeshServiceSet) MeshServiceSet {
	return NewMeshServiceSet(append(m.List(), set.List()...)...)
}

func (m *meshServiceSet) Difference(set MeshServiceSet) MeshServiceSet {
	newSet := m.set.Difference(set.Set())
	var newMeshServices []*zephyr_discovery.MeshService
	for key, _ := range newSet {
		val, ok := m.mapping[key]
		if !ok {
			val, ok = set.Map()[key]
		}
		newMeshServices = append(newMeshServices, val)
	}
	return NewMeshServiceSet(newMeshServices...)
}

func (m *meshServiceSet) Intersection(set MeshServiceSet) MeshServiceSet {
	newSet := m.set.Intersection(set.Set())
	var newMeshServices []*zephyr_discovery.MeshService
	for key, _ := range newSet {
		val, ok := m.mapping[key]
		if !ok {
			val, ok = set.Map()[key]
		}
		newMeshServices = append(newMeshServices, val)
	}
	return NewMeshServiceSet(newMeshServices...)
}
