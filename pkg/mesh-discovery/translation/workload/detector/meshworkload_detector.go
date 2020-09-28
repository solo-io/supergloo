package detector

import (
	"context"
	"strings"

	"github.com/solo-io/skv2/pkg/ezkube"

	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/utils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/types"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// the WorkloadDetector detects Workloads from workloads
// whose backing pods are injected with a Mesh sidecar.
// If no mesh is detected for the workload, nil is returned
type WorkloadDetector interface {
	DetectWorkload(workload types.Workload, meshes v1alpha2sets.MeshSet) *v1alpha2.Workload
}

const (
	replicaSetKind = "ReplicaSet"
	deploymentKind = "Deployment"
)

// detects a workload
type workloadDetector struct {
	ctx         context.Context
	pods        corev1sets.PodSet
	replicaSets appsv1sets.ReplicaSetSet
	detector    SidecarDetector
}

func NewWorkloadDetector(
	ctx context.Context,
	pods corev1sets.PodSet,
	replicaSets appsv1sets.ReplicaSetSet,
	detector SidecarDetector,
) WorkloadDetector {
	ctx = contextutils.WithLogger(ctx, "mesh-workload-detector")
	return &workloadDetector{
		ctx:         ctx,
		pods:        pods,
		replicaSets: replicaSets,
		detector:    detector,
	}
}

func (d workloadDetector) DetectWorkload(workload types.Workload, meshes v1alpha2sets.MeshSet) *v1alpha2.Workload {
	podsForWorkload := d.getPodsForWorkload(workload)

	mesh := d.getMeshForPods(podsForWorkload, meshes)

	if mesh == nil {
		return nil
	}

	meshRef := ezkube.MakeObjectRef(mesh)
	controllerRef := ezkube.MakeClusterObjectRef(workload)
	labels := workload.GetPodTemplate().Labels
	serviceAccount := workload.GetPodTemplate().Spec.ServiceAccountName

	outputMeta := utils.DiscoveredObjectMeta(workload)
	// append workload kind for uniqueness
	outputMeta.Name += "-" + strings.ToLower(workload.Kind())

	return &v1alpha2.Workload{
		ObjectMeta: outputMeta,
		Spec: v1alpha2.WorkloadSpec{
			WorkloadType: &v1alpha2.WorkloadSpec_Kubernetes{
				Kubernetes: &v1alpha2.WorkloadSpec_KubernetesWorkload{
					Controller:         controllerRef,
					PodLabels:          labels,
					ServiceAccountName: serviceAccount,
				},
			},
			Mesh: meshRef,
		},
	}
}

func (d workloadDetector) getMeshForPods(pods corev1sets.PodSet, meshes v1alpha2sets.MeshSet) *v1alpha2.Mesh {
	// as long as one pod is detected for a mesh, consider the set owned by that mesh.
	for _, pod := range pods.List() {
		if mesh := d.detector.DetectMeshSidecar(pod, meshes); mesh != nil {
			return mesh
		}
	}
	return nil
}

func (d workloadDetector) getPodsForWorkload(workload types.Workload) corev1sets.PodSet {
	podsForWorkload := corev1sets.NewPodSet()

	for _, pod := range d.pods.List() {
		if d.podIsOwnedOwnedByWorkload(pod, workload) {
			// this pod is owned by the workload in question
			podsForWorkload.Insert(pod)
		}
	}

	return podsForWorkload
}

func (d workloadDetector) podIsOwnedOwnedByWorkload(pod *corev1.Pod, workload types.Workload) bool {
	if pod.Namespace != workload.GetNamespace() || pod.ClusterName != workload.GetClusterName() {
		return false
	}

	// track the controlled object;
	// in the case of deployments
	var workloadReplica metav1.Object
	switch workload.Kind() {
	case deploymentKind:
		// pods created by deployments are owned by replicasets
		rsName := getControllerName(pod, replicaSetKind)
		if rsName == "" {
			return false
		}

		rsRef := &skv1.ClusterObjectRef{
			Name:        rsName,
			Namespace:   pod.Namespace,
			ClusterName: pod.ClusterName,
		}

		rs, err := d.replicaSets.Find(rsRef)
		if err != nil {
			contextutils.LoggerFrom(d.ctx).Warnw("replicaset not found for pod", "replicaset", sets.Key(rsRef), "pod", sets.Key(pod))
			return false
		}
		workloadReplica = rs
	default:
		// DaemonSets and StatefulSets directly
		// create pods
		workloadReplica = pod
	}

	workloadName := getControllerName(workloadReplica, workload.Kind())
	if workloadName == "" {
		// TODO(ilackarms): evaluate this assumption: currently, we
		// only consider pods owned by workloads to be part of a workload
		contextutils.LoggerFrom(d.ctx).Debugw("pod has no owner, ignoring for purposes of discovery", "pod", sets.Key(workloadReplica))
		return false
	}

	return workloadName == workload.GetName()
}

func getControllerName(obj metav1.Object, controllerKind string) string {
	for _, owner := range obj.GetOwnerReferences() {
		if owner.Controller != nil && *owner.Controller && controllerKind == owner.Kind {
			return owner.Name
		}
	}
	return ""
}
