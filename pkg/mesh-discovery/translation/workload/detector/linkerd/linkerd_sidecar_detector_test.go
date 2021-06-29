package linkerd_test

// TODO(EItanya): Uncomment to re-enable linkerd discovery
// Currently commented out because of dependency issues
//
// import (
// 	"context"
//
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
// 	v1sets  "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
// 	. "github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/workload/detector/linkerd"
// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )
//
// var _ = Describe("LinkerdSidecarDetector", func() {
// 	serviceAccountName := "service-account-name"
// 	ns := "namespace"
// 	clusterName := "cluster"
// 	podName := "pod"
//
// 	pod := func() *corev1.Pod {
// 		return &corev1.Pod{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Namespace:   ns,
// 				Name:        podName,
// 				ClusterName: clusterName,
// 			},
// 			Spec: corev1.PodSpec{
// 				Containers: []corev1.Container{
// 					{
// 						Name: "linkerd-proxy",
// 					},
// 				},
// 				ServiceAccountName: serviceAccountName,
// 			},
// 		}
// 	}
//
// 	linkerdMeshes := func(cluster string) v1sets.MeshSet {
// 		return v1sets.NewMeshSet(
// 			&v1alpha2.Mesh{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Namespace: "linkerd-system",
// 					Name:      "linkerd-cluster",
// 				},
// 				Spec: v1alpha2.MeshSpec{
// 					Type: &v1alpha2.MeshSpec_Linkerd{
// 						Linkerd: &v1alpha2.MeshSpec_LinkerdMesh{
// 							Installation: &v1alpha2.MeshInstallation{
// 								Cluster: cluster,
// 							},
// 						},
// 					},
// 				},
// 			},
// 		)
// 	}
//
// 	detector := NewSidecarDetector(context.TODO())
//
// 	It("detects workload when sidecar mesh is in cluster", func() {
// 		pod := pod()
//
// 		meshes := linkerdMeshes(clusterName)
//
// 		workload := detector.DetectMeshSidecar(pod, meshes)
// 		Expect(workload).To(Equal(meshes.List()[0]))
// 	})
// 	It("does not detect workload when sidecar mesh is of different cluster", func() {
// 		pod := pod()
//
// 		meshes := linkerdMeshes("different-" + clusterName)
//
// 		workload := detector.DetectMeshSidecar(pod, meshes)
// 		Expect(workload).To(BeNil())
// 	})
// 	It("does not detect workload when sidecar mesh is not present", func() {
// 		pod := pod()
//
// 		meshes := v1sets.NewMeshSet()
//
// 		workload := detector.DetectMeshSidecar(pod, meshes)
// 		Expect(workload).To(BeNil())
// 	})
// 	It("does not detect workload when sidecar is not present", func() {
// 		pod := &corev1.Pod{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Namespace:   ns,
// 				Name:        podName,
// 				ClusterName: clusterName,
// 			},
// 			Spec: corev1.PodSpec{
// 				Containers: []corev1.Container{
// 					{
// 						Image: "blah",
// 					},
// 				},
// 			},
// 		}
//
// 		meshes := linkerdMeshes(clusterName)
//
// 		workload := detector.DetectMeshSidecar(pod, meshes)
// 		Expect(workload).To(BeNil())
// 	})
//
// })
