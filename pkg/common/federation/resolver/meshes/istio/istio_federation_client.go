package istio_federation

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/types"
	"github.com/rotisserie/eris"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	istio_networking "github.com/solo-io/service-mesh-hub/pkg/api/istio/networking/v1alpha3"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/federation/dns"
	"github.com/solo-io/service-mesh-hub/pkg/common/federation/resolver/meshes"
	"github.com/solo-io/service-mesh-hub/pkg/common/federation/resolver/meshes/istio/proto_conversion"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/security/proto/envoy/config/filter/network/tcp_cluster_rewrite/v2alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ServiceNotInIstio = func(service *smh_discovery.MeshService) error {
		return eris.Errorf("Service %s.%s does not belong to an Istio mesh", service.Name, service.Namespace)
	}
	WorkloadNotInIstio = func(workload *smh_discovery.MeshWorkload) error {
		return eris.Errorf("Workload %s.%s does not belong to an Istio mesh", workload.Name, workload.Namespace)
	}
	ClusterNotReady = func(clusterName string) error {
		return eris.Errorf("Cluster '%s' is not fully registered yet", clusterName)
	}
)

const (
	DefaultGatewayPort     = 15443 // https://istio.io/docs/ops/deployment/requirements/#ports-used-by-istio
	DefaultGatewayProtocol = "TLS"
	DefaultGatewayPortName = "tls"

	EnvoySniClusterFilterName        = "envoy.filters.network.sni_cluster"
	EnvoyTcpClusterRewriteFilterName = "envoy.filters.network.tcp_cluster_rewrite"
)

type IstioFederationClient meshes.MeshFederationClient

// istio-specific implementation of federation resolution
func NewIstioFederationClient(
	dynamicClientGetter multicluster.DynamicClientGetter,
	meshClient smh_discovery.MeshClient,
	gatewayClientFactory istio_networking.GatewayClientFactory,
	envoyFilterClientFactory istio_networking.EnvoyFilterClientFactory,
	destinationRuleClientFactory istio_networking.DestinationRuleClientFactory,
	serviceEntryClientFactory istio_networking.ServiceEntryClientFactory,
	serviceClientFactory kubernetes_core.ServiceClientFactory,
	ipAssigner dns.IpAssigner,
	externalAccessPointGetter dns.ExternalAccessPointGetter,
) IstioFederationClient {
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
	dynamicClientGetter          multicluster.DynamicClientGetter
	meshClient                   smh_discovery.MeshClient
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
	installationNamespace string,
	virtualMesh *smh_networking.VirtualMesh,
	meshService *smh_discovery.MeshService,
) (eap dns.ExternalAccessPoint, err error) {
	meshForService, dynamicClient, err := i.getClientForMesh(ctx, meshService.Spec.GetMesh())
	if err != nil {
		return eap, err
	}

	// Make sure the gateway is in a good state
	err = i.ensureGatewayExists(ctx, dynamicClient, virtualMesh.GetName(), meshService, installationNamespace)
	if err != nil {
		return eap, eris.Wrapf(err, "Failed to configure the ingress gateway for service %s.%s",
			meshService.GetName(), meshService.GetNamespace())
	}

	// ensure that the envoy filter exists
	err = i.ensureEnvoyFilterExists(ctx, virtualMesh.GetName(), dynamicClient, installationNamespace, meshForService.Spec.GetCluster().GetName())
	if err != nil {
		return eap, eris.Wrapf(err, "Failed to configure the ingress gateway envoy filter for service %s.%s",
			meshService.GetName(), meshService.GetNamespace())
	}

	// finally, send back the external IP for the gateway we just set up
	return i.determineExternalIpForGateway(ctx, virtualMesh.GetName(), meshForService.Spec.GetCluster().GetName(), dynamicClient)
}

func (i *istioFederationClient) FederateClientSide(
	ctx context.Context,
	installationNamespace string,
	eap dns.ExternalAccessPoint,
	meshService *smh_discovery.MeshService,
	meshWorkload *smh_discovery.MeshWorkload,
) error {
	meshForWorkload, clientForWorkloadMesh, err := i.getClientForMesh(ctx, meshWorkload.Spec.GetMesh())
	if err != nil {
		return err
	}

	serviceMulticlusterName := meshService.Spec.GetFederation().GetMulticlusterDnsName()

	err = i.setUpServiceEntry(
		ctx,
		clientForWorkloadMesh,
		eap,
		meshService,
		installationNamespace,
		meshForWorkload.Spec.GetCluster().GetName(),
	)
	if err != nil {
		return err
	}

	return i.setUpDestinationRule(
		ctx,
		clientForWorkloadMesh,
		serviceMulticlusterName,
		installationNamespace,
	)
}

