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
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/utils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// the MeshWorkloadDetector detects MeshWorkloads from deployments whose pods are
// injected with a Mesh sidecar.
// in a k8s Deployment.
// If detection fails, an error is returned
// If no meshWorkload is detected, nil is returned
type MeshWorkloadDetector interface {
	DetectMeshWorkload(deployment *appsv1.Deployment) *v1alpha1.MeshWorkload
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

func (d meshWorkloadDetector) DetectMeshWorkload(deployment *appsv1.Deployment) *v1alpha1.MeshWorkload {
	podsForDeployment := d.getPodsForDeployment(deployment)

	mesh := d.getMeshForPods(podsForDeployment)

	if mesh == nil {
		return nil
	}

	meshRef := utils.MakeResourceRef(mesh)
	workloadRef := utils.MakeResourceRef(deployment)
	labels := deployment.Spec.Template.Labels
	serviceAccount := deployment.Spec.Template.Spec.ServiceAccountName

	return &v1alpha1.MeshWorkload{
		ObjectMeta: utils.DiscoveredObjectMeta(deployment),
		Spec:v1alpha1.MeshWorkloadSpec{
			KubeController:       &v1alpha1.MeshWorkloadSpec_KubeController{
				KubeControllerRef:    workloadRef,
				Labels:               labels,
				ServiceAccountName:   serviceAccount,
			},
			Mesh:                 meshRef,
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

func (d meshWorkloadDetector) getPodsForDeployment(deployment *appsv1.Deployment) corev1sets.PodSet {
	podsForDeployment := corev1sets.NewPodSet()

	for _, pod := range d.pods.List() {
		if pod.Namespace != deployment.Namespace || pod.ClusterName != deployment.ClusterName {
			continue
		}

		rsName := getControllerName(pod, replicaSetKind)
		if rsName == "" {
			// TODO(ilackarms): evaluate this assumption: currently, we
			// only consider pods owned by replicasets to be part of a workload
			contextutils.LoggerFrom(d.ctx).Debugw("pod has no owner, ignoring for purposes of discovery", "pod", sets.Key(pod))
			continue
		}

		rsRef := &v1.ClusterObjectRef{
			Name:        rsName,
			Namespace:   pod.Namespace,
			ClusterName: pod.ClusterName,
		}

		rs, err := d.replicaSets.Find(rsRef)
		if err != nil {
			contextutils.LoggerFrom(d.ctx).Warnw("cannot find replica set for pod", "pod", sets.Key(pod), "replicaset", sets.Key(rsRef))
			continue
		}

		deploymentName := getControllerName(rs, deploymentKind)
		if deploymentName == "" {
			// TODO(ilackarms): evaluate this assumption: currently, we
			// only consider pods owned by deployments to be part of a workload
			contextutils.LoggerFrom(d.ctx).Debugw("replicaset has no owner, ignoring for purposes of discovery", "rs", sets.Key(rs))
			continue
		}

		if deploymentName == deployment.Name {
			// this pod is owned by the deployment in question
			podsForDeployment.Insert(pod)
		}
	}

	return podsForDeployment
}

func getControllerName(obj metav1.Object, controllerKind string) string {
	for _, owner := range obj.GetOwnerReferences() {
		if owner.Controller != nil && *owner.Controller && controllerKind == owner.Kind {
			return owner.Name
		}
	}
	return ""
}
