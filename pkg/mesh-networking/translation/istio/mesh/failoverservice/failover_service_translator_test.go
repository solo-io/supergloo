package failoverservice

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	mock_hostutils "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/validation/failoverservice"
	mock_validation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/validation/failoverservice/mocks"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	"istio.io/istio/pkg/util/protomarshal"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("FailoverServiceTranslator", func() {
	var (
		ctrl                      *gomock.Controller
		ctx                       context.Context
		mockClusterDomainRegistry *mock_hostutils.MockClusterDomainRegistry
		mockValidator             *mock_validation.MockFailoverServiceValidator
		mockReporter              *mock_reporting.MockReporter
		failoverServiceTranslator Translator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockClusterDomainRegistry = mock_hostutils.NewMockClusterDomainRegistry(ctrl)
		mockValidator = mock_validation.NewMockFailoverServiceValidator(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		failoverServiceTranslator = &translator{
			ctx:            ctx,
			validator:      mockValidator,
			clusterDomains: mockClusterDomainRegistry,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate a FailoverService into ServiceEntries and EnvoyFilters", func() {
		failoverService := &discoveryv1alpha2.MeshStatus_AppliedFailoverService{
			Ref: &v1.ObjectRef{
				Name:      "failover-service",
				Namespace: "failover-service-namespace",
			},
			Spec: &networkingv1alpha2.FailoverServiceSpec{
				Hostname: "failover-1.failover-namespace.cluster1",
				Port: &networkingv1alpha2.FailoverServiceSpec_Port{
					Number:   9080,
					Protocol: "http",
				},
				Meshes: []*v1.ObjectRef{
					{
						Name:      "mesh-1",
						Namespace: "mesh-namespace-1",
					},
					{
						Name:      "mesh-2",
						Namespace: "mesh-namespace-2",
					},
				},
				BackingServices: []*networkingv1alpha2.FailoverServiceSpec_BackingService{
					{
						BackingServiceType: &networkingv1alpha2.FailoverServiceSpec_BackingService_KubeService{
							KubeService: &v1.ClusterObjectRef{
								Name:        "service-name-1",
								Namespace:   "service-namespace-1",
								ClusterName: "cluster-1",
							},
						},
					},
					{
						BackingServiceType: &networkingv1alpha2.FailoverServiceSpec_BackingService_KubeService{
							KubeService: &v1.ClusterObjectRef{
								Name:        "service-name-2",
								Namespace:   "service-namespace-2",
								ClusterName: "cluster-2",
							},
						},
					},
				},
			},
		}
		allTrafficTargets := []*discoveryv1alpha2.TrafficTarget{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh-service-1",
					Namespace: "default",
				},
				Spec: discoveryv1alpha2.TrafficTargetSpec{
					Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
						KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "service-name-1",
								Namespace:   "service-namespace-1",
								ClusterName: "cluster-1",
							},
							Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "service1.port1",
									Protocol: "tcp",
								},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh-service-2",
					Namespace: "default",
				},
				Spec: discoveryv1alpha2.TrafficTargetSpec{
					Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
						KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "service-name-2",
								Namespace:   "service-namespace-2",
								ClusterName: "cluster-2",
							},
							Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "service2.port1",
									Protocol: "tcp",
								},
							},
						},
					},
				},
			},
		}
		allMeshes := []*discoveryv1alpha2.Mesh{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh-1",
					Namespace: "mesh-namespace-1",
				},
				Spec: discoveryv1alpha2.MeshSpec{
					MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
						Istio: &discoveryv1alpha2.MeshSpec_Istio{
							Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
								Cluster:   "cluster-1",
								Namespace: "istio-system",
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh-2",
					Namespace: "mesh-namespace-2",
				},
				Spec: discoveryv1alpha2.MeshSpec{
					MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
						Istio: &discoveryv1alpha2.MeshSpec_Istio{
							Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
								Cluster:   "cluster-2",
								Namespace: "istio-system",
							},
						},
					},
				},
			},
		}

		in := input.NewInputSnapshotManualBuilder("").
			AddTrafficTargets(allTrafficTargets).
			AddMeshes(allMeshes).
			AddKubernetesClusters([]*v1alpha1.KubernetesCluster{{ObjectMeta: metav1.ObjectMeta{Name: "kube-cluster"}}}).
			AddVirtualMeshes([]*networkingv1alpha2.VirtualMesh{{ObjectMeta: metav1.ObjectMeta{Name: "virtual-mesh"}}}).
			Build()

		mockValidator.EXPECT().Validate(failoverservice.Inputs{
			TrafficTargets: in.TrafficTargets(),
			KubeClusters:   in.KubernetesClusters(),
			Meshes:         in.Meshes(),
			VirtualMeshes:  in.VirtualMeshes(),
		}, failoverService.Spec).Return(nil)

		for _, trafficTarget := range allTrafficTargets {
			kubeService := trafficTarget.Spec.GetKubeService()
			mockClusterDomainRegistry.EXPECT().GetServiceLocalFQDN(kubeService.Ref).Return(trafficTarget.Name + "." + trafficTarget.Namespace)
			mockClusterDomainRegistry.EXPECT().GetServiceGlobalFQDN(kubeService.Ref).Return(trafficTarget.Name + "." + trafficTarget.Namespace + ".global")
		}

		outputs := istio.NewBuilder(context.TODO(), "")
		failoverServiceTranslator.Translate(
			in,
			&discoveryv1alpha2.Mesh{
				Spec: discoveryv1alpha2.MeshSpec{MeshType: &discoveryv1alpha2.MeshSpec_Istio_{Istio: &discoveryv1alpha2.MeshSpec_Istio{}}},
			},
			failoverService,
			outputs,
			mockReporter,
		)

		expectedEnvoyFilterYamls := []string{
			`configPatches:
- applyTo: CLUSTER
  match:
    cluster:
      name: outbound|9080||failover-1.failover-namespace.cluster1
  patch:
    operation: REMOVE
- applyTo: CLUSTER
  match:
    cluster:
      name: outbound|9080||failover-1.failover-namespace.cluster1
  patch:
    operation: ADD
    value:
      cluster_type:
        name: envoy.clusters.aggregate
        typed_config:
          '@type': type.googleapis.com/udpa.type.v1.TypedStruct
          type_url: type.googleapis.com/envoy.config.cluster.aggregate.v2alpha.ClusterConfig
          value:
            clusters:
            - outbound|9080||mesh-service-1.default
            - outbound|9080||mesh-service-2.default.global
      connect_timeout: 1s
      lb_policy: CLUSTER_PROVIDED
      name: outbound|9080||failover-1.failover-namespace.cluster1
`,
			`configPatches:
- applyTo: CLUSTER
  match:
    cluster:
      name: outbound|9080||failover-1.failover-namespace.cluster1
  patch:
    operation: REMOVE
- applyTo: CLUSTER
  match:
    cluster:
      name: outbound|9080||failover-1.failover-namespace.cluster1
  patch:
    operation: ADD
    value:
      cluster_type:
        name: envoy.clusters.aggregate
        typed_config:
          '@type': type.googleapis.com/udpa.type.v1.TypedStruct
          type_url: type.googleapis.com/envoy.config.cluster.aggregate.v2alpha.ClusterConfig
          value:
            clusters:
            - outbound|9080||mesh-service-1.default.global
            - outbound|9080||mesh-service-2.default
      connect_timeout: 1s
      lb_policy: CLUSTER_PROVIDED
      name: outbound|9080||failover-1.failover-namespace.cluster1
`,
		}
		expectedServiceEntryYamls := []string{
			`addresses:
- 247.153.147.39
hosts:
- failover-1.failover-namespace.cluster1
location: MESH_INTERNAL
ports:
- name: http
  number: 9080
  protocol: http
resolution: DNS
`,
			`addresses:
- 247.153.147.39
hosts:
- failover-1.failover-namespace.cluster1
location: MESH_INTERNAL
ports:
- name: http
  number: 9080
  protocol: http
resolution: DNS
`,
		}
		expectedEnvoyFilterObjectMetas := []metav1.ObjectMeta{
			{
				Name:        "failover-service",
				Namespace:   "istio-system",
				ClusterName: "cluster-1",
				Labels:      metautils.TranslatedObjectLabels(),
			},
			{
				Name:        "failover-service",
				Namespace:   "istio-system",
				ClusterName: "cluster-2",
				Labels:      metautils.TranslatedObjectLabels(),
			},
		}
		expectedServiceEntryObjectMetas := []metav1.ObjectMeta{
			{
				Name:        "failover-service",
				Namespace:   defaults.GetPodNamespace(),
				ClusterName: "cluster-1",
				Labels:      metautils.TranslatedObjectLabels(),
			},
			{
				Name:        "failover-service",
				Namespace:   defaults.GetPodNamespace(),
				ClusterName: "cluster-2",
				Labels:      metautils.TranslatedObjectLabels(),
			},
		}
		var envoyFilterYamls []string
		var envoyFilterObjectMetas []metav1.ObjectMeta
		var serviceEntryYamls []string
		var serviceEntryObjectMetas []metav1.ObjectMeta
		for _, envoyFilter := range outputs.GetEnvoyFilters().List() {
			envoyFilterYaml, err := protomarshal.ToYAML(&envoyFilter.Spec)
			Expect(err).ToNot(HaveOccurred())
			envoyFilterYamls = append(envoyFilterYamls, envoyFilterYaml)
			envoyFilterObjectMetas = append(envoyFilterObjectMetas, envoyFilter.ObjectMeta)
		}
		for _, serviceEntry := range outputs.GetServiceEntries().List() {
			serviceEntryYaml, err := protomarshal.ToYAML(&serviceEntry.Spec)
			Expect(err).ToNot(HaveOccurred())
			serviceEntryYamls = append(serviceEntryYamls, serviceEntryYaml)
			serviceEntryObjectMetas = append(serviceEntryObjectMetas, serviceEntry.ObjectMeta)
		}

		Expect(envoyFilterYamls).To(ConsistOf(expectedEnvoyFilterYamls))
		Expect(serviceEntryYamls).To(ConsistOf(expectedServiceEntryYamls))
		Expect(envoyFilterObjectMetas).To(ConsistOf(expectedEnvoyFilterObjectMetas))
		Expect(serviceEntryObjectMetas).To(ConsistOf(expectedServiceEntryObjectMetas))
	})
})
