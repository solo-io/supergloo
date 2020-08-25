package osm_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/detector/osm"
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

	osmMeshes := func(cluster string) v1alpha2sets.MeshSet {
		return v1alpha2sets.NewMeshSet(
			&v1alpha2.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "service-mesh-hub",
					Name:      "osm-controller-osm-system-master-cluster",
				},
				Spec: v1alpha2.MeshSpec{
					MeshType: &v1alpha2.MeshSpec_Osm{
						Osm: &v1alpha2.MeshSpec_OSM{
							Installation: &v1alpha2.MeshSpec_MeshInstallation{
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

		meshes := v1alpha2sets.NewMeshSet()

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
