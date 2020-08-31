package linkerd

// TODO(EItanya): Uncomment to re-enable linkerd discovery
// Currently commented out because of dependency issues
//
// import (
// 	"context"
// 	"strings"
//
// 	"github.com/solo-io/go-utils/contextutils"
// 	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
// 	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
// 	"github.com/solo-io/skv2/contrib/pkg/sets"
// 	corev1 "k8s.io/api/core/v1"
// )
//
// // TODO(ilackarms): currently we produce a mesh ref that maps directly to the cluster
//
// // detects an linkerd sidecar
// type sidecarDetector struct {
// 	ctx context.Context
// }
//
// func NewSidecarDetector(ctx context.Context) *sidecarDetector {
// 	ctx = contextutils.WithLogger(ctx, "linkerd-sidecar-detector")
// 	return &sidecarDetector{ctx: ctx}
// }
//
// func (d sidecarDetector) DetectMeshSidecar(pod *corev1.Pod, meshes v1alpha2sets.MeshSet) *v1alpha2.Mesh {
// 	if !containsSidecarContainer(pod.Spec.Containers) {
// 		return nil
// 	}
//
// 	for _, mesh := range meshes.List() {
// 		linkerd := mesh.Spec.GetLinkerd()
// 		if linkerd == nil {
// 			continue
// 		}
//
// 		// TODO(ilackarms): currently we assume one mesh per cluster,
// 		// and that the control plane for a given sidecar is always
// 		// the mesh
// 		if linkerd.Installation.GetCluster() == pod.ClusterName {
// 			return mesh
// 		}
// 	}
//
// 	contextutils.LoggerFrom(d.ctx).Warnw("warning: no mesh found corresponding to pod with linkerd sidecar", "pod", sets.Key(pod))
//
// 	return nil
// }
//
// func containsSidecarContainer(containers []corev1.Container) bool {
// 	for _, container := range containers {
// 		if isSidecarName(container.Name) {
// 			return true
// 		}
// 	}
// 	return false
// }
//
// func isSidecarName(imageName string) bool {
// 	return strings.Contains(imageName, "linkerd") && strings.Contains(imageName, "proxy")
// }
