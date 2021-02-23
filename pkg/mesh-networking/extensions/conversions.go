package extensions

import (
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/extensions/v1alpha1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InputSnapshotToProto constructs a proto-compatible Discovery Snapshot from a networking input snapshot
func InputSnapshotToProto(in input.LocalSnapshot) *v1alpha1.DiscoverySnapshot {
	var meshes []*v1alpha1.MeshObject
	for _, mesh := range in.Meshes().List() {
		mesh := mesh
		meshes = append(meshes, &v1alpha1.MeshObject{
			Metadata: ObjectMetaToProto(mesh.ObjectMeta),
			Spec:     &mesh.Spec,
			Status:   &mesh.Status,
		})
	}
	var destinations []*v1alpha1.DestinationObject
	for _, destination := range in.Destinations().List() {
		destination := destination
		destinations = append(destinations, &v1alpha1.DestinationObject{
			Metadata: ObjectMetaToProto(destination.ObjectMeta),
			Spec:     &destination.Spec,
			Status:   &destination.Status,
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
		Meshes:       meshes,
		Destinations: destinations,
		Workloads:    workloads,
	}
}

// InputSnapshotFromProto constructs a Networking input snapshot from proto Discovery Snapshot
// This method is not intended to be used here, but called from implementing servers.
func InputSnapshotFromProto(name string, in *v1alpha1.DiscoverySnapshot) input.LocalSnapshot {
	builder := input.NewInputLocalSnapshotManualBuilder(name)

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

	// insert destinations
	var destinations discoveryv1alpha2.DestinationSlice
	for _, destination := range in.Destinations {
		destination := destination // pike
		destinations = append(destinations, &discoveryv1alpha2.Destination{
			ObjectMeta: ObjectMetaFromProto(destination.Metadata),
			Spec:       *destination.Spec,
			Status:     *destination.Status,
		})
	}
	builder.AddDestinations(destinations)

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
