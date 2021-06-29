package osm_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	. "github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/workload/detector/osm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("OsmSidecarDetector", func() {
	serviceAccountName := "service-account-name"
	ns := "namespace"
	clusterName := "cluster"
	podName := "pod"

	pod := func() *corev1.Pod {
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   ns,
				Name:        podName,
				ClusterName: clusterName,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: "envoyproxy/envoy-alpine:latest",
					},
				},
				InitContainers: []corev1.Container{
					{
						Image: "openservicemesh/init:latest",
						Name:  "osm-init",
					},
				},
				ServiceAccountName: serviceAccountName,
			},
		}
	}

	osmMeshes := func(cluster string) v1sets.MeshSet {
		return v1sets.NewMeshSet(
			&v1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "gloo-mesh",
					Name:      "osm-controller-osm-system-master-cluster",
				},
				Spec: v1.MeshSpec{
					Type: &v1.MeshSpec_Osm{
						Osm: &v1.MeshSpec_OSM{
							Installation: &v1.MeshInstallation{
								Namespace: ns,
								Cluster:   cluster,
							},
						},
					},
				},
			},
		)
	}

	detector := NewSidecarDetector(context.Background())

	It("detects workload when sidecar mesh is in cluster", func() {
		pod := pod()

		meshes := osmMeshes(clusterName)

		workload := detector.DetectMeshSidecar(pod, meshes)
		Expect(workload).To(Equal(meshes.List()[0]))
	})

	It("does not detect workload when sidecar mesh is of different cluster", func() {
		pod := pod()

		meshes := osmMeshes("different-" + clusterName)

		workload := detector.DetectMeshSidecar(pod, meshes)
		Expect(workload).To(BeNil())
	})

	It("does not detect workload when sidecar mesh is not present", func() {
		pod := pod()

		meshes := v1sets.NewMeshSet()

		workload := detector.DetectMeshSidecar(pod, meshes)
		Expect(workload).To(BeNil())
	})

	It("does not detect workload when sidecar is not present", func() {
		pod := pod()
		pod.Spec.Containers = nil

		meshes := osmMeshes(clusterName)

		workload := detector.DetectMeshSidecar(pod, meshes)
		Expect(workload).To(BeNil())
	})

	It("does not detect workload when sidecar is not present", func() {
		pod := pod()
		pod.Spec.InitContainers = nil

		meshes := osmMeshes(clusterName)

		workload := detector.DetectMeshSidecar(pod, meshes)
		Expect(workload).To(BeNil())
	})

})
