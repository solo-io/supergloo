package translation

import (
	discovery_smh_solo_io_v1alpha1_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
)

// the translator InputSnapshot contains the set of input resources for mesh
// translation
type Snapshot interface {
	// return the set of input MeshServices
	MeshServices() discovery_smh_solo_io_v1alpha1_sets.MeshServiceSet
	// return the set of input MeshWorkloads
	MeshWorkloads() discovery_smh_solo_io_v1alpha1_sets.MeshWorkloadSet
	// return the set of input Meshes
	Meshes() discovery_smh_solo_io_v1alpha1_sets.MeshSet
}

type snapshot struct {
	meshServices  discovery_smh_solo_io_v1alpha1_sets.MeshServiceSet
	meshWorkloads discovery_smh_solo_io_v1alpha1_sets.MeshWorkloadSet
	meshes        discovery_smh_solo_io_v1alpha1_sets.MeshSet
}

func NewSnapshot(
	meshServices discovery_smh_solo_io_v1alpha1_sets.MeshServiceSet,
	meshWorkloads discovery_smh_solo_io_v1alpha1_sets.MeshWorkloadSet,
	meshes discovery_smh_solo_io_v1alpha1_sets.MeshSet,
) Snapshot {
	return &snapshot{
		meshServices:  meshServices,
		meshWorkloads: meshWorkloads,
		meshes:        meshes,
	}
}

func (s snapshot) MeshServices() discovery_smh_solo_io_v1alpha1_sets.MeshServiceSet {
	return s.meshServices
}

func (s snapshot) MeshWorkloads() discovery_smh_solo_io_v1alpha1_sets.MeshWorkloadSet {
	return s.meshWorkloads
}

func (s snapshot) Meshes() discovery_smh_solo_io_v1alpha1_sets.MeshSet {
	return s.meshes
}
