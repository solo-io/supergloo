package istio_federation_test

import (
	"context"
	"fmt"

	types3 "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	istio_networking "github.com/solo-io/service-mesh-hub/pkg/clients/istio/networking"
	mock_istio_networking "github.com/solo-io/service-mesh-hub/pkg/clients/istio/networking/mock"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core/mocks"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mock_mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/mocks"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/dns"
	mock_dns "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/dns/mocks"
	istio_federation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/resolver/meshes/istio"
	mock_controller_runtime "github.com/solo-io/service-mesh-hub/test/mocks/controller-runtime"
	alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

			nonIstioMeshRef := &core_types.ResourceRef{
				Name:      "linkerd-mesh",
				Namespace: env.GetWriteNamespace(),
			}
			nonIstioMesh := &discovery_v1alpha1.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(nonIstioMeshRef),
				Spec: types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: "linkerd",
					},
					MeshType: &types.MeshSpec_Linkerd{},
				},
			}
			nonIstioMeshService := &discovery_v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{
					Name:      "linkerd-svc",
					Namespace: "application-ns",
				},
				Spec: types.MeshServiceSpec{
					Mesh: nonIstioMeshRef,
				},
			}
			virtualMesh := &networking_v1alpha1.VirtualMesh{
				Spec: types2.VirtualMeshSpec{
					Meshes: []*core_types.ResourceRef{nonIstioMeshRef},
				},
			}
			meshClient.EXPECT().
				Get(ctx, clients.ResourceRefToObjectKey(nonIstioMeshRef)).
				Return(nonIstioMesh, nil)
			clientGetter.EXPECT().
				GetClientForCluster("linkerd").
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
			istioMeshRef := &core_types.ResourceRef{
				Name:      "istio-mesh",
				Namespace: env.GetWriteNamespace(),
			}
			istioMesh := &discovery_v1alpha1.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(istioMeshRef),
				Spec: types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &types.MeshSpec_Istio{
						Istio: &types.MeshSpec_IstioMesh{
							Installation: &types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			}
			backingKubeService := &core_types.ResourceRef{
				Name:      "k8s-svc",
				Namespace: "application-ns",
			}
			istioMeshService := &discovery_v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: types.MeshServiceSpec{
					Mesh: istioMeshRef,
					Federation: &types.MeshServiceSpec_Federation{
						MulticlusterDnsName: dns.BuildMulticlusterDnsName(backingKubeService, clusterName),
					},
					KubeService: &types.MeshServiceSpec_KubeService{
						Ref: backingKubeService,
					},
				},
			}
			virtualMesh := &networking_v1alpha1.VirtualMesh{
				ObjectMeta: v1.ObjectMeta{
					Name:      "virtual-mesh-1",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: types2.VirtualMeshSpec{
					Meshes: []*core_types.ResourceRef{istioMeshRef},
				},
			}
			meshClient.EXPECT().
				Get(ctx, clients.ResourceRefToObjectKey(istioMeshRef)).
				Return(istioMesh, nil)
			clientGetter.EXPECT().
				GetClientForCluster(clusterName).
				Return(nil, nil)
			gatewayClient.EXPECT().
				Get(ctx, client.ObjectKey{
					Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
					Namespace: "istio-system",
				}).
				Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
			gatewayClient.EXPECT().
				Create(ctx, &v1alpha3.Gateway{
					ObjectMeta: v1.ObjectMeta{
						Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
						Namespace: "istio-system",
					},
					Spec: alpha3.Gateway{
						Servers: []*alpha3.Server{{
							Port: &alpha3.Port{
								Number:   istio_federation.DefaultGatewayPort,
								Protocol: istio_federation.DefaultGatewayProtocol,
								Name:     istio_federation.DefaultGatewayPortName,
							},
							Hosts: []string{
								// initially create the gateway with just the one service's host
								istio_federation.BuildMatchingMultiClusterHostName(istioMeshService.Spec.GetFederation()),
							},
							Tls: &alpha3.Server_TLSOptions{
								Mode: alpha3.Server_TLSOptions_AUTO_PASSTHROUGH,
							},
						}},
						Selector: istio_federation.BuildGatewayWorkloadSelector(),
					},
				}).
				Return(nil)

			envoyFilter := &v1alpha3.EnvoyFilter{
				ObjectMeta: v1.ObjectMeta{
					Name:      fmt.Sprintf("smh-%s-filter", virtualMesh.GetName()),
					Namespace: "istio-system",
				},
				Spec: alpha3.EnvoyFilter{
					ConfigPatches: []*alpha3.EnvoyFilter_EnvoyConfigObjectPatch{{
						ApplyTo: alpha3.EnvoyFilter_NETWORK_FILTER,
						Match: &alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
							Context: alpha3.EnvoyFilter_GATEWAY,
							ObjectTypes: &alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
								Listener: &alpha3.EnvoyFilter_ListenerMatch{
									PortNumber: istio_federation.DefaultGatewayPort,
									FilterChain: &alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
										Filter: &alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
											Name: istio_federation.EnvoySniClusterFilterName,
										},
									},
								},
							},
						},
						Patch: &alpha3.EnvoyFilter_Patch{
							Operation: alpha3.EnvoyFilter_Patch_INSERT_AFTER,
							Value:     mustBuildFilterPatch(clusterName),
						},
					}},
					WorkloadSelector: &alpha3.WorkloadSelector{
						Labels: istio_federation.BuildGatewayWorkloadSelector(),
					},
				},
			}
			envoyFilterClient.EXPECT().
				UpsertSpec(ctx, envoyFilter).
				Return(nil)

			var labels client.MatchingLabels = istio_federation.BuildGatewayWorkloadSelector()
			service := corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{{
						Name: istio_federation.DefaultGatewayPortName,
						Port: 3000,
					}},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{
							Hostname: "externally-resolvable-hostname.com",
						}},
					},
				},
			}
			serviceClient.EXPECT().
				List(ctx, labels).
				Return(&corev1.ServiceList{
					Items: []corev1.Service{service},
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
			istioMeshRef := &core_types.ResourceRef{
				Name:      "istio-mesh",
				Namespace: env.GetWriteNamespace(),
			}
			istioMesh := &discovery_v1alpha1.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(istioMeshRef),
				Spec: types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &types.MeshSpec_Istio{
						Istio: &types.MeshSpec_IstioMesh{
							Installation: &types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			}
			backingKubeService := &core_types.ResourceRef{
				Name:      "k8s-svc",
				Namespace: "application-ns",
			}
			istioMeshService := &discovery_v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: types.MeshServiceSpec{
					Mesh: istioMeshRef,
					Federation: &types.MeshServiceSpec_Federation{
						MulticlusterDnsName: dns.BuildMulticlusterDnsName(backingKubeService, "istio-cluster"),
					},
					KubeService: &types.MeshServiceSpec_KubeService{
						Ref: backingKubeService,
					},
				},
			}
			virtualMesh := &networking_v1alpha1.VirtualMesh{
				ObjectMeta: v1.ObjectMeta{
					Name:      "virtual-mesh-1",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: types2.VirtualMeshSpec{
					Meshes: []*core_types.ResourceRef{istioMeshRef},
				},
			}
			meshClient.EXPECT().
				Get(ctx, clients.ResourceRefToObjectKey(istioMeshRef)).
				Return(istioMesh, nil)
			clientGetter.EXPECT().
				GetClientForCluster(clusterName).
				Return(nil, nil)
			gateway := &v1alpha3.Gateway{
				ObjectMeta: v1.ObjectMeta{
					Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
					Namespace: "istio-system",
				},
				Spec: alpha3.Gateway{
					Servers: []*alpha3.Server{{
						Port: &alpha3.Port{
							Number:   istio_federation.DefaultGatewayPort,
							Protocol: istio_federation.DefaultGatewayProtocol,
							Name:     istio_federation.DefaultGatewayPortName,
						},
						Hosts: []string{
							// initially create the gateway with just the one service's host
							istioMeshService.Spec.GetFederation().GetMulticlusterDnsName(),
						},
						Tls: &alpha3.Server_TLSOptions{
							Mode: alpha3.Server_TLSOptions_AUTO_PASSTHROUGH,
						},
					}},
					Selector: istio_federation.BuildGatewayWorkloadSelector(),
				},
			}
			gatewayClient.EXPECT().
				Get(ctx, client.ObjectKey{
					Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
					Namespace: "istio-system",
				}).
				Return(gateway, nil)
			gatewayClient.EXPECT().
				Update(ctx, gateway).
				Return(nil)

			envoyFilter := &v1alpha3.EnvoyFilter{
				ObjectMeta: v1.ObjectMeta{
					Name:      fmt.Sprintf("smh-%s-filter", virtualMesh.GetName()),
					Namespace: "istio-system",
				},
				Spec: alpha3.EnvoyFilter{
					ConfigPatches: []*alpha3.EnvoyFilter_EnvoyConfigObjectPatch{{
						ApplyTo: alpha3.EnvoyFilter_NETWORK_FILTER,
						Match: &alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
							Context: alpha3.EnvoyFilter_GATEWAY,
							ObjectTypes: &alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
								Listener: &alpha3.EnvoyFilter_ListenerMatch{
									PortNumber: istio_federation.DefaultGatewayPort,
									FilterChain: &alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
										Filter: &alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
											Name: istio_federation.EnvoySniClusterFilterName,
										},
									},
								},
							},
						},
						Patch: &alpha3.EnvoyFilter_Patch{
							Operation: alpha3.EnvoyFilter_Patch_INSERT_AFTER,
							Value:     mustBuildFilterPatch(clusterName),
						},
					}},
					WorkloadSelector: &alpha3.WorkloadSelector{
						Labels: istio_federation.BuildGatewayWorkloadSelector(),
					},
				},
			}
			envoyFilterClient.EXPECT().
				UpsertSpec(ctx, envoyFilter).
				Return(nil)
			var labels client.MatchingLabels = istio_federation.BuildGatewayWorkloadSelector()
			service := corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{{
						Name: istio_federation.DefaultGatewayPortName,
						Port: 3000,
					}},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{
							Hostname: "externally-resolvable-hostname.com",
						}},
					},
				},
			}
			serviceClient.EXPECT().
				List(ctx, labels).
				Return(&corev1.ServiceList{
					Items: []corev1.Service{service},
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
			istioMeshRef := &core_types.ResourceRef{
				Name:      "istio-mesh",
				Namespace: env.GetWriteNamespace(),
			}
			istioMesh := &discovery_v1alpha1.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(istioMeshRef),
				Spec: types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &types.MeshSpec_Istio{
						Istio: &types.MeshSpec_IstioMesh{
							Installation: &types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			}
			backingKubeService := &core_types.ResourceRef{
				Name:      "k8s-svc",
				Namespace: "application-ns",
			}
			istioMeshService := &discovery_v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: types.MeshServiceSpec{
					Mesh: istioMeshRef,
					Federation: &types.MeshServiceSpec_Federation{
						MulticlusterDnsName: dns.BuildMulticlusterDnsName(backingKubeService, clusterName),
					},
					KubeService: &types.MeshServiceSpec_KubeService{
						Ref: backingKubeService,
					},
				},
			}
			virtualMesh := &networking_v1alpha1.VirtualMesh{
				ObjectMeta: v1.ObjectMeta{
					Name:      "virtual-mesh-1",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: types2.VirtualMeshSpec{
					Meshes: []*core_types.ResourceRef{istioMeshRef},
				},
			}
			meshClient.EXPECT().
				Get(ctx, clients.ResourceRefToObjectKey(istioMeshRef)).
				Return(istioMesh, nil)
			clientGetter.EXPECT().
				GetClientForCluster(clusterName).
				Return(nil, nil)
			gateway := &v1alpha3.Gateway{
				ObjectMeta: v1.ObjectMeta{
					Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
					Namespace: "istio-system",
				},
				Spec: alpha3.Gateway{
					Servers: []*alpha3.Server{{
						Port: &alpha3.Port{
							Number:   istio_federation.DefaultGatewayPort,
							Protocol: istio_federation.DefaultGatewayProtocol,
							Name:     istio_federation.DefaultGatewayPortName,
						},
						Hosts: []string{},
						Tls: &alpha3.Server_TLSOptions{
							Mode: alpha3.Server_TLSOptions_AUTO_PASSTHROUGH,
						},
					}},
					Selector: istio_federation.BuildGatewayWorkloadSelector(),
				},
			}
			gatewayClient.EXPECT().
				Get(ctx, client.ObjectKey{
					Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMesh.GetName()),
					Namespace: "istio-system",
				}).
				Return(gateway, nil)
			updatedGateway := *gateway
			updatedGateway.Spec.Servers = []*alpha3.Server{{
				Port: &alpha3.Port{
					Number:   istio_federation.DefaultGatewayPort,
					Protocol: istio_federation.DefaultGatewayProtocol,
					Name:     istio_federation.DefaultGatewayPortName,
				},
				Hosts: []string{istio_federation.BuildMatchingMultiClusterHostName(istioMeshService.Spec.GetFederation())},
				Tls: &alpha3.Server_TLSOptions{
					Mode: alpha3.Server_TLSOptions_AUTO_PASSTHROUGH,
				},
			}}
			gatewayClient.EXPECT().
				Update(ctx, &updatedGateway).
				Return(nil)

			envoyFilter := &v1alpha3.EnvoyFilter{
				ObjectMeta: v1.ObjectMeta{
					Name:      fmt.Sprintf("smh-%s-filter", virtualMesh.GetName()),
					Namespace: "istio-system",
				},
				Spec: alpha3.EnvoyFilter{
					ConfigPatches: []*alpha3.EnvoyFilter_EnvoyConfigObjectPatch{{
						ApplyTo: alpha3.EnvoyFilter_NETWORK_FILTER,
						Match: &alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
							Context: alpha3.EnvoyFilter_GATEWAY,
							ObjectTypes: &alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
								Listener: &alpha3.EnvoyFilter_ListenerMatch{
									PortNumber: istio_federation.DefaultGatewayPort,
									FilterChain: &alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
										Filter: &alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
											Name: istio_federation.EnvoySniClusterFilterName,
										},
									},
								},
							},
						},
						Patch: &alpha3.EnvoyFilter_Patch{
							Operation: alpha3.EnvoyFilter_Patch_INSERT_AFTER,
							Value:     mustBuildFilterPatch(clusterName),
						},
					}},
					WorkloadSelector: &alpha3.WorkloadSelector{
						Labels: istio_federation.BuildGatewayWorkloadSelector(),
					},
				},
			}
			envoyFilterClient.EXPECT().
				UpsertSpec(ctx, envoyFilter).
				Return(nil)
			var labels client.MatchingLabels = istio_federation.BuildGatewayWorkloadSelector()
			service := corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{{
						Name: istio_federation.DefaultGatewayPortName,
						Port: 3000,
					}},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{
							Hostname: "externally-resolvable-hostname.com",
						}},
					},
				},
			}
			serviceClient.EXPECT().
				List(ctx, labels).
				Return(&corev1.ServiceList{
					Items: []corev1.Service{service},
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

			nonIstioMeshRef := &core_types.ResourceRef{
				Name:      "linkerd-mesh",
				Namespace: env.GetWriteNamespace(),
			}
			nonIstioMesh := &discovery_v1alpha1.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(nonIstioMeshRef),
				Spec: types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: "linkerd",
					},
					MeshType: &types.MeshSpec_Linkerd{},
				},
			}
			nonIstioMeshWorkload := &discovery_v1alpha1.MeshWorkload{
				Spec: types.MeshWorkloadSpec{
					Mesh: nonIstioMeshRef,
				},
			}
			istioMeshService := &discovery_v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: types.MeshServiceSpec{},
			}
			meshClient.EXPECT().
				Get(ctx, clients.ResourceRefToObjectKey(nonIstioMeshRef)).
				Return(nonIstioMesh, nil)
			clientGetter.EXPECT().
				GetClientForCluster("linkerd").
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

			istioMeshRefService := &core_types.ResourceRef{
				Name:      "istio-mesh-1",
				Namespace: env.GetWriteNamespace(),
			}
			istioMeshRefWorkload := &core_types.ResourceRef{
				Name:      "istio-mesh-2",
				Namespace: env.GetWriteNamespace(),
			}
			istioMeshForService := &discovery_v1alpha1.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(istioMeshRefService),
				Spec: types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: "istio-cluster-svc",
					},
					MeshType: &types.MeshSpec_Istio{
						Istio: &types.MeshSpec_IstioMesh{
							Installation: &types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			}
			istioMeshForWorkload := &discovery_v1alpha1.Mesh{
				ObjectMeta: clients.ResourceRefToObjectMeta(istioMeshRefWorkload),
				Spec: types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: "istio-cluster-workload",
					},
					MeshType: &types.MeshSpec_Istio{
						Istio: &types.MeshSpec_IstioMesh{
							Installation: &types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			}
			meshWorkload := &discovery_v1alpha1.MeshWorkload{
				Spec: types.MeshWorkloadSpec{
					Mesh: istioMeshRefWorkload,
				},
			}
			backingKubeSvc := &core_types.ResourceRef{
				Name:      "application-svc",
				Namespace: "application-ns",
			}
			serviceMulticlusterDnsName := dns.BuildMulticlusterDnsName(backingKubeSvc, istioMeshForService.Spec.Cluster.Name)
			svcPort := &types.MeshServiceSpec_KubeService_KubeServicePort{
				Port:     9080,
				Name:     "http1",
				Protocol: "http",
			}
			meshService := &discovery_v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{
					Name:      "istio-svc",
					Namespace: "application-ns",
				},
				Spec: types.MeshServiceSpec{
					Mesh: istioMeshRefService,
					Federation: &types.MeshServiceSpec_Federation{
						MulticlusterDnsName: serviceMulticlusterDnsName,
					},
					KubeService: &types.MeshServiceSpec_KubeService{
						Ref: backingKubeSvc,
						Ports: []*types.MeshServiceSpec_KubeService_KubeServicePort{
							svcPort,
						},
					},
				},
			}
			meshClient.EXPECT().
				Get(ctx, clients.ResourceRefToObjectKey(istioMeshRefWorkload)).
				Return(istioMeshForWorkload, nil)
			workloadClient := mock_controller_runtime.NewMockClient(ctrl)
			clientGetter.EXPECT().
				GetClientForCluster("istio-cluster-workload").
				Return(workloadClient, nil)

			externalAddress := "externally-resolvable-hostname.com"
			port := uint32(32000)
			serviceEntryRef := &core_types.ResourceRef{
				Name:      serviceMulticlusterDnsName,
				Namespace: "istio-system",
			}
			serviceEntryClient.EXPECT().
				Get(ctx, clients.ResourceRefToObjectKey(serviceEntryRef)).
				Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
			ipAssigner.EXPECT().
				AssignIPOnCluster(ctx, istioMeshForWorkload.Spec.Cluster.Name).
				Return("255.255.255.255", nil)
			serviceEntry := &v1alpha3.ServiceEntry{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceEntryRef),
				Spec: alpha3.ServiceEntry{
					Addresses: []string{"255.255.255.255"},
					Endpoints: []*alpha3.ServiceEntry_Endpoint{{
						Address: externalAddress,
						Ports: map[string]uint32{
							svcPort.Name: port,
						},
					}},
					Hosts:    []string{serviceMulticlusterDnsName},
					Location: alpha3.ServiceEntry_MESH_INTERNAL,
					Ports: []*alpha3.Port{{
						Name:     svcPort.Name,
						Number:   svcPort.Port,
						Protocol: svcPort.Protocol,
					}},
					Resolution: alpha3.ServiceEntry_DNS,
				},
			}
			serviceEntryClient.EXPECT().
				Create(ctx, serviceEntry).
				Return(nil)
			destinationRuleRef := &core_types.ResourceRef{
				Name:      serviceMulticlusterDnsName,
				Namespace: "istio-system",
			}
			destinationRuleClient.EXPECT().
				Get(ctx, clients.ResourceRefToObjectKey(destinationRuleRef)).
				Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
			destinationRuleClient.EXPECT().Create(ctx, &v1alpha3.DestinationRule{
				ObjectMeta: clients.ResourceRefToObjectMeta(destinationRuleRef),
				Spec: alpha3.DestinationRule{
					Host: serviceMulticlusterDnsName,
					TrafficPolicy: &alpha3.TrafficPolicy{
						Tls: &alpha3.TLSSettings{
							// TODO this won't work with other mesh types https://github.com/solo-io/service-mesh-hub/issues/242
							Mode: alpha3.TLSSettings_ISTIO_MUTUAL,
						},
					},
				},
			}).
				Return(nil)

			eap := dns.ExternalAccessPoint{
				Address: externalAddress,
				Port:    port,
			}
			err := federationClient.FederateClientSide(ctx, eap, meshService, meshWorkload)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
