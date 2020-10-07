package appmesh

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"

	"github.com/solo-io/skv2/contrib/pkg/sets"

	corev1 "k8s.io/api/core/v1"
)

const (
	// Used to infer parent AppMesh Mesh name
	appMeshVirtualNodeEnvVarName = "APPMESH_VIRTUAL_NODE_NAME"
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
	sidecarContainer := getSidecar(pod.Spec.Containers)
	if sidecarContainer == nil {
		return nil
	}

	var sidecarMeshName string
	for _, envVar := range sidecarContainer.Env {
		if envVar.Name != appMeshVirtualNodeEnvVarName {
			continue
		}

		// Value takes format "mesh/<appmesh-mesh-name>/virtualNode/<virtual-node-name>"
		// VirtualNodeName value has significance for AppMesh functionality, reference:
		// https://docs.aws.amazon.com/eks/latest/userguide/mesh-k8s-integration.html
		split := strings.Split(envVar.Value, "/")
		if len(split) != 4 {
			contextutils.LoggerFrom(d.ctx).Warnw("warning: unexpected virtual node name format", "pod", sets.Key(pod), "virtualNode", envVar.Value)
			return nil
		}
		sidecarMeshName = split[1]
	}

	for _, mesh := range meshes.List() {
		appmesh := mesh.Spec.GetAwsAppMesh()
		if appmesh == nil {
			continue
		}

		// TODO joekelley this does not deduplicate across disparate accounts, which are not referenced on sidecars.
		if appmesh.AwsName == sidecarMeshName {
			return mesh
		}
	}

	contextutils.LoggerFrom(d.ctx).Warnw("warning: no mesh found corresponding to pod with appmesh sidecar", "pod", sets.Key(pod))
	return nil
}

func getSidecar(containers []corev1.Container) *corev1.Container {
	for _, container := range containers {
		if isSidecarImage(container.Image) {
			return &container
		}
	}
	return nil
}

func isSidecarImage(imageName string) bool {
	return strings.Contains(imageName, "appmesh") && strings.Contains(imageName, "envoy")
}
