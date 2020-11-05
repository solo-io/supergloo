package extensions

import (
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
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

// InputSnapshotFromProto constructs a Networking input snapshot from proto Discovery Snapshot
// This method is not intended to be used here, but called from implementating servers.
func InputSnapshotFromProto(name string, in *v1alpha1.DiscoverySnapshot) input.Snapshot {
	builder := input.NewInputSnapshotManualBuilder(name)

	// insert meshes
	var meshes discoveryv1alpha2.MeshSlice
	for _, mesh := range in.Meshes {
		meshes = append(meshes, &discoveryv1alpha2.Mesh{
			ObjectMeta: ObjectMetaFromProto(mesh.Metadata),
			Spec:       *mesh.Spec,
			Status:     *mesh.Status,
		})
	}
	builder.AddMeshes(meshes)

	// insert trafficTargets
	var trafficTargets discoveryv1alpha2.TrafficTargetSlice
	for _, trafficTarget := range in.TrafficTargets {
		trafficTargets = append(trafficTargets, &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: ObjectMetaFromProto(trafficTarget.Metadata),
			Spec:       *trafficTarget.Spec,
			Status:     *trafficTarget.Status,
		})
	}
	builder.AddTrafficTargets(trafficTargets)

	// insert workloads
	var workloads discoveryv1alpha2.WorkloadSlice
	for _, workload := range in.Workloads {
		workloads = append(workloads, &discoveryv1alpha2.Workload{
			ObjectMeta: ObjectMetaFromProto(workload.Metadata),
			Spec:       *workload.Spec,
			Status:     *workload.Status,
		})
	}
	builder.AddWorkloads(workloads)

	return builder.Build()
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
