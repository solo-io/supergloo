package linkerd_test

// TODO(EItanya): Uncomment to re-enable linkerd discovery
// Currently commented out because of dependency issues
//
// import (
// 	linkerdconfig "github.com/linkerd/linkerd2/controller/gen/config"
// 	"github.com/linkerd/linkerd2/pkg/config"
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
// 	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
// 	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
// 	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/labelutils"
// 	appsv1 "k8s.io/api/apps/v1"
// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//
// 	. "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector/linkerd"
// )
//
// var _ = Describe("LinkerdMeshDetector", func() {
// 	serviceAccountName := "service-account-name"
// 	meshNs := "namespace"
// 	clusterName := "cluster"
// 	clusterDomain := "cluster.domain"
// 	deploymentName := "linkerd"
//
// 	linkerdDeployment := func() *appsv1.Deployment {
// 		return &appsv1.Deployment{
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
// 								Image: "linkerd-io/controller:latest",
// 							},
// 						},
// 						ServiceAccountName: serviceAccountName,
// 					},
// 				},
// 			},
// 		}
// 	}
//
// 	linkerdConfigMap := func() corev1sets.ConfigMapSet {
// 		cfg := &linkerdconfig.All{
// 			Global: &linkerdconfig.Global{
// 				ClusterDomain: clusterDomain,
// 			},
// 			Proxy:   &linkerdconfig.Proxy{},
// 			Install: &linkerdconfig.Install{},
// 		}
// 		global, proxy, install, err := config.ToJSON(cfg)
// 		Expect(err).NotTo(HaveOccurred())
//
// 		return corev1sets.NewConfigMapSet(&corev1.ConfigMap{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name:        "linkerd-config",
// 				Namespace:   meshNs,
// 				ClusterName: clusterName,
// 			},
// 			Data: map[string]string{
// 				"global":  global,
// 				"proxy":   proxy,
// 				"install": install,
// 			},
// 		})
// 	}
//
// 	It("does not detect Linkerd when it is not there", func() {
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
// 		configMaps := corev1sets.NewConfigMapSet()
//
// 		detector := NewMeshDetector(configMaps)
//
// 		mesh, err := detector.DetectMesh(deployment)
// 		Expect(err).NotTo(HaveOccurred())
// 		Expect(mesh).To(BeNil())
// 	})
//
// 	It("detects a mesh from a deployment with the linkerd controller image", func() {
// 		configMaps := linkerdConfigMap()
// 		detector := NewMeshDetector(configMaps)
//
// 		deployment := linkerdDeployment()
// 		mesh, err := detector.DetectMesh(deployment)
// 		Expect(err).NotTo(HaveOccurred())
// 		Expect(mesh).To(Equal(&v1alpha2.Mesh{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name:      "linkerd-namespace-cluster",
// 				Namespace: defaults.GetPodNamespace(),
// 				Labels:    labelutils.ClusterLabels(clusterName),
// 			},
// 			Spec: v1alpha2.MeshSpec{
// 				MeshType: &v1alpha2.MeshSpec_Linkerd{Linkerd: &v1alpha2.MeshSpec_LinkerdMesh{
// 					Installation: &v1alpha2.MeshSpec_MeshInstallation{
// 						Namespace: meshNs,
// 						Cluster:   clusterName,
// 						Version:   "latest",
// 					},
// 					ClusterDomain: clusterDomain,
// 				}},
// 			},
// 		}))
// 	})
//
// })
