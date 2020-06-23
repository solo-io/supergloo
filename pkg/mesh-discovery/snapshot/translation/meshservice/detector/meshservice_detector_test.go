package detector_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	. "github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/meshservice/detector"
)

var _ = Describe("MeshserviceDetector", func() {

	serviceName := "name"
	serviceNs := "namespace"
	serviceCluster := "cluster"
	selectorLabels := map[string]string{"select": "me"}
	serviceLabels := map[string]string{"app": "coolapp"}

	deployment := &types.ResourceRef{
		Name:      "deployment",
		Namespace: serviceNs,
		Cluster:   serviceCluster,
	}
	mesh := &types.ResourceRef{
		Name:      "mesh",
		Namespace: "any",
	}

	makeWorkload := func(subset string) *v1alpha1.MeshWorkload {
		labels := map[string]string{
			"select": "me",
			"subset": subset,
		}
		return &v1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-workload-" + subset,
				Namespace: serviceNs,
			},
			Spec: v1alpha1.MeshWorkloadSpec{
				WorkloadType: &v1alpha1.MeshWorkloadSpec_Kubernetes{
					Kubernetes: &v1alpha1.MeshWorkloadSpec_KubernertesWorkload{
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

	It("translates a service with a backing meshworkload to a meshservice", func() {
		workloads := v1alpha1sets.NewMeshWorkloadSet(
			makeWorkload("v1"),
			makeWorkload("v2"),
		)
		svc := makeService()

		detector := NewMeshServiceDetector()

		meshService := detector.DetectMeshService(svc, workloads)

		Expect(meshService).To(Equal(&v1alpha1.MeshService{
			ObjectMeta: utils.DiscoveredObjectMeta(svc),
			Spec: v1alpha1.MeshServiceSpec{
				KubeService: &v1alpha1.MeshServiceSpec_KubeService{
					Ref:                    utils.MakeResourceRef(svc),
					WorkloadSelectorLabels: svc.Spec.Selector,
					Labels:                 svc.Labels,
					Ports: []*v1alpha1.MeshServiceSpec_KubeService_KubeServicePort{
						{
							Port:                 1234,
							Name:                 "port1",
							Protocol:             "TCP",
						},
						{
							Port:                 2345,
							Name:                 "port2",
							Protocol:             "UDP",
						},
					},
				},
				Mesh: mesh,
				Subsets: map[string]*v1alpha1.MeshServiceSpec_Subset{
					"subset": {
						Values: []string{"v1", "v2"},
					},
				},
			},
		}))
	})
})
