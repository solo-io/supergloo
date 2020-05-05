package istio_federation_test

import (
	"context"
	"fmt"

	types3 "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	istio_networking "github.com/solo-io/service-mesh-hub/pkg/api/istio/networking/v1alpha3"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mock_mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s/mocks"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/dns"
	mock_dns "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/dns/mocks"
	istio_federation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/resolver/meshes/istio"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
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
		ctrl *gomock.Controller
		ctx  context.Context

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
		It("errors if the service being federated is not Istio", func() {
			clientGetter := mock_mc_manager.NewMockDynamicClientGetter(ctrl)
			meshClient := mock_zephyr_discovery.NewMockMeshClient(ctrl)
			ipAssigner := mock_dns.NewMockIpAssigner(ctrl)
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
					return nil
				},
				func(_ client.Client) istio_networking.ServiceEntryClient {
					return nil
				},
				func(_ client.Client) kubernetes_core.ServiceClient {
					return nil
				},
				ipAssigner,
				externalAccessPointGetter,
			)

			nonIstioMeshRef := &zephyr_core_types.ResourceRef{
				Name:      "linkerd-mesh",
				Namespace: env.GetWriteNamespace(),
			}
			nonIstioMesh := &zephyr_discovery.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(nonIstioMeshRef),
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: "linkerd",
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}
			nonIstioMeshService := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "linkerd-svc",
					Namespace: "application-ns",
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: nonIstioMeshRef,
				},
			}
			virtualMesh := &zephyr_networking.VirtualMesh{
				Spec: zephyr_networking_types.VirtualMeshSpec{
					Meshes: []*zephyr_core_types.ResourceRef{nonIstioMeshRef},
				},
			}
			meshClient.EXPECT().
				GetMesh(ctx, clients.ResourceRefToObjectKey(nonIstioMeshRef)).
				Return(nonIstioMesh, nil)
			clientGetter.EXPECT().
				GetClientForCluster(ctx, "linkerd").
				Return(nil, nil)

			_, err := federationClient.FederateServiceSide(ctx, virtualMesh, nonIstioMeshService)
			Expect(err).To(HaveOccurred())
			Expect(err).To(testutils.HaveInErrorChain(istio_federation.ServiceNotInIstio(nonIstioMeshService)))
		})

		It("can resolve federation for a service belonging to an Istio mesh when no resources exist yet", func() {
			clientGetter := mock_mc_manager.NewMockDynamicClientGetter(ctrl)
			meshClient := mock_zephyr_discovery.NewMockMeshClient(ctrl)
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
			istioMeshRef := &zephyr_core_types.ResourceRef{
				Name:      "istio-mesh",
				Namespace: env.GetWriteNamespace(),
			}
			istioMesh := &zephyr_discovery.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(istioMeshRef),
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio{
						Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{
							Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			}
			backingKubeService := &zephyr_core_types.ResourceRef{
				Name:      "k8s-svc",
				Namespace: "application-ns",
			}
			istioMeshService := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: istioMeshRef,
					Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{
						MulticlusterDnsName: dns.BuildMulticlusterDnsName(backingKubeService, clusterName),
					},
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: backingKubeService,
					},
				},
			}
			virtualMesh := &zephyr_networking.VirtualMesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "virtual-mesh-1",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_networking_types.VirtualMeshSpec{
					Meshes: []*zephyr_core_types.ResourceRef{istioMeshRef},
				},
			}
			meshClient.EXPECT().
				GetMesh(ctx, clients.ResourceRefToObjectKey(istioMeshRef)).
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

			eap, err := federationClient.FederateServiceSide(ctx, virtualMesh, istioMeshService)
			Expect(eap.Address).To(Equal("externally-resolvable-hostname.com"))
			Expect(eap.Port).To(Equal(uint32(3000)))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can resolve federation when the resources exist already and the service has already been federated to the gateway", func() {
			clientGetter := mock_mc_manager.NewMockDynamicClientGetter(ctrl)
			meshClient := mock_zephyr_discovery.NewMockMeshClient(ctrl)
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
			istioMeshRef := &zephyr_core_types.ResourceRef{
				Name:      "istio-mesh",
				Namespace: env.GetWriteNamespace(),
			}
			istioMesh := &zephyr_discovery.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(istioMeshRef),
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio{
						Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{
							Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			}
			backingKubeService := &zephyr_core_types.ResourceRef{
				Name:      "k8s-svc",
				Namespace: "application-ns",
			}
			istioMeshService := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: istioMeshRef,
					Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{
						MulticlusterDnsName: dns.BuildMulticlusterDnsName(backingKubeService, "istio-cluster"),
					},
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: backingKubeService,
					},
				},
			}
			virtualMesh := &zephyr_networking.VirtualMesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "virtual-mesh-1",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_networking_types.VirtualMeshSpec{
					Meshes: []*zephyr_core_types.ResourceRef{istioMeshRef},
				},
			}
			meshClient.EXPECT().
				GetMesh(ctx, clients.ResourceRefToObjectKey(istioMeshRef)).
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

			eap, err := federationClient.FederateServiceSide(ctx, virtualMesh, istioMeshService)
			Expect(eap.Address).To(Equal("externally-resolvable-hostname.com"))
			Expect(eap.Port).To(Equal(uint32(3000)))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can resolve federation when the resources exist already and the service has NOT already been federated to the gateway", func() {
			clientGetter := mock_mc_manager.NewMockDynamicClientGetter(ctrl)
			meshClient := mock_zephyr_discovery.NewMockMeshClient(ctrl)
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
			istioMeshRef := &zephyr_core_types.ResourceRef{
				Name:      "istio-mesh",
				Namespace: env.GetWriteNamespace(),
			}
			istioMesh := &zephyr_discovery.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(istioMeshRef),
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio{
						Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{
							Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			}
			backingKubeService := &zephyr_core_types.ResourceRef{
				Name:      "k8s-svc",
				Namespace: "application-ns",
			}
			istioMeshService := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: istioMeshRef,
					Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{
						MulticlusterDnsName: dns.BuildMulticlusterDnsName(backingKubeService, clusterName),
					},
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: backingKubeService,
					},
				},
			}
			virtualMesh := &zephyr_networking.VirtualMesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "virtual-mesh-1",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_networking_types.VirtualMeshSpec{
					Meshes: []*zephyr_core_types.ResourceRef{istioMeshRef},
				},
			}
			meshClient.EXPECT().
				GetMesh(ctx, clients.ResourceRefToObjectKey(istioMeshRef)).
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

			eap, err := federationClient.FederateServiceSide(ctx, virtualMesh, istioMeshService)
			Expect(eap.Address).To(Equal("externally-resolvable-hostname.com"))
			Expect(eap.Port).To(Equal(uint32(3000)))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("FederateClientSide", func() {
		It("errors if the mesh workload does not belong to an Istio mesh", func() {
			clientGetter := mock_mc_manager.NewMockDynamicClientGetter(ctrl)
			meshClient := mock_zephyr_discovery.NewMockMeshClient(ctrl)
			ipAssigner := mock_dns.NewMockIpAssigner(ctrl)
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
					return nil
				},
				func(_ client.Client) istio_networking.ServiceEntryClient {
					return nil
				},
				func(_ client.Client) kubernetes_core.ServiceClient {
					return nil
				},
				ipAssigner,
				externalAccessPointGetter,
			)

			nonIstioMeshRef := &zephyr_core_types.ResourceRef{
				Name:      "linkerd-mesh",
				Namespace: env.GetWriteNamespace(),
			}
			nonIstioMesh := &zephyr_discovery.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(nonIstioMeshRef),
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: "linkerd",
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}
			nonIstioMeshWorkload := &zephyr_discovery.MeshWorkload{
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					Mesh: nonIstioMeshRef,
				},
			}
			istioMeshService := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{},
			}
			meshClient.EXPECT().
				GetMesh(ctx, clients.ResourceRefToObjectKey(nonIstioMeshRef)).
				Return(nonIstioMesh, nil)
			clientGetter.EXPECT().
				GetClientForCluster(ctx, "linkerd").
				Return(nil, nil)
			eap := dns.ExternalAccessPoint{
				Address: "abc.com",
				Port:    0,
			}
			err := federationClient.FederateClientSide(ctx, eap, istioMeshService, nonIstioMeshWorkload)
			Expect(err).To(testutils.HaveInErrorChain(istio_federation.WorkloadNotInIstio(nonIstioMeshWorkload)))
		})

		It("can resolve federation on the client side", func() {
			clientGetter := mock_mc_manager.NewMockDynamicClientGetter(ctrl)
			meshClient := mock_zephyr_discovery.NewMockMeshClient(ctrl)
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

			istioMeshRefService := &zephyr_core_types.ResourceRef{
				Name:      "istio-mesh-1",
				Namespace: env.GetWriteNamespace(),
			}
			istioMeshRefWorkload := &zephyr_core_types.ResourceRef{
				Name:      "istio-mesh-2",
				Namespace: env.GetWriteNamespace(),
			}
			istioMeshForService := &zephyr_discovery.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(istioMeshRefService),
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: "istio-cluster-svc",
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio{
						Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{
							Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			}
			istioMeshForWorkload := &zephyr_discovery.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(istioMeshRefWorkload),
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: "istio-cluster-workload",
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio{
						Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{
							Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			}
			meshWorkload := &zephyr_discovery.MeshWorkload{
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					Mesh: istioMeshRefWorkload,
				},
			}
			backingKubeSvc := &zephyr_core_types.ResourceRef{
				Name:      "application-svc",
				Namespace: "application-ns",
			}
			serviceMulticlusterDnsName := dns.BuildMulticlusterDnsName(backingKubeSvc, istioMeshForService.Spec.Cluster.Name)
			svcPort := &zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
				Port:     9080,
				Name:     "http1",
				Protocol: "http",
			}
			meshService := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: istioMeshRefService,
					Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{
						MulticlusterDnsName: serviceMulticlusterDnsName,
					},
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: backingKubeSvc,
						Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
							svcPort,
						},
					},
				},
			}
			meshClient.EXPECT().
				GetMesh(ctx, clients.ResourceRefToObjectKey(istioMeshRefWorkload)).
				Return(istioMeshForWorkload, nil)
			workloadClient := mock_controller_runtime.NewMockClient(ctrl)
			clientGetter.EXPECT().
				GetClientForCluster(ctx, "istio-cluster-workload").
				Return(workloadClient, nil)

			externalAddress := "externally-resolvable-hostname.com"
			port := uint32(32000)
			serviceEntryRef := &zephyr_core_types.ResourceRef{
				Name:      serviceMulticlusterDnsName,
				Namespace: "istio-system",
			}
			serviceEntryClient.EXPECT().
				GetServiceEntry(ctx, clients.ResourceRefToObjectKey(serviceEntryRef)).
				Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
			ipAssigner.EXPECT().
				AssignIPOnCluster(ctx, istioMeshForWorkload.Spec.Cluster.Name).
				Return("255.255.255.255", nil)
			serviceEntry := &istio_client_networking_types.ServiceEntry{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceEntryRef),
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
			destinationRuleRef := &zephyr_core_types.ResourceRef{
				Name:      serviceMulticlusterDnsName,
				Namespace: "istio-system",
			}
			destinationRuleClient.EXPECT().
				GetDestinationRule(ctx, clients.ResourceRefToObjectKey(destinationRuleRef)).
				Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
			destinationRuleClient.EXPECT().CreateDestinationRule(ctx, &istio_client_networking_types.DestinationRule{
				ObjectMeta: clients.ResourceRefToObjectMeta(destinationRuleRef),
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
			err := federationClient.FederateClientSide(ctx, eap, meshService, meshWorkload)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