func (i *istioFederationClient) setUpDestinationRule(
	ctx context.Context,
	clientForWorkloadMesh client.Client,
	serviceMulticlusterName string,
	installNamespace string,
) error {
	destinationRuleRef := &smh_core_types.ResourceRef{
		Name:      serviceMulticlusterName,
		Namespace: installNamespace,
	}

	destinationRuleClient := i.destinationRuleClientFactory(clientForWorkloadMesh)
	_, err := destinationRuleClient.GetDestinationRule(ctx, selection.ResourceRefToObjectKey(destinationRuleRef))

	if errors.IsNotFound(err) {
		return destinationRuleClient.CreateDestinationRule(ctx, &v1alpha3.DestinationRule{
			ObjectMeta: selection.ResourceRefToObjectMeta(destinationRuleRef),
			Spec: alpha3.DestinationRule{
				Host: serviceMulticlusterName,
				TrafficPolicy: &alpha3.TrafficPolicy{
					Tls: &alpha3.TLSSettings{
						// TODO this won't work with other mesh types https://github.com/solo-io/service-mesh-hub/issues/242
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
	meshService *smh_discovery.MeshService,
	installNamespace,
	workloadClusterName string,
) error {
	serviceEntryClient := i.serviceEntryClientFactory(clientForWorkloadMesh)

	computedRef := &smh_core_types.ResourceRef{
		Name:      meshService.Spec.GetFederation().GetMulticlusterDnsName(),
		Namespace: installNamespace,
	}

	endpoint := &alpha3.ServiceEntry_Endpoint{
		Address: eap.Address,
		Ports:   make(map[string]uint32),
	}
	var ports []*alpha3.Port
	for _, port := range meshService.Spec.GetKubeService().GetPorts() {
		ports = append(ports, &alpha3.Port{
			Number:   port.Port,
			Protocol: port.Protocol,
			Name:     port.Name,
		})
		endpoint.Ports[port.Name] = eap.Port
	}
	endpoints := []*alpha3.ServiceEntry_Endpoint{endpoint}

	existing, err := serviceEntryClient.GetServiceEntry(ctx, selection.ResourceRefToObjectKey(computedRef))
	if errors.IsNotFound(err) {
		// generate a unique IP within the workload cluster for the service entry to point to
		newIp, err := i.ipAssigner.AssignIPOnCluster(ctx, workloadClusterName)
		if err != nil {
			return err
		}
		serviceEntry := &v1alpha3.ServiceEntry{
			ObjectMeta: selection.ResourceRefToObjectMeta(computedRef),
			Spec: alpha3.ServiceEntry{
				Addresses:  []string{newIp},
				Hosts:      []string{meshService.Spec.GetFederation().GetMulticlusterDnsName()},
				Location:   alpha3.ServiceEntry_MESH_INTERNAL,
				Resolution: alpha3.ServiceEntry_DNS,
				Endpoints:  endpoints,
				Ports:      ports,
			},
		}

		return serviceEntryClient.CreateServiceEntry(ctx, serviceEntry)
	} else if err != nil {
		return err
	}

	// if the service entry already exists, update it with all ports and endpoints found on the target service.
	existing.Spec.Ports = ports
	existing.Spec.Endpoints = endpoints
	return serviceEntryClient.UpdateServiceEntry(ctx, existing)
}

func (i *istioFederationClient) determineExternalIpForGateway(
	ctx context.Context,
	virtualMeshName, clusterName string,
	dynamicClient client.Client,
) (eap dns.ExternalAccessPoint, err error) {

	// implicitly convert the map[string]string to a client.MatchingLabels
	var labels client.MatchingLabels = BuildGatewayWorkloadSelector()
	gatewayServiceList, err := i.serviceClientFactory(dynamicClient).ListService(ctx, labels)

	if err != nil {
		return eap, err
	}

	if len(gatewayServiceList.Items) == 0 {
		return eap, eris.Errorf("No gateway for virtual mesh %s has been initialized yet", virtualMeshName)
	} else if len(gatewayServiceList.Items) != 1 {
		return eap, eris.Errorf("Istio gateway for virtual mesh %s is in an unknown state with multiple services "+
			"for the ingress gateway", virtualMeshName)
	}

	service := gatewayServiceList.Items[0]

	if len(service.Spec.Ports) == 0 {
		return eap, eris.Errorf("service %s.%s in cluster %s is missing ports", service.ObjectMeta.Name,
			service.ObjectMeta.Namespace, clusterName)
	}

	return i.externalAccessPointGetter.GetExternalAccessPointForService(ctx, &service, DefaultGatewayPortName, clusterName)
}

func (i *istioFederationClient) ensureEnvoyFilterExists(
	ctx context.Context,
	vmName string,
	dynamicClient client.Client,
	installNamespace string,
	clusterName string,
) error {

	envoyFilterClient := i.envoyFilterClientFactory(dynamicClient)
	computedRef := &smh_core_types.ResourceRef{
		Name:      fmt.Sprintf("smh-%s-filter", vmName),
		Namespace: installNamespace,
	}

	filterPatch, err := BuildClusterReplacementPatch(clusterName)
	if err != nil {
		return err
	}

	// see https://github.com/solo-io/service-mesh-hub/issues/195 for details on this envoy filter config
	return envoyFilterClient.UpsertEnvoyFilterSpec(ctx, &v1alpha3.EnvoyFilter{
		ObjectMeta: selection.ResourceRefToObjectMeta(computedRef),
		Spec: alpha3.EnvoyFilter{
			ConfigPatches: []*alpha3.EnvoyFilter_EnvoyConfigObjectPatch{{
				ApplyTo: alpha3.EnvoyFilter_NETWORK_FILTER,
				Match: &alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
					Context: alpha3.EnvoyFilter_GATEWAY,
					ObjectTypes: &alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
						Listener: &alpha3.EnvoyFilter_ListenerMatch{
							PortNumber: DefaultGatewayPort,
							FilterChain: &alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
								Filter: &alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
									Name: EnvoySniClusterFilterName,
								},
							},
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

func (i *istioFederationClient) ensureGatewayExists(
	ctx context.Context,
	dynamicClient client.Client,
	virtualMeshName string,
	meshService *smh_discovery.MeshService,
	installNamespace string,
) error {

	gatewayClient := i.gatewayClientFactory(dynamicClient)

	computedGatewayRef := &smh_core_types.ResourceRef{
		Name:      fmt.Sprintf("smh-vm-%s-gateway", virtualMeshName),
		Namespace: installNamespace,
	}

	existingGateway, err := gatewayClient.GetGateway(ctx, selection.ResourceRefToObjectKey(computedGatewayRef))
	serviceDnsName := BuildMatchingMultiClusterHostName(meshService.Spec.GetFederation())
	if errors.IsNotFound(err) {
		// if the gateway wasn't found, then create our initial state
		return gatewayClient.CreateGateway(ctx, &v1alpha3.Gateway{
			ObjectMeta: selection.ResourceRefToObjectMeta(computedGatewayRef),
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

		err = gatewayClient.UpdateGateway(ctx, existingGateway)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
	Need to add a wildcard "*" matcher at the beginning of this DNS name because the SNI name
	which istio uses is not the exact host which we use on the ServiceEntry/DestinationRule.

	For example: If we are federating the pod with multicluster DNS `reviews.default.target-cluster`.
	`reviews.default.target-cluster` is the host we will use to populate the ServiceEntry/DestinationRule.
	However, istio has an internal representation of this hostname which will look something like:
	outbound_9080_v2_._review.default.target-cluster. This value will depend on the port, as well as subsets.

	This is the value which needs to be matched as well as translate into outbound_9080_v2_._review.default.svc.cluster.local
	on the target cluster.
*/
func BuildMatchingMultiClusterHostName(federationInfo *discovery_types.MeshServiceSpec_Federation) string {
	return fmt.Sprintf("*.%s", federationInfo.GetMulticlusterDnsName())
}

func (i *istioFederationClient) getClientForMesh(ctx context.Context, meshRef *smh_core_types.ResourceRef) (*smh_discovery.Mesh, client.Client, error) {
	mesh, err := i.meshClient.GetMesh(ctx, selection.ResourceRefToObjectKey(meshRef))
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, err := i.dynamicClientGetter.GetClientForCluster(ctx, mesh.Spec.GetCluster().GetName())
	if err != nil {
		return nil, nil, ClusterNotReady(mesh.Spec.GetCluster().GetName())
	}
	return mesh, dynamicClient, nil
}

// always use the same selector - we assert elsewhere that when a mesh joins a virtual mesh it must have a gateway set up
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
		ClusterReplacement: ".svc.cluster.local", // TODO https://github.com/solo-io/service-mesh-hub/issues/240
	})
	if err != nil {
		return nil, err
	}
	return &types.Struct{
		Fields: map[string]*types.Value{
			"name": {
				Kind: &types.Value_StringValue{
					StringValue: EnvoyTcpClusterRewriteFilterName,
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
