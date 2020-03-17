package istio_federation

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/types"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/clients"
	istio_networking "github.com/solo-io/mesh-projects/pkg/clients/istio/networking"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/pkg/proto_conversion"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/dns"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/resolver/meshes"
	alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/security/proto/envoy/config/filter/network/tcp_cluster_rewrite/v2alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ServiceNotInIstio = func(service *discovery_v1alpha1.MeshService) error {
		return eris.Errorf("Service %+v does not belong to an Istio mesh", service.ObjectMeta)
	}
	WorkloadNotInIstio = func(workload *discovery_v1alpha1.MeshWorkload) error {
		return eris.Errorf("Workload %+v does not belong to an Istio mesh", workload.ObjectMeta)
	}
	ClusterNotReady = func(clusterName string) error {
		return eris.Errorf("Cluster '%s' is not fully registered yet", clusterName)
	}
)

const (
	DefaultGatewayPort     = 15443 // https://istio.io/docs/ops/deployment/requirements/#ports-used-by-istio
	DefaultGatewayProtocol = "TLS"
	DefaultGatewayPortName = "tls"

	EnvoySniClusterFilterName = "envoy.filters.network.sni_cluster"

	ServiceEntryPort         = 9080 // seemingly arbitrary? port on the service entry in the local cluster that will get forwarded to the port of the remote gateway
	ServiceEntryPortName     = "http1"
	ServiceEntryPortProtocol = "http"
)

// istio-specific implementation of federation resolution
func NewIstioFederationClient(
	dynamicClientGetter mc_manager.DynamicClientGetter,
	meshClient discovery_core.MeshClient,
	gatewayClientFactory istio_networking.GatewayClientFactory,
	envoyFilterClientFactory istio_networking.EnvoyFilterClientFactory,
	destinationRuleClientFactory istio_networking.DestinationRuleClientFactory,
	serviceEntryClientFactory istio_networking.ServiceEntryClientFactory,
	serviceClientFactory kubernetes_core.ServiceClientFactory,
	ipAssigner dns.IpAssigner,
	externalAccessPointGetter dns.ExternalAccessPointGetter,
) meshes.MeshFederationClient {
	return &istioFederationClient{
		dynamicClientGetter:          dynamicClientGetter,
		meshClient:                   meshClient,
		gatewayClientFactory:         gatewayClientFactory,
		envoyFilterClientFactory:     envoyFilterClientFactory,
		destinationRuleClientFactory: destinationRuleClientFactory,
		serviceEntryClientFactory:    serviceEntryClientFactory,
		serviceClientFactory:         serviceClientFactory,
		ipAssigner:                   ipAssigner,
		externalAccessPointGetter:    externalAccessPointGetter,
	}
}

type istioFederationClient struct {
	dynamicClientGetter          mc_manager.DynamicClientGetter
	meshClient                   discovery_core.MeshClient
	gatewayClientFactory         istio_networking.GatewayClientFactory
	envoyFilterClientFactory     istio_networking.EnvoyFilterClientFactory
	destinationRuleClientFactory istio_networking.DestinationRuleClientFactory
	serviceEntryClientFactory    istio_networking.ServiceEntryClientFactory
	serviceClientFactory         kubernetes_core.ServiceClientFactory
	externalAccessPointGetter    dns.ExternalAccessPointGetter
	ipAssigner                   dns.IpAssigner
}

