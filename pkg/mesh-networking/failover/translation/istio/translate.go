package istio

import (
	"context"
	"fmt"

	proto_types "github.com/gogo/protobuf/types"
	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/translation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/dns"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	istio_networking "istio.io/api/networking/v1alpha3"
	istio_client_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TranslatorId = "istio-failoverservice-translator"
)

type istioFailoverServiceTranslator struct {
	ipAssigner dns.IpAssigner
}

type IstioFailoverServiceTranslator translation.FailoverServiceTranslator

func NewIstioFailoverServiceTranslator(ipAssigner dns.IpAssigner) IstioFailoverServiceTranslator {
	return &istioFailoverServiceTranslator{
		ipAssigner: ipAssigner,
	}
}

func (i *istioFailoverServiceTranslator) Translate(
	ctx context.Context,
	failoverService *smh_networking.FailoverService,
	prioritizedMeshServices []*smh_discovery.MeshService,
	allMeshes v1alpha1sets.MeshSet,
) (failover.MeshOutputs, *types.FailoverServiceStatus_TranslatorError) {
	output := failover.NewMeshOutputs()
	var translatorErr *types.FailoverServiceStatus_TranslatorError
	serviceEntries, envoyFilters, err := i.translate(ctx, failoverService, prioritizedMeshServices, allMeshes)
	if err != nil {
		translatorErr = i.translatorErr(err)
	} else {
		output.ServiceEntries.Insert(serviceEntries...)
		output.EnvoyFilters.Insert(envoyFilters...)
	}
	return output, translatorErr
}

// Translate FailoverService into ServiceEntry and EnvoyFilter.
func (i *istioFailoverServiceTranslator) translate(
	ctx context.Context,
	failoverService *smh_networking.FailoverService,
	prioritizedMeshServices []*smh_discovery.MeshService,
	allMeshes v1alpha1sets.MeshSet,
) ([]*istio_client_networking.ServiceEntry, []*istio_client_networking.EnvoyFilter, error) {
	var multierr *multierror.Error
	if len(prioritizedMeshServices) < 1 {
		return nil, nil, eris.New("FailoverService has fewer than 1 MeshService.")
	}
	serviceEntries, err := i.translateServiceEntries(ctx, failoverService, allMeshes)
	if err != nil {
		multierr = multierror.Append(multierr, err)
	}
	envoyFilters, err := i.translateEnvoyFilters(failoverService, prioritizedMeshServices, allMeshes)
	if err != nil {
		multierr = multierror.Append(multierr, err)
	}
	return serviceEntries, envoyFilters, multierr.ErrorOrNil()
}

