package detector

import (
	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
)

// a deployment MeshDetector detects Mesh control plane instances (e.g. Pilot)
// in a k8s Deployment.
// If detection fails, an error is returned
// If no mesh is detected, nil is returned
// Separate Detectors are implemented for different Mesh types / versions.
type MeshDetector interface {
	DetectMeshes(in input.Snapshot) (v1alpha2.MeshSlice, error)
}

// wrapper for multiple mesh detectors.
// returns the first successfully detected mesh
type MeshDetectors []MeshDetector

func (d MeshDetectors) DetectMeshes(in input.Snapshot) (v1alpha2.MeshSlice, error) {
	var allMeshes v1alpha2.MeshSlice
	var errs error
	for _, detector := range d {
		meshes, err := detector.DetectMeshes(in)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		allMeshes = append(allMeshes, meshes...)
	}
	return allMeshes, errs
}