func (i *istioFederationClient) FederateServiceSide(
	ctx context.Context,
	meshGroup *networking_v1alpha1.MeshGroup,
	meshService *discovery_v1alpha1.MeshService,
) (eap dns.ExternalAccessPoint, err error) {
	meshForService, dynamicClient, err := i.getClientForMesh(ctx, meshService.Spec.GetMesh())
	if err != nil {
		return eap, err
	}

	if meshForService.Spec.GetIstio() == nil {
		return eap, ServiceNotInIstio(meshService)
	}

	installNamespace := meshForService.Spec.GetIstio().GetInstallation().GetInstallationNamespace()

	// Make sure the gateway is in a good state
	err = i.ensureGatewayExists(ctx, dynamicClient, meshGroup.GetName(), meshService, installNamespace)
	if err != nil {
		return eap, eris.Wrapf(err, "Failed to configure the ingress gateway for service %+v", meshService.ObjectMeta)
	}

	// ensure that the envoy filter exists
	err = i.ensureEnvoyFilterExists(ctx, meshGroup.GetName(), dynamicClient, installNamespace, meshForService.Spec.GetCluster().GetName())
	if err != nil {
		return eap, eris.Wrapf(err, "Failed to configure the ingress gateway envoy filter for service %+v", meshService.ObjectMeta)
	}

	// finally, send back the external IP for the gateway we just set up
	return i.determineExternalIpForGateway(ctx, meshGroup.GetName(), meshForService.Spec.GetCluster().GetName(), dynamicClient)
}

func (i *istioFederationClient) FederateClientSide(
	ctx context.Context,
	eap dns.ExternalAccessPoint,
	meshService *discovery_v1alpha1.MeshService,
	meshWorkload *discovery_v1alpha1.MeshWorkload,
) error {
	meshForWorkload, clientForWorkloadMesh, err := i.getClientForMesh(ctx, meshWorkload.Spec.GetMesh())
	if err != nil {
		return err
	}

	if meshForWorkload.Spec.GetIstio() == nil {
		return WorkloadNotInIstio(meshWorkload)
	}

	installNamespace := meshForWorkload.Spec.GetIstio().GetInstallation().GetInstallationNamespace()

	serviceMulticlusterName := meshService.Spec.GetFederation().GetMulticlusterDnsName()

	err = i.setUpServiceEntry(
		ctx,
		clientForWorkloadMesh,
		eap,
		installNamespace,
		serviceMulticlusterName,
		meshForWorkload.Spec.GetCluster().GetName(),
	)
	if err != nil {
		return err
	}

	return i.setUpDestinationRule(
		ctx,
		clientForWorkloadMesh,
		serviceMulticlusterName,
		installNamespace,
	)
}

func (i *istioFederationClient) setUpDestinationRule(
	ctx context.Context,
	clientForWorkloadMesh client.Client,
	serviceMulticlusterName string,
	installNamespace string,
) error {
	destinationRuleRef := &core_types.ResourceRef{
		Name:      serviceMulticlusterName,
		Namespace: installNamespace,
	}

	destinationRuleClient := i.destinationRuleClientFactory(clientForWorkloadMesh)
	_, err := destinationRuleClient.Get(ctx, clients.ResourceRefToObjectKey(destinationRuleRef))

	if errors.IsNotFound(err) {
		return destinationRuleClient.Create(ctx, &v1alpha3.DestinationRule{
			ObjectMeta: clients.ResourceRefToObjectMeta(destinationRuleRef),
			Spec: alpha3.DestinationRule{
				Host: serviceMulticlusterName,
				TrafficPolicy: &alpha3.TrafficPolicy{
					Tls: &alpha3.TLSSettings{
						// TODO this won't work with other mesh types https://github.com/solo-io/mesh-projects/issues/242
						Mode: alpha3.TLSSettings_ISTIO_MUTUAL,
					},
				},
			},
		})
	}
	return err
}

