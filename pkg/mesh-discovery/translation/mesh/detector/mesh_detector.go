package detector

import (
	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
)

// a MeshDetector detects Mesh control plane instances (e.g. Pilot)
// in a snapshot of discovery resources.
// If detection fails, an error is returned
// If no mesh is detected, an empty list is returned
// Separate Detectors are implemented for different Mesh types / versions.
type MeshDetector interface {
	DetectMeshes(in input.DiscoveryInputSnapshot, settings *settingsv1.DiscoverySettings) (v1.MeshSlice, error)
}

// wrapper for multiple mesh detectors.
// returns all detected meshes
type MeshDetectors []MeshDetector

func (d MeshDetectors) DetectMeshes(in input.DiscoveryInputSnapshot, settings *settingsv1.DiscoverySettings) (v1.MeshSlice, error) {
	var allMeshes v1.MeshSlice
	var errs error
	for _, detector := range d {
		meshes, err := detector.DetectMeshes(in, settings)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		allMeshes = append(allMeshes, meshes...)
	}
	return allMeshes, errs
}
