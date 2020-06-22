package detector

import (
	"context"
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/meshworkload/types"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// the MeshWorkloadDetector detects MeshWorkloads from workloads
// whose backing pods are injected with a Mesh sidecar.
// If no mesh is detected for the workload, nil is returned
type MeshWorkloadDetector interface {
	DetectMeshWorkload(workload types.Workload) *v1alpha1.MeshWorkload
}

const (
	replicaSetKind = "ReplicaSet"
	deploymentKind = "Deployment"
)

// detects a workload
type meshWorkloadDetector struct {
	ctx         context.Context
	pods        corev1sets.PodSet
	replicaSets appsv1sets.ReplicaSetSet
	meshes      v1alpha1sets.MeshSet
	detector    SidecarDetector
}

func NewMeshWorkloadDetector(
	ctx context.Context,
	pods corev1sets.PodSet,
	replicaSets appsv1sets.ReplicaSetSet,
	meshes v1alpha1sets.MeshSet,
	detector SidecarDetector,
) MeshWorkloadDetector {
	ctx = contextutils.WithLogger(ctx, "mesh-workload-detector")
	return &meshWorkloadDetector{
		ctx:         ctx,
		pods:        pods,
		replicaSets: replicaSets,
		meshes:      meshes,
		detector:    detector,
	}
}

func (d meshWorkloadDetector) DetectMeshWorkload(workload types.Workload) *v1alpha1.MeshWorkload {
	podsForWorkload := d.getPodsForWorkload(workload)

	mesh := d.getMeshForPods(podsForWorkload)

	if mesh == nil {
		return nil
	}

	meshRef := utils.MakeResourceRef(mesh)
	controllerRef := utils.MakeResourceRef(workload)
	labels := workload.GetPodTemplate().Labels
	serviceAccount := workload.GetPodTemplate().Spec.ServiceAccountName

	outputMeta := utils.DiscoveredObjectMeta(workload)
	// append workload kind for uniqueness
	outputMeta.Name += "-" + strings.ToLower(workload.Kind())

	return &v1alpha1.MeshWorkload{
		ObjectMeta: outputMeta,
		Spec: v1alpha1.MeshWorkloadSpec{
			WorkloadType: &v1alpha1.MeshWorkloadSpec_Kubernetes{
				Kubernetes: &v1alpha1.MeshWorkloadSpec_KubernertesWorkload{
					Controller:         controllerRef,
					PodLabels:          labels,
					ServiceAccountName: serviceAccount,
				},
			},
			Mesh: meshRef,
		},
	}
}

func (d meshWorkloadDetector) getMeshForPods(pods corev1sets.PodSet) *v1alpha1.Mesh {
	// as long as one pod is detected for a mesh, consider the set owned by that mesh.
	for _, pod := range pods.List() {
		if mesh := d.detector.DetectMeshSidecar(pod, d.meshes); mesh != nil {
			return mesh
		}
	}
	return nil
}

func (d meshWorkloadDetector) getPodsForWorkload(workload types.Workload) corev1sets.PodSet {
	podsForWorkload := corev1sets.NewPodSet()

	for _, pod := range d.pods.List() {
		if d.podIsOwnedOwnedByWorkload(pod, workload) {
			// this pod is owned by the workload in question
			podsForWorkload.Insert(pod)
		}
	}

	return podsForWorkload
}

func (d meshWorkloadDetector) podIsOwnedOwnedByWorkload(pod *corev1.Pod, workload types.Workload) bool {
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

		rsRef := &v1.ClusterObjectRef{
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