func (i *istioFederationClient) setUpServiceEntry(
	ctx context.Context,
	clientForWorkloadMesh client.Client,
	eap dns.ExternalAccessPoint,
	installNamespace,
	serviceMulticlusterDnsName,
	workloadClusterName string,
) error {
	serviceEntryClient := i.serviceEntryClientFactory(clientForWorkloadMesh)

	computedRef := &core_types.ResourceRef{
		Name:      serviceMulticlusterDnsName,
		Namespace: installNamespace,
	}

	_, err := serviceEntryClient.Get(ctx, clients.ResourceRefToObjectKey(computedRef))
	if errors.IsNotFound(err) {
		// generate a unique IP within the workload cluster for the service entry to point to
		newIp, err := i.ipAssigner.AssignIPOnCluster(ctx, workloadClusterName)
		if err != nil {
			return err
		}

		return serviceEntryClient.Create(ctx, &v1alpha3.ServiceEntry{
			ObjectMeta: clients.ResourceRefToObjectMeta(computedRef),
			Spec: alpha3.ServiceEntry{
				Addresses: []string{newIp},
				Endpoints: []*alpha3.ServiceEntry_Endpoint{{
					Address: eap.Address,
					Ports: map[string]uint32{
						ServiceEntryPortName: eap.Port,
					},
				}},
				Hosts:    []string{serviceMulticlusterDnsName},
				Location: alpha3.ServiceEntry_MESH_INTERNAL,
				Ports: []*alpha3.Port{{
					Name:     ServiceEntryPortName,
					Number:   ServiceEntryPort,
					Protocol: ServiceEntryPortProtocol,
				}},
				Resolution: alpha3.ServiceEntry_DNS,
			},
		})
	}

	return err
}

func (i *istioFederationClient) determineExternalIpForGateway(
	ctx context.Context,
	meshGroupName, clusterName string,
	dynamicClient client.Client,
) (eap dns.ExternalAccessPoint, err error) {

	// implicitly convert the map[string]string to a client.MatchingLabels
	var labels client.MatchingLabels = BuildGatewayWorkloadSelector()
	gatewayServiceList, err := i.serviceClientFactory(dynamicClient).List(ctx, labels)

	if err != nil {
		return eap, err
	}

	if len(gatewayServiceList.Items) == 0 {
		return eap, eris.Errorf("No gateway for group %s has been initialized yet", meshGroupName)
	} else if len(gatewayServiceList.Items) != 1 {
		return eap, eris.Errorf("Istio gateway for group %s is in an unknown state with multiple services for the ingress gateway", meshGroupName)
	}

	service := gatewayServiceList.Items[0]

	if len(service.Spec.Ports) == 0 {
		return eap, eris.Errorf("service %+v is missing ports", service.ObjectMeta)
	}

	return i.externalAccessPointGetter.GetExternalAccessPointForService(ctx, &service, DefaultGatewayPortName, clusterName)
}

func (i *istioFederationClient) ensureEnvoyFilterExists(
	ctx context.Context,
	groupName string,
	dynamicClient client.Client,
	installNamespace string,
	clusterName string,
) error {

	envoyFilterClient := i.envoyFilterClientFactory(dynamicClient)
	computedRef := &core_types.ResourceRef{
		Name:      fmt.Sprintf("smh-%s-filter", groupName),
		Namespace: installNamespace,
	}

	_, err := envoyFilterClient.Get(ctx, clients.ResourceRefToObjectKey(computedRef))
	if errors.IsNotFound(err) {
		filterPatch, err := BuildClusterReplacementPatch(clusterName)
		if err != nil {
			return err
		}

		// see https://github.com/solo-io/mesh-projects/issues/195 for details on this envoy filter config
		return envoyFilterClient.Create(ctx, &v1alpha3.EnvoyFilter{
			ObjectMeta: clients.ResourceRefToObjectMeta(computedRef),
			Spec: alpha3.EnvoyFilter{
				ConfigPatches: []*alpha3.EnvoyFilter_EnvoyConfigObjectPatch{{
					ApplyTo: alpha3.EnvoyFilter_NETWORK_FILTER,
					Match: &alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
						Context: alpha3.EnvoyFilter_GATEWAY,
						ObjectTypes: &alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
							Listener: &alpha3.EnvoyFilter_ListenerMatch{
								PortNumber: DefaultGatewayPort,
							},
						},
					},
					Patch: &alpha3.EnvoyFilter_Patch{
						Operation: alpha3.EnvoyFilter_Patch_INSERT_AFTER,
						Value:     filterPatch,
					},
				}},
				WorkloadSelector: &alpha3.WorkloadSelector{
					Labels: BuildGatewayWorkloadSelector(),
				},
			},
		})
	}

	return err
}

