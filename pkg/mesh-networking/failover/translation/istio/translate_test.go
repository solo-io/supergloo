package istio_test

import (
	"context"
	"fmt"

	proto_types "github.com/gogo/protobuf/types"
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
	istio_networking "istio.io/api/networking/v1alpha3"
	istio_client_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Translate", func() {
	var (
		ctrl            *gomock.Controller
		ctx             context.Context
		mockIpAssigner  *mock_dns.MockIpAssigner
		istioTranslator translation.FailoverServiceTranslator
		failoverService = &v1alpha1.FailoverService{
			Spec: types.FailoverServiceSpec{
				Hostname:  "failoverservice.hostname",
				Namespace: "failoverservice-namespace",
				Port: &types.FailoverServiceSpec_Port{
					Port:     9080,
					Name:     "portname",
					Protocol: "tcp",
				},
				Cluster: "clustername",
			},
		}
		prioritizedMeshServices = []*v1alpha12.MeshService{
			{
				Spec: types2.MeshServiceSpec{
					KubeService: &types2.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Cluster: "cluster1",
						},
						Ports: []*types2.MeshServiceSpec_KubeService_KubeServicePort{
							{
								Port:     9080,
								Name:     "service1.port1",
								Protocol: "tcp",
							},
						},
					},
					Federation: &types2.MeshServiceSpec_Federation{
						MulticlusterDnsName: "service1.multiclusterdnsname",
					},
				},
			},
			{
				Spec: types2.MeshServiceSpec{
					KubeService: &types2.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Cluster: "cluster2",
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

	var protoStringValue = func(s string) *proto_types.Value {
		return &proto_types.Value{
			Kind: &proto_types.Value_StringValue{StringValue: s},
		}
	}

	var expectedServiceEntry = func(ip string) *istio_client_networking.ServiceEntry {
		return &istio_client_networking.ServiceEntry{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name:        failoverService.GetName(),
				Namespace:   failoverService.GetNamespace(),
				ClusterName: failoverService.Spec.GetCluster(),
			},
			Spec: istio_networking.ServiceEntry{
				Hosts: []string{failoverService.Spec.GetHostname()},
				Ports: []*istio_networking.Port{
					{
						Number:   failoverService.Spec.GetPort().GetPort(),
						Protocol: failoverService.Spec.GetPort().GetProtocol(),
						Name:     failoverService.Spec.GetPort().GetName(),
					},
				},
				Addresses: []string{ip},
				// Treat remote cluster services as part of the service mesh as all clusters in the service mesh share the same root of trust.
				Location:   istio_networking.ServiceEntry_MESH_INTERNAL,
				Resolution: istio_networking.ServiceEntry_DNS,
			},
		}
	}

	var expectedEnvoyFilter = func() *istio_client_networking.EnvoyFilter {
		failoverServiceClusterString := fmt.Sprintf("outbound|%d||%s",
			failoverService.Spec.GetPort().GetPort(),
			failoverService.Spec.GetHostname(),
		)
		return &istio_client_networking.EnvoyFilter{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name:        failoverService.GetName(),
				Namespace:   failoverService.GetNamespace(),
				ClusterName: failoverService.Spec.GetCluster(),
			},
			Spec: istio_networking.EnvoyFilter{
				ConfigPatches: []*istio_networking.EnvoyFilter_EnvoyConfigObjectPatch{
					// Replace the default Envoy configuration for Istio ServiceEntry with custom Envoy failover config
					{
						ApplyTo: istio_networking.EnvoyFilter_CLUSTER,
						Match: &istio_networking.EnvoyFilter_EnvoyConfigObjectMatch{
							Context: istio_networking.EnvoyFilter_ANY,
							ObjectTypes: &istio_networking.EnvoyFilter_EnvoyConfigObjectMatch_Cluster{
								Cluster: &istio_networking.EnvoyFilter_ClusterMatch{
									Name: failoverServiceClusterString,
								},
							},
						},
						Patch: &istio_networking.EnvoyFilter_Patch{
							Operation: istio_networking.EnvoyFilter_Patch_REMOVE,
						},
					},
					{
						ApplyTo: istio_networking.EnvoyFilter_CLUSTER,
						Match: &istio_networking.EnvoyFilter_EnvoyConfigObjectMatch{
							Context: istio_networking.EnvoyFilter_ANY,
							ObjectTypes: &istio_networking.EnvoyFilter_EnvoyConfigObjectMatch_Cluster{
								Cluster: &istio_networking.EnvoyFilter_ClusterMatch{
									Name: failoverServiceClusterString,
								},
							},
						},
						Patch: &istio_networking.EnvoyFilter_Patch{
							Operation: istio_networking.EnvoyFilter_Patch_ADD,
							Value: &proto_types.Struct{
								Fields: map[string]*proto_types.Value{
									"name":            protoStringValue(failoverService.Spec.GetHostname()),
									"connect_timeout": protoStringValue("1s"),
									"lb_policy":       protoStringValue("CLUSTER_PROVIDED"),
									"cluster_type": {
										Kind: &proto_types.Value_StructValue{
											StructValue: &proto_types.Struct{
												Fields: map[string]*proto_types.Value{
													"name": protoStringValue("envoy.clusters.aggregate"),
													"typed_config": {
														Kind: &proto_types.Value_StructValue{
															StructValue: &proto_types.Struct{
																Fields: map[string]*proto_types.Value{
																	"@type":    protoStringValue("type.googleapis.com/udpa.type.v1.TypedStruct"),
																	"type_url": protoStringValue("type.googleapis.com/envoy.config.cluster.aggregate.v2alpha.ClusterConfig"),
																	"value": {
																		Kind: &proto_types.Value_StructValue{
																			StructValue: &proto_types.Struct{
																				Fields: map[string]*proto_types.Value{
																					"clusters": {
																						Kind: &proto_types.Value_ListValue{ListValue: &proto_types.ListValue{
																							Values: []*proto_types.Value{
																								protoStringValue(fmt.Sprintf("outbound|%d||%s", prioritizedMeshServices[0].Spec.GetKubeService().GetPorts()[0].Port, prioritizedMeshServices[0].Spec.GetFederation().GetMulticlusterDnsName())),
																								protoStringValue(fmt.Sprintf("outbound|%d||%s", prioritizedMeshServices[1].Spec.GetKubeService().GetPorts()[0].Port, prioritizedMeshServices[1].Spec.GetFederation().GetMulticlusterDnsName())),
																							},
																						}},
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
	}

	It("should translate FailoverService to ServiceEntry and EnvoyFilter", func() {
		ip := "ip.string"
		mockIpAssigner.
			EXPECT().
			AssignIPOnCluster(ctx, failoverService.Spec.GetCluster()).
			Return(ip, nil)

		outputSnapshot, translatorError := istioTranslator.Translate(ctx, failoverService, prioritizedMeshServices)
		Expect(translatorError).To(BeNil())
		Expect(outputSnapshot.ServiceEntries).To(HaveLen(1))
		Expect(outputSnapshot.ServiceEntries[0]).To(Equal(expectedServiceEntry(ip)))
		Expect(outputSnapshot.EnvoyFilters).To(HaveLen(1))
		Expect(outputSnapshot.EnvoyFilters[0]).To(Equal(expectedEnvoyFilter()))
	})
})