func (i *istioFailoverServiceTranslator) translateServiceEntries(
	ctx context.Context,
	failoverService *smh_networking.FailoverService,
	allMeshes v1alpha1sets.MeshSet,
) ([]*istio_client_networking.ServiceEntry, error) {
	var multierr *multierror.Error
	var serviceEntries []*istio_client_networking.ServiceEntry
	for _, meshRef := range failoverService.Spec.GetMeshes() {
		mesh, err := allMeshes.Find(&v1.ClusterObjectRef{
			Name:      meshRef.GetName(),
			Namespace: meshRef.GetNamespace(),
		})
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		ip, err := i.ipAssigner.AssignIPOnCluster(ctx, mesh.Spec.GetCluster().GetName())
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		serviceEntry := &istio_client_networking.ServiceEntry{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name:        failoverService.GetName(),
				Namespace:   failoverService.Spec.GetNamespace(),
				ClusterName: mesh.Spec.GetCluster().GetName(),
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
		serviceEntries = append(serviceEntries, serviceEntry)
	}
	return serviceEntries, multierr.ErrorOrNil()
}

func (i *istioFailoverServiceTranslator) translateEnvoyFilters(
	failoverService *smh_networking.FailoverService,
	prioritizedMeshServices []*smh_discovery.MeshService,
	allMeshes v1alpha1sets.MeshSet,
) ([]*istio_client_networking.EnvoyFilter, error) {
	var multierr *multierror.Error
	var envoyFilters []*istio_client_networking.EnvoyFilter
	for _, meshRef := range failoverService.Spec.GetMeshes() {
		mesh, err := allMeshes.Find(&v1.ClusterObjectRef{
			Name:      meshRef.GetName(),
			Namespace: meshRef.GetNamespace(),
		})
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		patches, err := i.buildFailoverEnvoyPatches(failoverService, prioritizedMeshServices, mesh)
		if err != nil {
			return nil, err
		}
		envoyFilter := &istio_client_networking.EnvoyFilter{
			// EnvoyFilter must be located in the root config namespace ('istio-system' by default) in order to apply to all workloads in the Mesh.
			ObjectMeta: k8s_meta.ObjectMeta{
				Name:        failoverService.GetName(),
				Namespace:   fetchIstioInstallationNamespace(mesh),
				ClusterName: mesh.Spec.GetCluster().GetName(),
			},
			Spec: istio_networking.EnvoyFilter{
				ConfigPatches: patches,
			},
		}
		envoyFilters = append(envoyFilters, envoyFilter)
	}
	return envoyFilters, nil
}

func (i *istioFailoverServiceTranslator) buildFailoverEnvoyPatches(
	failoverService *smh_networking.FailoverService,
	prioritizedServices []*smh_discovery.MeshService,
	mesh *smh_discovery.Mesh,
) ([]*istio_networking.EnvoyFilter_EnvoyConfigObjectPatch, error) {
	var failoverAggregateClusterPatches []*istio_networking.EnvoyFilter_EnvoyConfigObjectPatch
	failoverServiceClusterString := buildIstioEnvoyClusterName(failoverService.Spec.GetPort().GetPort(), failoverService.Spec.GetHostname())
	envoyFailoverPatch, err := i.buildEnvoyFailoverPatch(
		failoverServiceClusterString,
		mesh.Spec.GetCluster().GetName(),
		prioritizedServices,
	)
	if err != nil {
		return nil, err
	}
	// EnvoyFilter patches representing the aggregate cluster for the failover service.
	failoverAggregateClusterPatch := []*istio_networking.EnvoyFilter_EnvoyConfigObjectPatch{
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
	}
	failoverAggregateClusterPatches = append(failoverAggregateClusterPatches, failoverAggregateClusterPatch...)
	return failoverAggregateClusterPatches, nil
}

func (i *istioFailoverServiceTranslator) buildEnvoyFailoverPatch(
	failoverServiceEnvoyClusterName string,
	failoverServiceCluster string,
	prioritizedServices []*smh_discovery.MeshService,
) (*istio_networking.EnvoyFilter_Patch, error) {
	orderedFailoverList, err := i.convertServicesToEnvoyClusterList(prioritizedServices, failoverServiceCluster)
	if err != nil {
		return nil, err
	}
	return &istio_networking.EnvoyFilter_Patch{
		Operation: istio_networking.EnvoyFilter_Patch_ADD,
		Value: &proto_types.Struct{
			Fields: map[string]*proto_types.Value{
				"name":            protoStringValue(failoverServiceEnvoyClusterName),
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
		for _, port := range meshService.Spec.GetKubeService().GetPorts() {
			var hostname string
			if meshService.Spec.GetKubeService().GetRef().GetCluster() == failoverServiceClusterName {
				// Local k8s DNS
				hostname = metadata.BuildLocalFQDN(meshService)
			} else {
				// Multicluster remote DNS
				hostname = meshService.Spec.GetFederation().GetMulticlusterDnsName()
			}
			failoverCluster := protoStringValue(buildIstioEnvoyClusterName(port.GetPort(), hostname))
			orderedFailoverList.ListValue.Values = append(orderedFailoverList.ListValue.Values, failoverCluster)
		}
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

func buildIstioEnvoyClusterName(port uint32, hostname string) string {
	return fmt.Sprintf("outbound|%d||%s", port, hostname)
}

func fetchIstioInstallationNamespace(mesh *smh_discovery.Mesh) string {
	if mesh.Spec.GetIstio1_5() != nil {
		return mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation().GetInstallationNamespace()
	} else {
		return mesh.Spec.GetIstio1_6().GetMetadata().GetInstallation().GetInstallationNamespace()
	}
}