func (i *istioFederationClient) ensureGatewayExists(
	ctx context.Context,
	dynamicClient client.Client,
	groupName string,
	meshService *discovery_v1alpha1.MeshService,
	installNamespace string,
) error {

	gatewayClient := i.gatewayClientFactory(dynamicClient)

	computedGatewayRef := &core_types.ResourceRef{
		Name:      fmt.Sprintf("smh-group-%s-gateway", groupName),
		Namespace: installNamespace,
	}

	existingGateway, err := gatewayClient.Get(ctx, clients.ResourceRefToObjectKey(computedGatewayRef))
	serviceDnsName := meshService.Spec.GetFederation().GetMulticlusterDnsName()
	if errors.IsNotFound(err) {
		// if the gateway wasn't found, then create our initial state
		return gatewayClient.Create(ctx, &v1alpha3.Gateway{
			ObjectMeta: clients.ResourceRefToObjectMeta(computedGatewayRef),
			Spec: alpha3.Gateway{
				Servers: []*alpha3.Server{{
					Port: &alpha3.Port{
						Number:   DefaultGatewayPort,
						Protocol: DefaultGatewayProtocol,
						Name:     DefaultGatewayPortName,
					},
					Hosts: []string{
						// initially create the gateway with just the one service's host
						serviceDnsName,
					},
					Tls: &alpha3.Server_TLSOptions{
						Mode: alpha3.Server_TLSOptions_AUTO_PASSTHROUGH,
					},
				}},
				Selector: BuildGatewayWorkloadSelector(),
			},
		})
	} else if err != nil {
		return err
	} else {
		// we assume that we have set up the gateway correctly and that it hasn't been tampered with
		// in that case, we will always have at least one entry in here
		serverConfig := existingGateway.Spec.GetServers()[0]

		// prevent duplicate hostnames from getting into the Hosts list
		gatewayHasHostname := false
		for _, hostName := range serverConfig.GetHosts() {
			if hostName == serviceDnsName {
				gatewayHasHostname = true
			}
		}
		if gatewayHasHostname {
			return nil
		}

		// update this gateway with the multicluster dns name for this mesh service
		serverConfig.Hosts = append(serverConfig.GetHosts(), serviceDnsName)

		err = gatewayClient.Update(ctx, existingGateway)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *istioFederationClient) getClientForMesh(ctx context.Context, meshRef *core_types.ResourceRef) (*discovery_v1alpha1.Mesh, client.Client, error) {
	mesh, err := i.meshClient.Get(ctx, clients.ResourceRefToObjectKey(meshRef))
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, ok := i.dynamicClientGetter.GetClientForCluster(mesh.Spec.GetCluster().GetName())
	if !ok {
		return nil, nil, ClusterNotReady(mesh.Spec.GetCluster().GetName())
	}
	return mesh, dynamicClient, nil
}

// always use the same selector - we assert elsewhere that when a mesh joins a group it must have a gateway set up
// so this will use the same gateway workload as was already set up
func BuildGatewayWorkloadSelector() map[string]string {
	return map[string]string{
		// "istio": "ingressgateway" is a known string pair to Istio- it's semantically meaningful but unfortunately not exported from anywhere
		// their ingress gateway is hardcoded in their own implementation to have this label
		// https://github.com/istio/istio/blob/4e27ddc64f6a12e622c4cd5c836f5d7edf94e971/istioctl/cmd/describe.go#L1138
		"istio": "ingressgateway",
	}
}

func BuildClusterReplacementPatch(clusterName string) (*types.Struct, error) {
	clusterReplacement, err := proto_conversion.MessageToStruct(&v2alpha1.TcpClusterRewrite{
		ClusterPattern:     fmt.Sprintf("\\.%s$", clusterName),
		ClusterReplacement: ".svc.cluster.local", // TODO https://github.com/solo-io/mesh-projects/issues/240
	})
	if err != nil {
		return nil, err
	}
	return &types.Struct{
		Fields: map[string]*types.Value{
			"name": {
				Kind: &types.Value_StringValue{
					StringValue: EnvoySniClusterFilterName,
				},
			},
			"config": {
				Kind: &types.Value_StructValue{
					StructValue: clusterReplacement,
				},
			},
		},
	}, nil
}
