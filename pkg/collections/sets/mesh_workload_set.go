package sets

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"k8s.io/apimachinery/pkg/util/sets"
)

type meshWorkloadSet struct {
	set     sets.String
	mapping map[string]*zephyr_discovery.MeshWorkload
}

func NewMeshWorkloadSet(meshWorkloads ...*zephyr_discovery.MeshWorkload) MeshWorkloadSet {
	set := sets.NewString()
	mapping := map[string]*zephyr_discovery.MeshWorkload{}
	for _, meshWorkload := range meshWorkloads {
		key := clients.ToUniqueSingleClusterString(meshWorkload.ObjectMeta)
		set.Insert(key)
		mapping[key] = meshWorkload
	}
	return &meshWorkloadSet{set: set}
}

func (m *meshWorkloadSet) Set() sets.String {
	return m.set
}

func (m *meshWorkloadSet) List() []*zephyr_discovery.MeshWorkload {
	var meshWorkloads []*zephyr_discovery.MeshWorkload
	for _, key := range m.set.List() {
		meshWorkloads = append(meshWorkloads, m.mapping[key])
	}
	return meshWorkloads
}

func (m *meshWorkloadSet) Map() map[string]*zephyr_discovery.MeshWorkload {
	return m.mapping
}

func (m *meshWorkloadSet) Insert(
	meshWorkloads ...*zephyr_discovery.MeshWorkload,
) {
	for _, meshWorkload := range meshWorkloads {
		key := clients.ToUniqueSingleClusterString(meshWorkload.ObjectMeta)
		m.mapping[key] = meshWorkload
		m.set.Insert(key)
	}
}

func (m *meshWorkloadSet) Has(meshWorkload *zephyr_discovery.MeshWorkload) bool {
	return m.set.Has(clients.ToUniqueSingleClusterString(meshWorkload.ObjectMeta))
}

func (m *meshWorkloadSet) Equal(
	meshWorkloadSet MeshWorkloadSet,
) bool {
	return m.set.Equal(meshWorkloadSet.Set())
}

func (m *meshWorkloadSet) Delete(meshWorkload *zephyr_discovery.MeshWorkload) {
	key := clients.ToUniqueSingleClusterString(meshWorkload.ObjectMeta)
	delete(m.mapping, key)
	m.set.Delete(key)
}

func (m *meshWorkloadSet) Union(set MeshWorkloadSet) MeshWorkloadSet {
	return NewMeshWorkloadSet(append(m.List(), set.List()...)...)
}

func (m *meshWorkloadSet) Difference(set MeshWorkloadSet) MeshWorkloadSet {
	newSet := m.set.Difference(set.Set())
	var newMeshWorkloads []*zephyr_discovery.MeshWorkload
	for key, _ := range newSet {
		val, ok := m.mapping[key]
		if !ok {
			val, ok = set.Map()[key]
		}
		newMeshWorkloads = append(newMeshWorkloads, val)
	}
	return NewMeshWorkloadSet(newMeshWorkloads...)
}

func (m *meshWorkloadSet) Intersection(set MeshWorkloadSet) MeshWorkloadSet {
	newSet := m.set.Intersection(set.Set())
	var newMeshWorkloads []*zephyr_discovery.MeshWorkload
	for key, _ := range newSet {
		val, ok := m.mapping[key]
		if !ok {
			val, ok = set.Map()[key]
		}
		newMeshWorkloads = append(newMeshWorkloads, val)
	}
	return NewMeshWorkloadSet(newMeshWorkloads...)
}
