package detector_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/utils"
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	. "github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/traffictarget/detector"
)

var _ = Describe("TrafficTargetDetector", func() {

	var (
		ctx context.Context

		serviceName    = "name"
		serviceNs      = "namespace"
		serviceCluster = "cluster"
		selectorLabels = map[string]string{"select": "me"}
		serviceLabels  = map[string]string{"app": "coolapp"}

		mesh = &skv1.ObjectRef{
			Name:      "mesh",
			Namespace: "any",
		}
	)

	buildLabels := func(subset string) map[string]string {
		labels := map[string]string{
			"select": "me",
			"subset": subset,
		}
		return labels
	}

	buildDeployment := func(subset string) *skv1.ClusterObjectRef {
		return &skv1.ClusterObjectRef{
			Name:        "deployment-" + subset,
			Namespace:   serviceNs,
			ClusterName: serviceCluster,
		}
	}

	makeWorkload := func(subset string) *v1alpha2.Workload {
		labels := buildLabels(subset)
		return &v1alpha2.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-workload-" + subset,
				Namespace: serviceNs,
			},
			Spec: v1alpha2.WorkloadSpec{
				WorkloadType: &v1alpha2.WorkloadSpec_Kubernetes{
					Kubernetes: &v1alpha2.WorkloadSpec_KubernetesWorkload{
						Controller:         buildDeployment(subset),
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
						Protocol:    "TCP",
						AppProtocol: pointer.StringPtr("HTTP"),
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

	It("translates a service with a backing workload to a traffictarget", func() {
		endpoints := v1sets.NewEndpointsSet()
		workloads := v1alpha2sets.NewWorkloadSet(
			makeWorkload("v1"),
			makeWorkload("v2"),
		)
		meshes := v1alpha2sets.NewMeshSet()
		svc := makeService()

		detector := NewTrafficTargetDetector()

		trafficTarget := detector.DetectTrafficTarget(ctx, svc, workloads, meshes, endpoints)

		Expect(trafficTarget).To(Equal(&v1alpha2.TrafficTarget{
			ObjectMeta: utils.DiscoveredObjectMeta(svc),
			Spec: v1alpha2.TrafficTargetSpec{
				Type: &v1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &v1alpha2.TrafficTargetSpec_KubeService{
						Ref:                    ezkube.MakeClusterObjectRef(svc),
						WorkloadSelectorLabels: svc.Spec.Selector,
						Labels:                 svc.Labels,
						Ports: []*v1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
							{
								Port:        1234,
								Name:        "port1",
								Protocol:    "TCP",
								AppProtocol: "HTTP",
							},
							{
								Port:     2345,
								Name:     "port2",
								Protocol: "UDP",
							},
						},
						Subsets: map[string]*v1alpha2.TrafficTargetSpec_KubeService_Subset{
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

	It("translates a service with endpoints if the backing workload is in flat network virtual mesh", func() {
		endpoints := v1sets.NewEndpointsSet(
			&corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:        serviceName,
					Namespace:   serviceNs,
					ClusterName: serviceCluster,
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1",
								TargetRef: &corev1.ObjectReference{
									Name:      buildDeployment("v1").Name + "-819320",
									Namespace: serviceNs,
								},
							},
							{
								IP: "2",
								TargetRef: &corev1.ObjectReference{
									Name:      buildDeployment("v2").Name + "-332434",
									Namespace: serviceNs,
								},
							},
						},
						Ports: []corev1.EndpointPort{
							{
								Name:        "port1",
								Port:        7000,
								Protocol:    "TCP",
								AppProtocol: pointer.StringPtr("HTTP"),
							},
						},
					},
				},
			},
		)
		workloads := v1alpha2sets.NewWorkloadSet(
			makeWorkload("v1"),
			makeWorkload("v2"),
		)
		meshes := v1alpha2sets.NewMeshSet()
		svc := makeService()

		detector := NewTrafficTargetDetector()

		trafficTarget := detector.DetectTrafficTarget(ctx, svc, workloads, meshes, endpoints)

		Expect(trafficTarget).To(Equal(&v1alpha2.TrafficTarget{
			ObjectMeta: utils.DiscoveredObjectMeta(svc),
			Spec: v1alpha2.TrafficTargetSpec{
				Type: &v1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &v1alpha2.TrafficTargetSpec_KubeService{
						Ref:                    ezkube.MakeClusterObjectRef(svc),
						WorkloadSelectorLabels: svc.Spec.Selector,
						Labels:                 svc.Labels,
						Ports: []*v1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
							{
								Port:        1234,
								Name:        "port1",
								Protocol:    "TCP",
								AppProtocol: "HTTP",
							},
							{
								Port:     2345,
								Name:     "port2",
								Protocol: "UDP",
							},
						},
						Subsets: map[string]*v1alpha2.TrafficTargetSpec_KubeService_Subset{
							"subset": {
								Values: []string{"v1", "v2"},
							},
						},
						EndpointSubsets: []*v1alpha2.TrafficTargetSpec_KubeService_EndpointsSubset{
							{
								Endpoints: []*v1alpha2.TrafficTargetSpec_KubeService_EndpointsSubset_Endpoint{
									{
										IpAddress: "1",
										Labels:    buildLabels("v1"),
									},
									{
										IpAddress: "2",
										Labels:    buildLabels("v2"),
									},
								},
								Ports: []*v1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
									{
										Port:        7000,
										Name:        "port1",
										Protocol:    "TCP",
										AppProtocol: "HTTP",
									},
								},
							},
						},
					},
				},
				Mesh: mesh,
			},
		}))
	})

	It("translates a service with a discovery annotation to a trafficTarget", func() {
		endpoints := v1sets.NewEndpointsSet()
		workloads := v1alpha2sets.NewWorkloadSet()
		meshes := v1alpha2sets.NewMeshSet(&v1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hello",
				Namespace: defaults.GetPodNamespace(),
			},
			Spec: v1alpha2.MeshSpec{
				MeshType: &v1alpha2.MeshSpec_Osm{
					Osm: &v1alpha2.MeshSpec_OSM{
						Installation: &v1alpha2.MeshSpec_MeshInstallation{
							Cluster: serviceCluster,
						},
					},
				},
			},
		})
		svc := makeService()
		svc.Annotations = map[string]string{
			DiscoveryMeshAnnotation: "true",
		}

		detector := NewTrafficTargetDetector()

		trafficTarget := detector.DetectTrafficTarget(ctx, svc, workloads, meshes, endpoints)

		Expect(trafficTarget).To(Equal(&v1alpha2.TrafficTarget{
			ObjectMeta: utils.DiscoveredObjectMeta(svc),
			Spec: v1alpha2.TrafficTargetSpec{
				Type: &v1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &v1alpha2.TrafficTargetSpec_KubeService{
						Ref:                    ezkube.MakeClusterObjectRef(svc),
						WorkloadSelectorLabels: svc.Spec.Selector,
						Labels:                 svc.Labels,
						Ports: []*v1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
							{
								Port:        1234,
								Name:        "port1",
								Protocol:    "TCP",
								AppProtocol: "HTTP",
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
