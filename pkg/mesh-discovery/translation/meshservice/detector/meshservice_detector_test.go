package detector_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/utils"
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	. "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/meshservice/detector"
)

var _ = Describe("MeshserviceDetector", func() {

	var (
		ctx context.Context

		serviceName    = "name"
		serviceNs      = "namespace"
		serviceCluster = "cluster"
		selectorLabels = map[string]string{"select": "me"}
		serviceLabels  = map[string]string{"app": "coolapp"}

		deployment = &skv1.ClusterObjectRef{
			Name:        "deployment",
			Namespace:   serviceNs,
			ClusterName: serviceCluster,
		}
		mesh = &skv1.ObjectRef{
			Name:      "mesh",
			Namespace: "any",
		}
	)

	makeWorkload := func(subset string) *v1alpha2.MeshWorkload {
		labels := map[string]string{
			"select": "me",
			"subset": subset,
		}
		return &v1alpha2.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-workload-" + subset,
				Namespace: serviceNs,
			},
			Spec: v1alpha2.MeshWorkloadSpec{
				WorkloadType: &v1alpha2.MeshWorkloadSpec_Kubernetes{
					Kubernetes: &v1alpha2.MeshWorkloadSpec_KubernertesWorkload{
						Controller:         deployment,
						PodLabels:          labels,
						ServiceAccountName: "any",
					},
				},
				Mesh: mesh,
			},
		}
	}

	makeService := func() *corev1.Service {
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   serviceNs,
				ClusterName: serviceCluster,
				Name:        serviceName,
				Labels:      serviceLabels,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name: "port1",
						Port: 1234,
						TargetPort: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "http",
						},
						Protocol: "TCP",
					},
					{
						Name: "port2",
						Port: 2345,
						TargetPort: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "grpc",
						},
						Protocol: "UDP",
					},
				},
				Selector: selectorLabels,
			},
		}
	}

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("translates a service with a backing meshworkload to a meshservice", func() {
		workloads := v1alpha2sets.NewMeshWorkloadSet(
			makeWorkload("v1"),
			makeWorkload("v2"),
		)
		meshes := v1alpha2sets.NewMeshSet()
		svc := makeService()

		detector := NewMeshServiceDetector(ctx)

		meshService := detector.DetectMeshService(svc, workloads, meshes)

		Expect(meshService).To(Equal(&v1alpha2.MeshService{
			ObjectMeta: utils.DiscoveredObjectMeta(svc),
			Spec: v1alpha2.MeshServiceSpec{
				Type: &v1alpha2.MeshServiceSpec_KubeService_{
					KubeService: &v1alpha2.MeshServiceSpec_KubeService{
						Ref:                    ezkube.MakeClusterObjectRef(svc),
						WorkloadSelectorLabels: svc.Spec.Selector,
						Labels:                 svc.Labels,
						Ports: []*v1alpha2.MeshServiceSpec_KubeService_KubeServicePort{
							{
								Port:     1234,
								Name:     "port1",
								Protocol: "TCP",
							},
							{
								Port:     2345,
								Name:     "port2",
								Protocol: "UDP",
							},
						},
						Subsets: map[string]*v1alpha2.MeshServiceSpec_KubeService_Subset{
							"subset": {
								Values: []string{"v1", "v2"},
							},
						},
					},
				},
				Mesh: mesh,
			},
		}))
	})

	It("translates a service with a discovery annotation to a meshservice", func() {
		workloads := v1alpha2sets.NewMeshWorkloadSet()
		meshes := v1alpha2sets.NewMeshSet(&v1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hello",
				Namespace: defaults.GetPodNamespace(),
			},
		})
		svc := makeService()
		svc.Annotations = map[string]string{
			DiscoveryMeshNameAnnotation: "hello",
		}

		detector := NewMeshServiceDetector(ctx)

		meshService := detector.DetectMeshService(svc, workloads, meshes)

		Expect(meshService).To(Equal(&v1alpha2.MeshService{
			ObjectMeta: utils.DiscoveredObjectMeta(svc),
			Spec: v1alpha2.MeshServiceSpec{
				Type: &v1alpha2.MeshServiceSpec_KubeService_{
					KubeService: &v1alpha2.MeshServiceSpec_KubeService{
						Ref:                    ezkube.MakeClusterObjectRef(svc),
						WorkloadSelectorLabels: svc.Spec.Selector,
						Labels:                 svc.Labels,
						Ports: []*v1alpha2.MeshServiceSpec_KubeService_KubeServicePort{
							{
								Port:     1234,
								Name:     "port1",
								Protocol: "TCP",
							},
							{
								Port:     2345,
								Name:     "port2",
								Protocol: "UDP",
							},
						},
					},
				},
				Mesh: &skv1.ObjectRef{
					Name:      "hello",
					Namespace: defaults.GetPodNamespace(),
				},
			},
		}))
	})
})
