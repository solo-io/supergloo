package extensions

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/extensions/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InputSnapshotToProto constructs a proto-compatible Discovery Snapshot from a networking input snapshot
func InputSnapshotToProto(in input.Snapshot) *v1alpha1.DiscoverySnapshot {
	var meshes []*v1alpha1.MeshObject
	for _, mesh := range in.Meshes().List() {
		mesh := mesh
		meshes = append(meshes, &v1alpha1.MeshObject{
			Metadata: ObjectMetaToProto(mesh.ObjectMeta),
			Spec:     &mesh.Spec,
			Status:   &mesh.Status,
		})
	}
	var trafficTargets []*v1alpha1.TrafficTargetObject
	for _, trafficTarget := range in.TrafficTargets().List() {
		trafficTarget := trafficTarget
		trafficTargets = append(trafficTargets, &v1alpha1.TrafficTargetObject{
			Metadata: ObjectMetaToProto(trafficTarget.ObjectMeta),
			Spec:     &trafficTarget.Spec,
			Status:   &trafficTarget.Status,
		})
	}
	var workloads []*v1alpha1.WorkloadObject
	for _, workload := range in.Workloads().List() {
		workload := workload
		workloads = append(workloads, &v1alpha1.WorkloadObject{
			Metadata: ObjectMetaToProto(workload.ObjectMeta),
			Spec:     &workload.Spec,
			Status:   &workload.Status,
		})
	}
	return &v1alpha1.DiscoverySnapshot{
		Meshes:         meshes,
		TrafficTargets: trafficTargets,
		Workloads:      workloads,
	}
}

// ObjectMetaToProto constructs a proto-compatible version of a k8s ObjectMeta
func ObjectMetaToProto(meta metav1.ObjectMeta) *v1alpha1.ObjectMeta {
	return &v1alpha1.ObjectMeta{
		Name:        meta.Name,
		Namespace:   meta.Namespace,
		ClusterName: meta.ClusterName,
		Labels:      meta.Labels,
		Annotations: meta.Annotations,
	}
}

// ObjectMetaToProto constructs a k8s ObjectMeta from a proto-compatible version
func ObjectMetaFromProto(meta *v1alpha1.ObjectMeta) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        meta.GetName(),
		Namespace:   meta.GetNamespace(),
		ClusterName: meta.GetClusterName(),
		Labels:      meta.GetLabels(),
		Annotations: meta.GetAnnotations(),
	}
}
