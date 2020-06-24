package istio

import (
	"context"
	"fmt"

	proto_types "github.com/gogo/protobuf/types"
	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/translation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/dns"
	istio_networking "istio.io/api/networking/v1alpha3"
	istio_client_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TranslatorId = "istio-failoverservice-translator"
)

var (
	ServiceWithNoPortError = func(service *types2.MeshServiceSpec_KubeService) error {
		return eris.Errorf("Encountered Service with no exposed ports, %s.%s.%s",
			service.GetRef().GetName(), service.GetRef().GetNamespace(), service.GetRef().GetCluster())
	}
)

type istioFailoverServiceTranslator struct {
	ipAssigner dns.IpAssigner
}

func NewIstioFailoverServiceTranslator(ipAssigner dns.IpAssigner) translation.FailoverServiceTranslator {
	return &istioFailoverServiceTranslator{ipAssigner: ipAssigner}
}

func (i *istioFailoverServiceTranslator) Translate(
	ctx context.Context,
	failoverService *smh_networking.FailoverService,
	prioritizedMeshServices []*smh_discovery.MeshService,
) (failover.MeshOutputs, *types.FailoverServiceStatus_TranslatorError) {
	output := failover.MeshOutputs{}
	var translatorErr *types.FailoverServiceStatus_TranslatorError
	serviceEntry, envoyFilter, err := i.translate(ctx, failoverService, prioritizedMeshServices)
	if err != nil {
		translatorErr = i.translatorErr(err)
	} else {
		output.ServiceEntries = []*istio_client_networking.ServiceEntry{serviceEntry}
		output.EnvoyFilters = []*istio_client_networking.EnvoyFilter{envoyFilter}
	}
	return output, translatorErr
}

// Translate FailoverService into ServiceEntry and EnvoyFilter.
func (i *istioFailoverServiceTranslator) translate(
	ctx context.Context,
	failoverService *smh_networking.FailoverService,
	prioritizedMeshServices []*smh_discovery.MeshService,
) (*istio_client_networking.ServiceEntry, *istio_client_networking.EnvoyFilter, error) {
	var multiErr *multierror.Error
	serviceEntry, err := i.translateServiceEntry(ctx, failoverService)
	if err != nil {
		multiErr = multierror.Append(multiErr, err)
	}
	envoyFilter, err := i.translateEnvoyFilter(failoverService, prioritizedMeshServices)
	if err != nil {
		multiErr = multierror.Append(multiErr, err)
	}
	err = multiErr.ErrorOrNil()
	if err != nil {
		return nil, nil, err
	}
	return serviceEntry, envoyFilter, nil
}

func (i *istioFailoverServiceTranslator) translateServiceEntry(
	ctx context.Context,
	failoverService *smh_networking.FailoverService,
) (*istio_client_networking.ServiceEntry, error) {
	ip, err := i.ipAssigner.AssignIPOnCluster(ctx, failoverService.Spec.GetCluster())
	if err != nil {
		return nil, err
	}
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
	}, nil
}

func (i *istioFailoverServiceTranslator) translateEnvoyFilter(
	failoverService *smh_networking.FailoverService,
	meshServices []*smh_discovery.MeshService,
) (*istio_client_networking.EnvoyFilter, error) {
	envoyFailoverPatch, err := i.buildEnvoyFailoverPatch(failoverService, meshServices)
	if err != nil {
		return nil, err
	}
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
					Patch: envoyFailoverPatch,
				},
			},
		},
	}, nil
}

func (i *istioFailoverServiceTranslator) buildEnvoyFailoverPatch(
	failoverService *smh_networking.FailoverService,
	meshServices []*smh_discovery.MeshService,
) (*istio_networking.EnvoyFilter_Patch, error) {
	orderedFailoverList, err := i.convertServicesToEnvoyClusterList(meshServices, failoverService.Spec.GetCluster())
	if err != nil {
		return nil, err
	}
	return &istio_networking.EnvoyFilter_Patch{
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
																	Kind: orderedFailoverList,
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
	}, nil
}

// Convert list of MeshServices corresponding to FailoverService.Spec.services to
// a list of Envoy cluster strings
func (i *istioFailoverServiceTranslator) convertServicesToEnvoyClusterList(
	meshServices []*smh_discovery.MeshService,
	failoverServiceClusterName string,
) (*proto_types.Value_ListValue, error) {
	orderedFailoverList := &proto_types.Value_ListValue{ListValue: &proto_types.ListValue{}}
	for _, meshService := range meshServices {
		meshService := meshService
		// TODO how to handle multiple Service ports?
		ports := meshService.Spec.GetKubeService().GetPorts()
		if len(ports) < 1 {
			return nil, ServiceWithNoPortError(meshService.Spec.GetKubeService())
		}
		port := ports[0].GetPort()
		var hostname string
		if meshService.Spec.GetKubeService().GetRef().GetCluster() == failoverServiceClusterName {
			// Local k8s DNS
			hostname = metadata.BuildLocalFQDN(meshService)
		} else {
			// Multicluster remote DNS
			hostname = meshService.Spec.GetFederation().GetMulticlusterDnsName()
		}
		failoverCluster := protoStringValue(fmt.Sprintf("outbound|%d||%s", port, hostname))
		orderedFailoverList.ListValue.Values = append(orderedFailoverList.ListValue.Values, failoverCluster)
	}
	return orderedFailoverList, nil
}

func (i *istioFailoverServiceTranslator) translatorErr(err error) *types.FailoverServiceStatus_TranslatorError {
	return &types.FailoverServiceStatus_TranslatorError{
		TranslatorId: TranslatorId,
		ErrorMessage: err.Error(),
	}
}

func protoStringValue(s string) *proto_types.Value {
	return &proto_types.Value{
		Kind: &proto_types.Value_StringValue{
			StringValue: s,
		},
	}
}
