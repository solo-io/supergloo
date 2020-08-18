package osm

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	corev1 "k8s.io/api/core/v1"
)

// TODO(ilackarms): currently we produce a mesh ref that maps directly to the cluster

const (
	// proxy image
	sidecarProxy = "envoyproxy/envoy-alpine"
	// init container image
	proxyInit = "openservicemesh/init"
	// init container name
	proxyInitName = "osm-init"
)

// detects an osm sidecar
type sidecarDetector struct {
	ctx context.Context
}

func NewSidecarDetector(ctx context.Context) *sidecarDetector {
	ctx = contextutils.WithLogger(ctx, "linkerd-sidecar-detector")
	return &sidecarDetector{ctx: ctx}
}

/*
	OSM uses vanilla envoy sidecars currently, specifically `envoyproxy/envoy-alpine`.
*/
func (s *sidecarDetector) DetectMeshSidecar(pod *corev1.Pod, meshes v1alpha2sets.MeshSet) *v1alpha2.Mesh {
	if !(containsInitContainer(pod.Spec.InitContainers) && containsSidecar(pod.Spec.Containers)) {
		return nil
	}

	for _, mesh := range meshes.List() {
		osmMesh := mesh.Spec.GetOsm()
		if osmMesh == nil {
			continue
		}

		// TODO(ilackarms): currently we assume one mesh per cluster,
		// and that the control plane for a given sidecar is always
		// the mesh
		if osmMesh.Installation.GetCluster() == pod.ClusterName {
			return mesh
		}
	}

	contextutils.LoggerFrom(s.ctx).Warnw("warning: no mesh found corresponding to pod with osm sidecar", "pod", sets.Key(pod))

	return nil
}

func containsInitContainer(containers []corev1.Container) bool {
	for _, container := range containers {
		if strings.Contains(container.Image, proxyInit) && strings.Contains(container.Name, proxyInitName) {
			return true
		}
	}
	return false
}

func containsSidecar(containers []corev1.Container) bool {
	for _, container := range containers {
		if strings.Contains(container.Image, sidecarProxy) {
			return true
		}
	}
	return false
}
