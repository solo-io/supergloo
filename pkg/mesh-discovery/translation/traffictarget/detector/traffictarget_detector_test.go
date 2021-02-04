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
	"istio.io/api/label"
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

	makeWorkload := func(subset string) *v1alpha2.Workload {
		labels := map[string]string{
			"select": "me",
			"subset": subset,
		}
		return &v1alpha2.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-workload-" + subset,
				Namespace: serviceNs,
			},
			Spec: v1alpha2.WorkloadSpec{
				WorkloadType: &v1alpha2.WorkloadSpec_Kubernetes{
					Kubernetes: &v1alpha2.WorkloadSpec_KubernetesWorkload{
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

	It("translates a service with a backing workload to a traffic target", func() {
		endpoints := v1sets.NewEndpointsSet()
		pods := v1sets.NewPodSet()
		nodes := v1sets.NewNodeSet()
		workloads := v1alpha2sets.NewWorkloadSet(
			makeWorkload("v1"),
			makeWorkload("v2"),
		)
		meshes := v1alpha2sets.NewMeshSet()
		svc := makeService()

		detector := NewTrafficTargetDetector()

		trafficTarget := detector.DetectTrafficTarget(ctx, svc, endpoints, pods, nodes, workloads, meshes)

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

	It("translates a service with endpoints to a traffic target", func() {
		endpoints := v1sets.NewEndpointsSet(
			&corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:        serviceName,
					Namespace:   serviceNs,
					ClusterName: serviceCluster,
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{{IP: "1"}, {IP: "2"}},
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
		pods := v1sets.NewPodSet()
		nodes := v1sets.NewNodeSet()
		workloads := v1alpha2sets.NewWorkloadSet(
			makeWorkload("v1"),
			makeWorkload("v2"),
		)
		meshes := v1alpha2sets.NewMeshSet()
		svc := makeService()

		detector := NewTrafficTargetDetector()

		trafficTarget := detector.DetectTrafficTarget(ctx, svc, endpoints, pods, nodes, workloads, meshes)

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
						Endpoints: []*v1alpha2.TrafficTargetSpec_KubeService_EndpointsSubset{
							{
								LocalityIpAddresses: []*v1alpha2.TrafficTargetSpec_KubeService_EndpointsSubset_LocalityIp{
									{Ip: "1"},
									{Ip: "2"},
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

	It("translates a service with a discovery annotation to a traffic target", func() {
		endpoints := v1sets.NewEndpointsSet()
		workloads := v1alpha2sets.NewWorkloadSet()
		pods := v1sets.NewPodSet()
		nodes := v1sets.NewNodeSet()
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

		trafficTarget := detector.DetectTrafficTarget(ctx, svc, endpoints, pods, nodes, workloads, meshes)

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

	It("adds locality info to a traffic target", func() {
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
							{IP: "1", NodeName: pointer.StringPtr("node1")},
							{IP: "2", NodeName: pointer.StringPtr("node2")},
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
		pods := v1sets.NewPodSet(
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "pod1",
					Namespace:   serviceNs,
					ClusterName: serviceCluster,
					Labels:      selectorLabels,
				},
				Spec: corev1.PodSpec{
					NodeName: "node1",
				},
			},
		)
		nodes := v1sets.NewNodeSet(
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "node1",
					ClusterName: serviceCluster,
					Labels: map[string]string{
						corev1.LabelZoneRegionStable:        "region1",
						corev1.LabelZoneFailureDomainStable: "zone1",
						label.IstioSubZone:                  "subzone1",
					},
				},
			},
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "node2",
					ClusterName: serviceCluster,
					Labels: map[string]string{
						corev1.LabelZoneRegionStable:        "region1",
						corev1.LabelZoneFailureDomainStable: "zone2",
						label.IstioSubZone:                  "subzone2",
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

		trafficTarget := detector.DetectTrafficTarget(ctx, svc, endpoints, pods, nodes, workloads, meshes)

		Expect(trafficTarget).To(Equal(&v1alpha2.TrafficTarget{
			ObjectMeta: utils.DiscoveredObjectMeta(svc),
			Spec: v1alpha2.TrafficTargetSpec{
				Type: &v1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &v1alpha2.TrafficTargetSpec_KubeService{
						Region:                 "region1",
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
						Endpoints: []*v1alpha2.TrafficTargetSpec_KubeService_EndpointsSubset{
							{
								LocalityIpAddresses: []*v1alpha2.TrafficTargetSpec_KubeService_EndpointsSubset_LocalityIp{
									{
										Ip: "1",
										SubLocality: &v1alpha2.TrafficTargetSpec_KubeService_EndpointsSubset_SubLocality{
											Zone:    "zone1",
											Subzone: "subzone1",
										},
									},
									{
										Ip: "2",
										SubLocality: &v1alpha2.TrafficTargetSpec_KubeService_EndpointsSubset_SubLocality{
											Zone:    "zone2",
											Subzone: "subzone2",
										},
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

})
