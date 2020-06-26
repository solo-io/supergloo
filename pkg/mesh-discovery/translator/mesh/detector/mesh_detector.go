package detector

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

// a deployment MeshDetector detects Mesh control plane instance (e.g. Pilot)
// in a k8s Deployment.
// If detection fails, an error is returned
// If no mesh is detected, nil is returned
// Separate Detectors are implemented for different Mesh types / versions.
type MeshDetector interface {
	DetectMesh(deployment *appsv1.Deployment) (*v1alpha1.Mesh, error)
}

// wrapper for multiple mesh detectors.
// returns the first successfully detected mesh
type MeshDetectors []MeshDetector

func (d MeshDetectors) DetectMesh(deployment *appsv1.Deployment) (*v1alpha1.Mesh, error) {
	for _, detector := range d {
		if mesh, err := detector.DetectMesh(deployment); err != nil {
			return nil, err
		} else if mesh != nil {
			return mesh, nil
		}
	}
	return nil, nil
}
