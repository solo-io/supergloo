package istio

import (
	"context"
	"fmt"
	"strings"

	"github.com/rotisserie/eris"
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

func (d sidecarDetector) DetectMeshSidecar(pod *corev1.Pod, meshes v1alpha2sets.MeshSet) (*v1alpha2.Mesh, *v1alpha2.WorkloadSpec_ProxyInstance) {
	container, isProxy := findSidecarContainer(pod.Spec.Containers)
	if !isProxy {
		return nil, nil
	}

	instance, err := getProxyInstance(pod, container)
	if err != nil {
		contextutils.LoggerFrom(d.ctx).Warnf("could not detect instance from proxy container: %v", err)
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
			return mesh, instance
		}
	}

	contextutils.LoggerFrom(d.ctx).Warnw("warning: no mesh found corresponding to pod with istio sidecar", "pod", sets.Key(pod))

	return nil, nil
}

func findSidecarContainer(containers []corev1.Container) (corev1.Container, bool) {
	for _, container := range containers {
		if isSidecarImage(container.Image) {
			return container, true
		}
	}
	return corev1.Container{}, false
}

func isSidecarImage(imageName string) bool {
	return strings.Contains(imageName, "istio") && strings.Contains(imageName, "proxy")
}

func getProxyInstance(pod *corev1.Pod, container corev1.Container) (*v1alpha2.WorkloadSpec_ProxyInstance, error) {
	proxyType, err := getProxyType(container)
	if err != nil {
		return nil, err
	}

	return &v1alpha2.WorkloadSpec_ProxyInstance{
		ProxyNodeId: getNodeID(pod, proxyType),
	}, nil
}

// note: this uses a heuristic method to determine the proxy type
// at time of writing, this is can only be determined by inspecting the
// arguments to the proxy in the container
func getProxyType(container corev1.Container) (string, error) {
	if len(container.Args) < 2 {
		return "", eris.Errorf("args list too short")
	}
	// note(ilackarms): this assumption can potentially break between istio versions;
	// we may need to parameterize this function in the future with istio version
	proxyType := container.Args[1]
	switch proxyType {
	case "router", "sidecar":
		return proxyType, nil
	default:
		return "", eris.Errorf("second arg does not resolve to proxy type")
	}
}

func getNodeID(pod *corev1.Pod, proxyType string) string {
	if proxyType != "" {
		return fmt.Sprintf("%s~%s~%s.%s~%s.svc.cluster.local", proxyType, pod.Status.PodIP, pod.Name, pod.Namespace, pod.Namespace)
	}
	if strings.HasPrefix(pod.Name, "istio-ingressgateway") || strings.HasPrefix(pod.Name, "istio-egressgateway") {
		return fmt.Sprintf("router~%s~%s.%s~%s.svc.cluster.local", pod.Status.PodIP, pod.Name, pod.Namespace, pod.Namespace)
	}
	if strings.HasPrefix(pod.Name, "istio-ingress") {
		return fmt.Sprintf("ingress~%s~%s.%s~%s.svc.cluster.local", pod.Status.PodIP, pod.Name, pod.Namespace, pod.Namespace)
	}
	return fmt.Sprintf("sidecar~%s~%s.%s~%s.svc.cluster.local", pod.Status.PodIP, pod.Name, pod.Namespace, pod.Namespace)
}
