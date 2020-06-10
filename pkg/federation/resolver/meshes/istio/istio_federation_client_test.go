package istio_federation_test

import (
	"context"
	"fmt"

	types3 "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	istio_networking "github.com/solo-io/service-mesh-hub/pkg/api/istio/networking/v1alpha3"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/federation/dns"
	mock_dns "github.com/solo-io/service-mesh-hub/pkg/federation/dns/mocks"
	istio_federation "github.com/solo-io/service-mesh-hub/pkg/federation/resolver/meshes/istio"
	mock_multicluster "github.com/solo-io/service-mesh-hub/pkg/kube/multicluster/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	mock_smh_discovery "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	mock_istio_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/istio/networking/v1beta1"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	mock_controller_runtime "github.com/solo-io/service-mesh-hub/test/mocks/controller-runtime"
	istio_networking_types "istio.io/api/networking/v1alpha3"
	istio_client_networking_types "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Istio Federation Decider", func() {
	var (
		ctrl                  *gomock.Controller
		ctx                   context.Context
		installationNamespace = "istio-system"

		mustBuildFilterPatch = func(clusterName string) *types3.Struct {
			val, err := istio_federation.BuildClusterReplacementPatch(clusterName)
			Expect(err).NotTo(HaveOccurred(), "Should be able to build the cluster replacement filter patch")
			return val
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("FederateServiceSide", func() {
		It("can resolve federation for a service belonging to an Istio mesh when no resources exist yet", func() {
			clientGetter := mock_multicluster.NewMockDynamicClientGetter(ctrl)
			meshClient := mock_smh_discovery.NewMockMeshClient(ctrl)
			ipAssigner := mock_dns.NewMockIpAssigner(ctrl)
			gatewayClient := mock_istio_networking.NewMockGatewayClient(ctrl)
			envoyFilterClient := mock_istio_networking.NewMockEnvoyFilterClient(ctrl)
			serviceClient := mock_kubernetes_core.NewMockServiceClient(ctrl)
			externalAccessPointGetter := mock_dns.NewMockExternalAccessPointGetter(ctrl)

			federationClient := istio_federation.NewIstioFederationClient(
				clientGetter,
				meshClient,
				func(_ client.Client) istio_networking.GatewayClient {
					return gatewayClient
				},
				func(_ client.Client) istio_networking.EnvoyFilterClient {
					return envoyFilterClient
				},
				func(_ client.Client) istio_networking.DestinationRuleClient {
					return nil
				},
				func(_ client.Client) istio_networking.ServiceEntryClient {
					return nil
				},
				func(_ client.Client) kubernetes_core.ServiceClient {
					return serviceClient
				},
				ipAssigner,
				externalAccessPointGetter,
			)

			clusterName := "istio-cluster"
			istioMeshRef := &smh_core_types.ResourceRef{
				Name:      "istio-mesh",
				Namespace: container_runtime.GetWriteNamespace(),
			}
			istioMesh := &smh_discovery.Mesh{
				ObjectMeta: selection.ResourceRefToObjectMeta(istioMeshRef),
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
						Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
							Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
								Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
									InstallationNamespace: installationNamespace,
								},
							},
						},
					},
				},
			}
			backingKubeService := &smh_core_types.ResourceRef{
				Name:      "k8s-svc",
				Namespace: "application-ns",
			}
			istioMeshService := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: istioMeshRef,
					Federation: &smh_discovery_types.MeshServiceSpec_Federation{
						MulticlusterDnsName: dns.BuildMulticlusterDnsName(backingKubeService, clusterName),
					},
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: backingKubeService,
					},
				},
			}
			virtualMesh := &smh_networking.VirtualMesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "virtual-mesh-1",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_networking_types.VirtualMeshSpec{
					Meshes: []*smh_core_types.ResourceRef{istioMeshRef},
				},
			}
			meshClient.EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(istioMeshRef)).
				Return(istioMesh, nil)
			clientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName).
				Return(nil, nil)
			gatewayClient.EXPECT().
				GetGateway(ctx, client.ObjectKey{
					Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
					Namespace: "istio-system",
				}).
				Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
			gatewayClient.EXPECT().
				CreateGateway(ctx, &istio_client_networking_types.Gateway{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
						Namespace: "istio-system",
					},
					Spec: istio_networking_types.Gateway{
						Servers: []*istio_networking_types.Server{{
							Port: &istio_networking_types.Port{
								Number:   istio_federation.DefaultGatewayPort,
								Protocol: istio_federation.DefaultGatewayProtocol,
								Name:     istio_federation.DefaultGatewayPortName,
							},
							Hosts: []string{
								// initially create the gateway with just the one service's host
								istio_federation.BuildMatchingMultiClusterHostName(istioMeshService.Spec.GetFederation()),
							},
							Tls: &istio_networking_types.Server_TLSOptions{
								Mode: istio_networking_types.Server_TLSOptions_AUTO_PASSTHROUGH,
							},
						}},
						Selector: istio_federation.BuildGatewayWorkloadSelector(),
					},
				}).
				Return(nil)

			envoyFilter := &istio_client_networking_types.EnvoyFilter{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      fmt.Sprintf("smh-%s-filter", virtualMesh.GetName()),
					Namespace: "istio-system",
				},
				Spec: istio_networking_types.EnvoyFilter{
					ConfigPatches: []*istio_networking_types.EnvoyFilter_EnvoyConfigObjectPatch{{
						ApplyTo: istio_networking_types.EnvoyFilter_NETWORK_FILTER,
						Match: &istio_networking_types.EnvoyFilter_EnvoyConfigObjectMatch{
							Context: istio_networking_types.EnvoyFilter_GATEWAY,
							ObjectTypes: &istio_networking_types.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
								Listener: &istio_networking_types.EnvoyFilter_ListenerMatch{
									PortNumber: istio_federation.DefaultGatewayPort,
									FilterChain: &istio_networking_types.EnvoyFilter_ListenerMatch_FilterChainMatch{
										Filter: &istio_networking_types.EnvoyFilter_ListenerMatch_FilterMatch{
											Name: istio_federation.EnvoySniClusterFilterName,
										},
									},
								},
							},
						},
						Patch: &istio_networking_types.EnvoyFilter_Patch{
							Operation: istio_networking_types.EnvoyFilter_Patch_INSERT_AFTER,
							Value:     mustBuildFilterPatch(clusterName),
						},
					}},
					WorkloadSelector: &istio_networking_types.WorkloadSelector{
						Labels: istio_federation.BuildGatewayWorkloadSelector(),
					},
				},
			}
			envoyFilterClient.EXPECT().
				UpsertEnvoyFilterSpec(ctx, envoyFilter).
				Return(nil)

			var labels client.MatchingLabels = istio_federation.BuildGatewayWorkloadSelector()
			service := k8s_core_types.Service{
				Spec: k8s_core_types.ServiceSpec{
					Ports: []k8s_core_types.ServicePort{{
						Name: istio_federation.DefaultGatewayPortName,
						Port: 3000,
					}},
				},
				Status: k8s_core_types.ServiceStatus{
					LoadBalancer: k8s_core_types.LoadBalancerStatus{
						Ingress: []k8s_core_types.LoadBalancerIngress{{
							Hostname: "externally-resolvable-hostname.com",
						}},
					},
				},
			}
			serviceClient.EXPECT().
				ListService(ctx, labels).
				Return(&k8s_core_types.ServiceList{
					Items: []k8s_core_types.Service{service},
				}, nil)

			externalAccessPointGetter.EXPECT().
				GetExternalAccessPointForService(ctx, &service, istio_federation.DefaultGatewayPortName, clusterName).
				Return(dns.ExternalAccessPoint{
					Address: "externally-resolvable-hostname.com",
					Port:    uint32(3000),
				}, nil)

			eap, err := federationClient.FederateServiceSide(ctx, installationNamespace, virtualMesh, istioMeshService)
			Expect(eap.Address).To(Equal("externally-resolvable-hostname.com"))
			Expect(eap.Port).To(Equal(uint32(3000)))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can resolve federation when the resources exist already and the service has already been federated to the gateway", func() {
			clientGetter := mock_multicluster.NewMockDynamicClientGetter(ctrl)
			meshClient := mock_smh_discovery.NewMockMeshClient(ctrl)
			ipAssigner := mock_dns.NewMockIpAssigner(ctrl)
			gatewayClient := mock_istio_networking.NewMockGatewayClient(ctrl)
			envoyFilterClient := mock_istio_networking.NewMockEnvoyFilterClient(ctrl)
			serviceClient := mock_kubernetes_core.NewMockServiceClient(ctrl)
			externalAccessPointGetter := mock_dns.NewMockExternalAccessPointGetter(ctrl)

			federationClient := istio_federation.NewIstioFederationClient(
				clientGetter,
				meshClient,
				func(_ client.Client) istio_networking.GatewayClient {
					return gatewayClient
				},
				func(_ client.Client) istio_networking.EnvoyFilterClient {
					return envoyFilterClient
				},
				func(_ client.Client) istio_networking.DestinationRuleClient {
					return nil
				},
				func(_ client.Client) istio_networking.ServiceEntryClient {
					return nil
				},
				func(_ client.Client) kubernetes_core.ServiceClient {
					return serviceClient
				},
				ipAssigner,
				externalAccessPointGetter,
			)

			clusterName := "istio-cluster"
			istioMeshRef := &smh_core_types.ResourceRef{
				Name:      "istio-mesh",
				Namespace: container_runtime.GetWriteNamespace(),
			}
			istioMesh := &smh_discovery.Mesh{
				ObjectMeta: selection.ResourceRefToObjectMeta(istioMeshRef),
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
						Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
							Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
								Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
									InstallationNamespace: installationNamespace,
								},
							},
						},
					},
				},
			}
			backingKubeService := &smh_core_types.ResourceRef{
				Name:      "k8s-svc",
				Namespace: "application-ns",
			}
			istioMeshService := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: istioMeshRef,
					Federation: &smh_discovery_types.MeshServiceSpec_Federation{
						MulticlusterDnsName: dns.BuildMulticlusterDnsName(backingKubeService, "istio-cluster"),
					},
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: backingKubeService,
					},
				},
			}
			virtualMesh := &smh_networking.VirtualMesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "virtual-mesh-1",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_networking_types.VirtualMeshSpec{
					Meshes: []*smh_core_types.ResourceRef{istioMeshRef},
				},
			}
			meshClient.EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(istioMeshRef)).
				Return(istioMesh, nil)
			clientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName).
				Return(nil, nil)
			gateway := &istio_client_networking_types.Gateway{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
					Namespace: "istio-system",
				},
				Spec: istio_networking_types.Gateway{
					Servers: []*istio_networking_types.Server{{
						Port: &istio_networking_types.Port{
							Number:   istio_federation.DefaultGatewayPort,
							Protocol: istio_federation.DefaultGatewayProtocol,
							Name:     istio_federation.DefaultGatewayPortName,
						},
						Hosts: []string{
							// initially create the gateway with just the one service's host
							istioMeshService.Spec.GetFederation().GetMulticlusterDnsName(),
						},
						Tls: &istio_networking_types.Server_TLSOptions{
							Mode: istio_networking_types.Server_TLSOptions_AUTO_PASSTHROUGH,
						},
					}},
					Selector: istio_federation.BuildGatewayWorkloadSelector(),
				},
			}
			gatewayClient.EXPECT().
				GetGateway(ctx, client.ObjectKey{
					Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
					Namespace: "istio-system",
				}).
				Return(gateway, nil)
			gatewayClient.EXPECT().
				UpdateGateway(ctx, gateway).
				Return(nil)

			envoyFilter := &istio_client_networking_types.EnvoyFilter{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      fmt.Sprintf("smh-%s-filter", virtualMesh.GetName()),
					Namespace: "istio-system",
				},
				Spec: istio_networking_types.EnvoyFilter{
					ConfigPatches: []*istio_networking_types.EnvoyFilter_EnvoyConfigObjectPatch{{
						ApplyTo: istio_networking_types.EnvoyFilter_NETWORK_FILTER,
						Match: &istio_networking_types.EnvoyFilter_EnvoyConfigObjectMatch{
							Context: istio_networking_types.EnvoyFilter_GATEWAY,
							ObjectTypes: &istio_networking_types.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
								Listener: &istio_networking_types.EnvoyFilter_ListenerMatch{
									PortNumber: istio_federation.DefaultGatewayPort,
									FilterChain: &istio_networking_types.EnvoyFilter_ListenerMatch_FilterChainMatch{
										Filter: &istio_networking_types.EnvoyFilter_ListenerMatch_FilterMatch{
											Name: istio_federation.EnvoySniClusterFilterName,
										},
									},
								},
							},
						},
						Patch: &istio_networking_types.EnvoyFilter_Patch{
							Operation: istio_networking_types.EnvoyFilter_Patch_INSERT_AFTER,
							Value:     mustBuildFilterPatch(clusterName),
						},
					}},
					WorkloadSelector: &istio_networking_types.WorkloadSelector{
						Labels: istio_federation.BuildGatewayWorkloadSelector(),
					},
				},
			}
			envoyFilterClient.EXPECT().
				UpsertEnvoyFilterSpec(ctx, envoyFilter).
				Return(nil)
			var labels client.MatchingLabels = istio_federation.BuildGatewayWorkloadSelector()
			service := k8s_core_types.Service{
				Spec: k8s_core_types.ServiceSpec{
					Ports: []k8s_core_types.ServicePort{{
						Name: istio_federation.DefaultGatewayPortName,
						Port: 3000,
					}},
				},
				Status: k8s_core_types.ServiceStatus{
					LoadBalancer: k8s_core_types.LoadBalancerStatus{
						Ingress: []k8s_core_types.LoadBalancerIngress{{
							Hostname: "externally-resolvable-hostname.com",
						}},
					},
				},
			}
			serviceClient.EXPECT().
				ListService(ctx, labels).
				Return(&k8s_core_types.ServiceList{
					Items: []k8s_core_types.Service{service},
				}, nil)

			externalAccessPointGetter.EXPECT().
				GetExternalAccessPointForService(ctx, &service, istio_federation.DefaultGatewayPortName, clusterName).
				Return(dns.ExternalAccessPoint{
					Address: "externally-resolvable-hostname.com",
					Port:    uint32(3000),
				}, nil)

			eap, err := federationClient.FederateServiceSide(ctx, installationNamespace, virtualMesh, istioMeshService)
			Expect(eap.Address).To(Equal("externally-resolvable-hostname.com"))
			Expect(eap.Port).To(Equal(uint32(3000)))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can resolve federation when the resources exist already and the service has NOT already been federated to the gateway", func() {
			clientGetter := mock_multicluster.NewMockDynamicClientGetter(ctrl)
			meshClient := mock_smh_discovery.NewMockMeshClient(ctrl)
			ipAssigner := mock_dns.NewMockIpAssigner(ctrl)
			gatewayClient := mock_istio_networking.NewMockGatewayClient(ctrl)
			envoyFilterClient := mock_istio_networking.NewMockEnvoyFilterClient(ctrl)
			serviceClient := mock_kubernetes_core.NewMockServiceClient(ctrl)
			externalAccessPointGetter := mock_dns.NewMockExternalAccessPointGetter(ctrl)

			federationClient := istio_federation.NewIstioFederationClient(
				clientGetter,
				meshClient,
				func(_ client.Client) istio_networking.GatewayClient {
					return gatewayClient
				},
				func(_ client.Client) istio_networking.EnvoyFilterClient {
					return envoyFilterClient
				},
				func(_ client.Client) istio_networking.DestinationRuleClient {
					return nil
				},
				func(_ client.Client) istio_networking.ServiceEntryClient {
					return nil
				},
				func(_ client.Client) kubernetes_core.ServiceClient {
					return serviceClient
				},
				ipAssigner,
				externalAccessPointGetter,
			)

			clusterName := "istio-cluster"
			istioMeshRef := &smh_core_types.ResourceRef{
				Name:      "istio-mesh",
				Namespace: container_runtime.GetWriteNamespace(),
			}
			istioMesh := &smh_discovery.Mesh{
				ObjectMeta: selection.ResourceRefToObjectMeta(istioMeshRef),
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
						Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
							Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
								Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
									InstallationNamespace: installationNamespace,
								},
							},
						},
					},
				},
			}
			backingKubeService := &smh_core_types.ResourceRef{
				Name:      "k8s-svc",
				Namespace: "application-ns",
			}
			istioMeshService := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: istioMeshRef,
					Federation: &smh_discovery_types.MeshServiceSpec_Federation{
						MulticlusterDnsName: dns.BuildMulticlusterDnsName(backingKubeService, clusterName),
					},
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: backingKubeService,
					},
				},
			}
			virtualMesh := &smh_networking.VirtualMesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "virtual-mesh-1",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_networking_types.VirtualMeshSpec{
					Meshes: []*smh_core_types.ResourceRef{istioMeshRef},
				},
			}
			meshClient.EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(istioMeshRef)).
				Return(istioMesh, nil)
			clientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName).
				Return(nil, nil)
			gateway := &istio_client_networking_types.Gateway{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
					Namespace: "istio-system",
				},
				Spec: istio_networking_types.Gateway{
					Servers: []*istio_networking_types.Server{{
						Port: &istio_networking_types.Port{
							Number:   istio_federation.DefaultGatewayPort,
							Protocol: istio_federation.DefaultGatewayProtocol,
							Name:     istio_federation.DefaultGatewayPortName,
						},
						Hosts: []string{},
						Tls: &istio_networking_types.Server_TLSOptions{
							Mode: istio_networking_types.Server_TLSOptions_AUTO_PASSTHROUGH,
						},
					}},
					Selector: istio_federation.BuildGatewayWorkloadSelector(),
				},
			}
			gatewayClient.EXPECT().
				GetGateway(ctx, client.ObjectKey{
					Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
					Namespace: "istio-system",
				}).
				Return(gateway, nil)
			updatedGateway := *gateway
			updatedGateway.Spec.Servers = []*istio_networking_types.Server{{
				Port: &istio_networking_types.Port{
					Number:   istio_federation.DefaultGatewayPort,
					Protocol: istio_federation.DefaultGatewayProtocol,
					Name:     istio_federation.DefaultGatewayPortName,
				},
				Hosts: []string{istio_federation.BuildMatchingMultiClusterHostName(istioMeshService.Spec.GetFederation())},
				Tls: &istio_networking_types.Server_TLSOptions{
					Mode: istio_networking_types.Server_TLSOptions_AUTO_PASSTHROUGH,
				},
			}}
			gatewayClient.EXPECT().
				UpdateGateway(ctx, &updatedGateway).
				Return(nil)

			envoyFilter := &istio_client_networking_types.EnvoyFilter{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      fmt.Sprintf("smh-%s-filter", virtualMesh.GetName()),
					Namespace: "istio-system",
				},
				Spec: istio_networking_types.EnvoyFilter{
					ConfigPatches: []*istio_networking_types.EnvoyFilter_EnvoyConfigObjectPatch{{
						ApplyTo: istio_networking_types.EnvoyFilter_NETWORK_FILTER,
						Match: &istio_networking_types.EnvoyFilter_EnvoyConfigObjectMatch{
							Context: istio_networking_types.EnvoyFilter_GATEWAY,
							ObjectTypes: &istio_networking_types.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
								Listener: &istio_networking_types.EnvoyFilter_ListenerMatch{
									PortNumber: istio_federation.DefaultGatewayPort,
									FilterChain: &istio_networking_types.EnvoyFilter_ListenerMatch_FilterChainMatch{
										Filter: &istio_networking_types.EnvoyFilter_ListenerMatch_FilterMatch{
											Name: istio_federation.EnvoySniClusterFilterName,
										},
									},
								},
							},
						},
						Patch: &istio_networking_types.EnvoyFilter_Patch{
							Operation: istio_networking_types.EnvoyFilter_Patch_INSERT_AFTER,
							Value:     mustBuildFilterPatch(clusterName),
						},
					}},
					WorkloadSelector: &istio_networking_types.WorkloadSelector{
						Labels: istio_federation.BuildGatewayWorkloadSelector(),
					},
				},
			}
			envoyFilterClient.EXPECT().
				UpsertEnvoyFilterSpec(ctx, envoyFilter).
				Return(nil)
			var labels client.MatchingLabels = istio_federation.BuildGatewayWorkloadSelector()
			service := k8s_core_types.Service{
				Spec: k8s_core_types.ServiceSpec{
					Ports: []k8s_core_types.ServicePort{{
						Name: istio_federation.DefaultGatewayPortName,
						Port: 3000,
					}},
				},
				Status: k8s_core_types.ServiceStatus{
					LoadBalancer: k8s_core_types.LoadBalancerStatus{
						Ingress: []k8s_core_types.LoadBalancerIngress{{
							Hostname: "externally-resolvable-hostname.com",
						}},
					},
				},
			}
			serviceClient.EXPECT().
				ListService(ctx, labels).
				Return(&k8s_core_types.ServiceList{
					Items: []k8s_core_types.Service{service},
				}, nil)

			externalAccessPointGetter.EXPECT().
				GetExternalAccessPointForService(ctx, &service, istio_federation.DefaultGatewayPortName, clusterName).
				Return(dns.ExternalAccessPoint{
					Address: "externally-resolvable-hostname.com",
					Port:    uint32(3000),
				}, nil)

			eap, err := federationClient.FederateServiceSide(ctx, installationNamespace, virtualMesh, istioMeshService)
			Expect(eap.Address).To(Equal("externally-resolvable-hostname.com"))
			Expect(eap.Port).To(Equal(uint32(3000)))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("FederateClientSide", func() {
		It("can resolve federation on the client side", func() {
			clientGetter := mock_multicluster.NewMockDynamicClientGetter(ctrl)
			meshClient := mock_smh_discovery.NewMockMeshClient(ctrl)
			ipAssigner := mock_dns.NewMockIpAssigner(ctrl)
			serviceEntryClient := mock_istio_networking.NewMockServiceEntryClient(ctrl)
			destinationRuleClient := mock_istio_networking.NewMockDestinationRuleClient(ctrl)
			externalAccessPointGetter := mock_dns.NewMockExternalAccessPointGetter(ctrl)

			federationClient := istio_federation.NewIstioFederationClient(
				clientGetter,
				meshClient,
				func(_ client.Client) istio_networking.GatewayClient {
					return nil
				},
				func(_ client.Client) istio_networking.EnvoyFilterClient {
					return nil
				},
				func(_ client.Client) istio_networking.DestinationRuleClient {
					return destinationRuleClient
				},
				func(_ client.Client) istio_networking.ServiceEntryClient {
					return serviceEntryClient
				},
				func(_ client.Client) kubernetes_core.ServiceClient {
					return nil
				},
				ipAssigner,
				externalAccessPointGetter,
			)

			istioMeshRefService := &smh_core_types.ResourceRef{
				Name:      "istio-mesh-1",
				Namespace: container_runtime.GetWriteNamespace(),
			}
			istioMeshRefWorkload := &smh_core_types.ResourceRef{
				Name:      "istio-mesh-2",
				Namespace: container_runtime.GetWriteNamespace(),
			}
			istioMeshForService := &smh_discovery.Mesh{
				ObjectMeta: selection.ResourceRefToObjectMeta(istioMeshRefService),
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: "istio-cluster-svc",
					},
					MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
						Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
							Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
								Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
									InstallationNamespace: installationNamespace,
								},
							},
						},
					},
				},
			}
			istioMeshForWorkload := &smh_discovery.Mesh{
				ObjectMeta: selection.ResourceRefToObjectMeta(istioMeshRefWorkload),
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: "istio-cluster-workload",
					},
					MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
						Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
							Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
								Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
									InstallationNamespace: installationNamespace,
								},
							},
						},
					},
				},
			}
			meshWorkload := &smh_discovery.MeshWorkload{
				Spec: smh_discovery_types.MeshWorkloadSpec{
					Mesh: istioMeshRefWorkload,
				},
			}
			backingKubeSvc := &smh_core_types.ResourceRef{
				Name:      "application-svc",
				Namespace: "application-ns",
			}
			serviceMulticlusterDnsName := dns.BuildMulticlusterDnsName(backingKubeSvc, istioMeshForService.Spec.Cluster.Name)
			svcPort := &smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
				Port:     9080,
				Name:     "http1",
				Protocol: "http",
			}
			meshService := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: istioMeshRefService,
					Federation: &smh_discovery_types.MeshServiceSpec_Federation{
						MulticlusterDnsName: serviceMulticlusterDnsName,
					},
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: backingKubeSvc,
						Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
							svcPort,
						},
					},
				},
			}
			meshClient.EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(istioMeshRefWorkload)).
				Return(istioMeshForWorkload, nil)
			workloadClient := mock_controller_runtime.NewMockClient(ctrl)
			clientGetter.EXPECT().
				GetClientForCluster(ctx, "istio-cluster-workload").
				Return(workloadClient, nil)

			externalAddress := "externally-resolvable-hostname.com"
			port := uint32(32000)
			serviceEntryRef := &smh_core_types.ResourceRef{
				Name:      serviceMulticlusterDnsName,
				Namespace: "istio-system",
			}
			serviceEntryClient.EXPECT().
				GetServiceEntry(ctx, selection.ResourceRefToObjectKey(serviceEntryRef)).
				Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
			ipAssigner.EXPECT().
				AssignIPOnCluster(ctx, istioMeshForWorkload.Spec.Cluster.Name).
				Return("255.255.255.255", nil)
			serviceEntry := &istio_client_networking_types.ServiceEntry{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceEntryRef),
				Spec: istio_networking_types.ServiceEntry{
					Addresses: []string{"255.255.255.255"},
					Endpoints: []*istio_networking_types.ServiceEntry_Endpoint{{
						Address: externalAddress,
						Ports: map[string]uint32{
							svcPort.Name: port,
						},
					}},
					Hosts:    []string{serviceMulticlusterDnsName},
					Location: istio_networking_types.ServiceEntry_MESH_INTERNAL,
					Ports: []*istio_networking_types.Port{{
						Name:     svcPort.Name,
						Number:   svcPort.Port,
						Protocol: svcPort.Protocol,
					}},
					Resolution: istio_networking_types.ServiceEntry_DNS,
				},
			}
			serviceEntryClient.EXPECT().
				CreateServiceEntry(ctx, serviceEntry).
				Return(nil)
			destinationRuleRef := &smh_core_types.ResourceRef{
				Name:      serviceMulticlusterDnsName,
				Namespace: "istio-system",
			}
			destinationRuleClient.EXPECT().
				GetDestinationRule(ctx, selection.ResourceRefToObjectKey(destinationRuleRef)).
				Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
			destinationRuleClient.EXPECT().CreateDestinationRule(ctx, &istio_client_networking_types.DestinationRule{
				ObjectMeta: selection.ResourceRefToObjectMeta(destinationRuleRef),
				Spec: istio_networking_types.DestinationRule{
					Host: serviceMulticlusterDnsName,
					TrafficPolicy: &istio_networking_types.TrafficPolicy{
						Tls: &istio_networking_types.TLSSettings{
							// TODO this won't work with other mesh types https://github.com/solo-io/service-mesh-hub/issues/242
							Mode: istio_networking_types.TLSSettings_ISTIO_MUTUAL,
						},
					},
				},
			}).Return(nil)

			eap := dns.ExternalAccessPoint{
				Address: externalAddress,
				Port:    port,
			}
			err := federationClient.FederateClientSide(ctx, installationNamespace, eap, meshService, meshWorkload)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
