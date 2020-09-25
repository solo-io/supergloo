package appmesh

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"

	"github.com/solo-io/skv2/contrib/pkg/sets"

	corev1 "k8s.io/api/core/v1"
)

// detects an appmesh sidecar
type sidecarDetector struct {
	ctx context.Context
}

func NewSidecarDetector(ctx context.Context) *sidecarDetector {
	ctx = contextutils.WithLogger(ctx, "appmesh-sidecar-detector")
	return &sidecarDetector{ctx: ctx}
}

func (d sidecarDetector) DetectMeshSidecar(pod *corev1.Pod, meshes v1alpha2sets.MeshSet) *v1alpha2.Mesh {
	if !containsSidecarContainer(pod.Spec.Containers) {
		return nil
	}

	for _, mesh := range meshes.List() {
		appmesh := mesh.Spec.GetAwsAppMesh()
		if appmesh == nil {
			continue
		}

		// TODO joekelley this assumes that each cluster is managed by only one mesh; instead use virtual node name env var?
		if stringutils.ContainsString(pod.ClusterName, appmesh.Clusters) {
			return mesh
		}
	}

	contextutils.LoggerFrom(d.ctx).Warnw("warning: no mesh found corresponding to pod with appmesh sidecar", "pod", sets.Key(pod))

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
	return strings.Contains(imageName, "appmesh") && strings.Contains(imageName, "envoy")
}
