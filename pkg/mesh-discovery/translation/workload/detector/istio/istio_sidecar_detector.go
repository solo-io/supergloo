package istio

import (
	"context"
	"strings"

	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	corev1 "k8s.io/api/core/v1"
)

// TODO(ilackarms): currently we produce a mesh ref that maps directly to the cluster

// detects an istio sidecar
type sidecarDetector struct {
	ctx context.Context
}

func NewSidecarDetector(ctx context.Context) *sidecarDetector {
	ctx = contextutils.WithLogger(ctx, "istio-sidecar-detector")
	return &sidecarDetector{ctx: ctx}
}

func (d sidecarDetector) DetectMeshSidecar(pod *corev1.Pod, meshes v1alpha2sets.MeshSet) *v1alpha2.Mesh {
	if !containsSidecarContainer(pod.Spec.Containers) {
		return nil
	}

	for _, mesh := range meshes.List() {
		istio := mesh.Spec.GetIstio()
		if istio == nil {
			continue
		}

		// TODO(ilackarms): currently we assume one mesh per cluster,
		// and that the control plane for a given sidecar is always
		// the mesh
		if istio.Installation.GetCluster() == pod.ClusterName {
			return mesh
		}
	}

	contextutils.LoggerFrom(d.ctx).Warnw("warning: no mesh found corresponding to pod with istio sidecar", "pod", sets.Key(pod))

	return nil
}

func containsSidecarContainer(containers []corev1.Container) bool {
	for _, container := range containers {
		if isSidecarImage(container.Image) {
			return true
		}
	}
	return false
}

func isSidecarImage(imageName string) bool {
	return strings.Contains(imageName, "istio") && strings.Contains(imageName, "proxy")
}
