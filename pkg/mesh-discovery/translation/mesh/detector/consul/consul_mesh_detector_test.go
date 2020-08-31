package consul_test

// TODO(EItanya): Uncomment to re-enable consul discovery
// Currently commented out because of dependency issues
//
// import (
// 	"fmt"
//
// 	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
// 	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
// 	consul "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector/consul"
// 	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/labelutils"
// 	appsv1 "k8s.io/api/apps/v1"
// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// )
//
// const (
// 	meshNs         = "namespace"
// 	deploymentName = "consul-server"
// 	consulVersion  = "latest"
// 	clusterName    = "cluster"
// )
//
// var _ = Describe("Consul Mesh Detector", func() {
// 	var (
// 		detector = consul.NewMeshDetector()
// 	)
//
// 	It("doesn't discover consul if it is not present", func() {
//
// 		deployment := &appsv1.Deployment{
// 			ObjectMeta: metav1.ObjectMeta{Namespace: "a", Name: "a"},
// 			Spec: appsv1.DeploymentSpec{
// 				Template: corev1.PodTemplateSpec{
// 					Spec: corev1.PodSpec{
// 						Containers: []corev1.Container{
// 							{
// 								Image: "test-image",
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}
//
// 		mesh, err := detector.DetectMesh(deployment)
// 		Expect(err).NotTo(HaveOccurred())
// 		Expect(mesh).To(BeNil())
// 	})
//
// 	It("can discover consul", func() {
//
// 		consulContainer := consulDeployment().Spec.Template.Spec.Containers[0]
// 		deployment := &appsv1.Deployment{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Namespace:   meshNs,
// 				Name:        deploymentName,
// 				ClusterName: clusterName,
// 			},
// 			Spec: appsv1.DeploymentSpec{
// 				Template: corev1.PodTemplateSpec{
// 					Spec: corev1.PodSpec{
// 						Containers: []corev1.Container{
// 							{
// 								Image: "test-image",
// 							},
// 							consulContainer,
// 						},
// 					},
// 				},
// 			},
// 		}
//
// 		mesh, err := detector.DetectMesh(deployment)
// 		Expect(err).NotTo(HaveOccurred())
// 		Expect(mesh).To(Equal(&v1alpha2.Mesh{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name:      "consul-server-namespace-cluster",
// 				Namespace: defaults.GetPodNamespace(),
// 				Labels:    labelutils.ClusterLabels(clusterName),
// 			},
// 			Spec: v1alpha2.MeshSpec{
// 				MeshType: &v1alpha2.MeshSpec_ConsulConnect{ConsulConnect: &v1alpha2.MeshSpec_ConsulConnectMesh{
// 					Installation: &v1alpha2.MeshSpec_MeshInstallation{
// 						Namespace: meshNs,
// 						Cluster:   clusterName,
// 						Version:   "latest",
// 					},
// 				}},
// 			},
// 		}))
// 	})
//
// })
//
// func consulDeployment() *appsv1.Deployment {
// 	return &appsv1.Deployment{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Namespace: meshNs,
// 			Name:      "consul-server",
// 		},
// 		Spec: appsv1.DeploymentSpec{
// 			Template: corev1.PodTemplateSpec{
// 				Spec: corev1.PodSpec{
// 					Containers: []corev1.Container{{
// 						Image: fmt.Sprintf("consul:%s", consulVersion),
// 						Command: []string{
// 							"/bin/sh",
// 							"-ec",
// 							`
// CONSUL_FULLNAME="consul-consul"
//
// exec /bin/consul agent \
//   -advertise="${POD_IP}" \
//   -bind=0.0.0.0 \
//   -bootstrap-expect=1 \
//   -client=0.0.0.0 \
//   -config-dir=/consul/config \
//   -datacenter=minidc \
//   -data-dir=/consul/data \
//   -domain=consul \
//   -hcl="connect { enabled = true }" \
//   -ui \
//   -retry-join=${CONSUL_FULLNAME}-server-0.${CONSUL_FULLNAME}-server.${NAMESPACE}.svc \
//   -server`,
// 						},
// 					}},
// 				},
// 			},
// 		},
// 	}
// }
