package istio_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/meshworkload/detector/istio"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("IstioSidecarDetector", func() {
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
						Image: "istio-proxy:latest",
					},
				},
				ServiceAccountName: serviceAccountName,
			},
		}
	}

	istioMeshes := func(cluster string) v1alpha2sets.MeshSet {
		return v1alpha2sets.NewMeshSet(
			&v1alpha2.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "istio-system",
					Name:      "istio-cluster",
				},
				Spec: v1alpha2.MeshSpec{
					MeshType: &v1alpha2.MeshSpec_Istio_{
						Istio: &v1alpha2.MeshSpec_Istio{
							Installation: &v1alpha2.MeshSpec_MeshInstallation{
								Cluster: cluster,
							},
						},
					},
				},
			},
		)
	}

	detector := NewSidecarDetector(context.TODO())

	It("detects workload when sidecar mesh is in cluster", func() {
		pod := pod()

		meshes := istioMeshes(clusterName)

		meshWorkload := detector.DetectMeshSidecar(pod, meshes)
		Expect(meshWorkload).To(Equal(meshes.List()[0]))
	})
	It("does not detect workload when sidecar mesh is of different cluster", func() {
		pod := pod()

		meshes := istioMeshes("different-" + clusterName)

		meshWorkload := detector.DetectMeshSidecar(pod, meshes)
		Expect(meshWorkload).To(BeNil())
	})
	It("does not detect workload when sidecar mesh is not present", func() {
		pod := pod()

		meshes := v1alpha2sets.NewMeshSet()

		meshWorkload := detector.DetectMeshSidecar(pod, meshes)
		Expect(meshWorkload).To(BeNil())
	})
	It("does not detect workload when sidecar is not present", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   ns,
				Name:        podName,
				ClusterName: clusterName,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: "blah",
					},
				},
			},
		}

		meshes := istioMeshes(clusterName)

		meshWorkload := detector.DetectMeshSidecar(pod, meshes)
		Expect(meshWorkload).To(BeNil())
	})

})
