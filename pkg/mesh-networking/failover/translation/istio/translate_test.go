package istio_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	v1alpha12 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/translation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/translation/istio"
	mock_dns "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/dns/mocks"
	"istio.io/istio/pkg/util/protomarshal"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Translate", func() {
	var (
		ctrl            *gomock.Controller
		ctx             context.Context
		mockIpAssigner  *mock_dns.MockIpAssigner
		istioTranslator translation.FailoverServiceTranslator
		failoverService = &v1alpha1.FailoverService{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name:      "failover-service-name",
				Namespace: "failover-service-namespace",
			},
			Spec: types.FailoverServiceSpec{
				Hostname:  "service1-name.service1-namespace.cluster1",
				Namespace: "service1-namespace",
				Port: &types.FailoverServiceSpec_Port{
					Port:     9080,
					Name:     "http1",
					Protocol: "http",
				},
				Cluster: "cluster1",
			},
		}
		prioritizedMeshServices = []*v1alpha12.MeshService{
			{
				Spec: types2.MeshServiceSpec{
					KubeService: &types2.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      "service2-name",
							Namespace: "service2-namespace",
							Cluster:   "cluster2",
						},
						Ports: []*types2.MeshServiceSpec_KubeService_KubeServicePort{
							{
								Port:     9080,
								Name:     "service2.port1",
								Protocol: "tcp",
							},
						},
					},
					Federation: &types2.MeshServiceSpec_Federation{
						MulticlusterDnsName: "service2.multiclusterdnsname",
					},
				},
			},
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockIpAssigner = mock_dns.NewMockIpAssigner(ctrl)
		istioTranslator = istio.NewIstioFailoverServiceTranslator(mockIpAssigner)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate FailoverService to ServiceEntry and EnvoyFilter", func() {
		ip := "ip.string"
		expectedServiceEntryString := fmt.Sprintf(`addresses:
- %s
hosts:
- service1-name.service1-namespace.cluster1
location: MESH_INTERNAL
ports:
- name: http1
  number: 9080
  protocol: http
resolution: DNS
`, ip)

		expectedEnvoyFilterYamlString := `configPatches:
- applyTo: CLUSTER
  match:
    cluster:
      name: outbound|9080||service1-name.service1-namespace.cluster1
  patch:
    operation: REMOVE
- applyTo: CLUSTER
  match:
    cluster:
      name: outbound|9080||service1-name.service1-namespace.cluster1
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
            - outbound|9080||service2.multiclusterdnsname
      connect_timeout: 1s
      lb_policy: CLUSTER_PROVIDED
      name: outbound|9080||service1-name.service1-namespace.cluster1
`
		mockIpAssigner.
			EXPECT().
			AssignIPOnCluster(ctx, failoverService.Spec.GetCluster()).
			Return(ip, nil)
		outputSnapshot, translatorError := istioTranslator.Translate(ctx, failoverService, prioritizedMeshServices)
		Expect(translatorError).To(BeNil())
		envoyFilter := outputSnapshot.EnvoyFilters.List()[0]
		// EnvoyFilter must be in the same namespace as workloads backing the target service.
		Expect(envoyFilter.GetNamespace()).To(Equal(failoverService.Spec.GetNamespace()))
		envoyFilterYaml, err := protomarshal.ToYAML(&envoyFilter.Spec)
		Expect(err).ToNot(HaveOccurred())
		Expect(envoyFilterYaml).To(Equal(expectedEnvoyFilterYamlString))
		serviceEntryYaml, err := protomarshal.ToYAML(&outputSnapshot.ServiceEntries.List()[0].Spec)
		Expect(err).ToNot(HaveOccurred())
		Expect(serviceEntryYaml).To(Equal(expectedServiceEntryString))
	})
})
