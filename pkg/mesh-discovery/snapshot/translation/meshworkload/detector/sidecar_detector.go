package detector

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	corev1 "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./sidecar_detector.go -destination mocks/sidecar_detector.go

// a sidecar detector detects injected Mesh sidecars in a Pod
type SidecarDetector interface {
	// returns a ref to a mesh if the provided Pod contains a sidecar
	// pointing at that mesh. returns nil if the
	DetectMeshSidecar(pod *corev1.Pod, meshes v1alpha1sets.MeshSet) *v1alpha1.Mesh
}

// wrapper for multiple mesh detectors.
// returns the first successfully detected mesh
type SidecarDetectors []SidecarDetector

func (d SidecarDetectors) DetectMeshSidecar(pod *corev1.Pod, meshes v1alpha1sets.MeshSet) *v1alpha1.Mesh {
	for _, detector := range d {
		if mesh := detector.DetectMeshSidecar(pod, meshes); mesh != nil {
			return mesh
		}
	}
	return nil
}
