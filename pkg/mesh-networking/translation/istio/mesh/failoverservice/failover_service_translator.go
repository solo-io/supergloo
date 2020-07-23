package failoverservice

import (
	"bytes"
	"context"
	"fmt"

	udpa_type_v1 "github.com/cncf/udpa/go/udpa/type/v1"
	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_config_cluster_aggregate_v2alpha "github.com/envoyproxy/go-control-plane/envoy/config/cluster/aggregate/v2alpha"
	"github.com/envoyproxy/go-control-plane/pkg/conversion"
	gogo_jsonpb "github.com/gogo/protobuf/jsonpb"
	gogo_proto_types "github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	networkingv1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/meshserviceutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/resourceidutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/validation/failoverservice"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Outputs of translating a FailoverService for a single Mesh
type Outputs struct {
	EnvoyFilters   networkingv1alpha3sets.EnvoyFilterSet
	ServiceEntries networkingv1alpha3sets.ServiceEntrySet
}

// The FailoverService translator translates a FailoverService for a single Mesh.
type Translator interface {
	// Translate translates the FailoverService into a ServiceEntry representing the new service and an accompanying EnvoyFilter.
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.Snapshot,
		mesh *discoveryv1alpha2.Mesh,
		failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
		reporter reporting.Reporter,
	) Outputs
}

type translator struct {
	ctx            context.Context
	validator      failoverservice.FailoverServiceValidator
	clusterDomains hostutils.ClusterDomainRegistry
}

func NewTranslator(ctx context.Context, clusterDomains hostutils.ClusterDomainRegistry) Translator {
	return &translator{
		ctx:            ctx,
		validator:      failoverservice.NewFailoverServiceValidator(),
		clusterDomains: clusterDomains,
	}
}

func (t *translator) Translate(
	in input.Snapshot,
	mesh *discoveryv1alpha2.Mesh,
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	reporter reporting.Reporter,
) Outputs {
	outputs := Outputs{
		EnvoyFilters:   networkingv1alpha3sets.NewEnvoyFilterSet(),
		ServiceEntries: networkingv1alpha3sets.NewServiceEntrySet(),
	}
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.MeshType)
		return outputs
	}

	// If validation fails, report the errors to the Meshes and do not translate.
	validationErrors := t.validator.Validate(failoverservice.Inputs{
		MeshServices:  in.MeshServices(),
		KubeClusters:  in.KubernetesClusters(),
		Meshes:        in.Meshes(),
		VirtualMeshes: in.VirtualMeshes(),
	}, failoverService.Spec)
	if validationErrors != nil {
		reporter.ReportFailoverService(failoverService.Ref, validationErrors)
	}

	serviceEntries, envoyFilters, err := t.translate(failoverService, in.MeshServices().List(), in.Meshes())
	if err != nil {
		t.reportErrorsToMeshes(failoverService, in.Meshes(), err, reporter)
	} else {
		outputs.EnvoyFilters.Insert(envoyFilters...)
		outputs.ServiceEntries.Insert(serviceEntries...)
	}

	return outputs
}

func (t *translator) reportErrorsToMeshes(
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	allMeshes v1alpha2sets.MeshSet,
	reportedErr error,
	reporter reporting.Reporter,
) {
	for _, meshRef := range failoverService.Spec.Meshes {
		mesh, err := allMeshes.Find(meshRef)
		if err != nil {
			continue // Mesh reference not found
		}
		reporter.ReportFailoverServiceToMesh(mesh, failoverService.Ref, reportedErr)
	}
}

// Translate FailoverService into ServiceEntry and EnvoyFilter.
func (t *translator) translate(
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	allMeshServices []*discoveryv1alpha2.MeshService,
	allMeshes v1alpha2sets.MeshSet,
) ([]*networkingv1alpha3.ServiceEntry, []*networkingv1alpha3.EnvoyFilter, error) {
	prioritizedMeshServices, err := t.collectMeshServicesForFailoverService(failoverService.Spec, allMeshServices)
	if err != nil {
		return nil, nil, err
	}
	var multierr *multierror.Error
	if len(prioritizedMeshServices) < 1 {
		return nil, nil, eris.New("FailoverService has fewer than 1 MeshService.")
	}
	serviceEntries, err := t.translateServiceEntries(failoverService, allMeshes)
	if err != nil {
		multierr = multierror.Append(multierr, err)
	}
	envoyFilters, err := t.translateEnvoyFilters(failoverService, prioritizedMeshServices, allMeshes)
	if err != nil {
		multierr = multierror.Append(multierr, err)
	}
	return serviceEntries, envoyFilters, multierr.ErrorOrNil()
}

/*
	Collect, in priority order as declared in the FailoverService, the relevant MeshServices.
	The first MeshService is guaranteed to be the FailoverService's target service.
	If a MeshService cannot be found, return an error
*/
func (t *translator) collectMeshServicesForFailoverService(
	failoverServiceSpec *v1alpha2.FailoverServiceSpec,
	allMeshServices []*discoveryv1alpha2.MeshService,
) ([]*discoveryv1alpha2.MeshService, error) {
	var prioritizedMeshServices []*discoveryv1alpha2.MeshService
	for _, typedServiceRef := range failoverServiceSpec.ComponentServices {
		// TODO(harveyxia) add support for non-k8s services
		serviceRef := typedServiceRef.GetKubeService()
		var matchingMeshService *discoveryv1alpha2.MeshService
		for _, meshService := range allMeshServices {
			if !resourceidutils.ClusterRefsEqual(serviceRef, meshService.Spec.GetKubeService().Ref) {
				continue
			}
			matchingMeshService = meshService
		}
		if matchingMeshService == nil {
			// Should never happen because it would be caught in validation.
			return nil, failoverservice.FailoverServiceNotFound(serviceRef)
		}
		prioritizedMeshServices = append(prioritizedMeshServices, matchingMeshService)
	}
	return prioritizedMeshServices, nil
}

func (t *translator) translateServiceEntries(
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	allMeshes v1alpha2sets.MeshSet,
) ([]*networkingv1alpha3.ServiceEntry, error) {
	var multierr *multierror.Error
	var serviceEntries []*networkingv1alpha3.ServiceEntry
	for _, meshRef := range failoverService.Spec.GetMeshes() {
		mesh, err := allMeshes.Find(meshRef)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}

		ip, err := meshserviceutils.ConstructUniqueIpForFailoverService(failoverService.Ref)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		// Locate ServiceEntry in default SMH namespace—by default a ServiceEntry is exported to all namespaces.
		serviceEntry := &networkingv1alpha3.ServiceEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name:        failoverService.Ref.Name,
				Namespace:   defaults.GetPodNamespace(),
				ClusterName: mesh.Spec.GetIstio().Installation.Cluster,
			},
			Spec: networkingv1alpha3spec.ServiceEntry{
				Hosts: []string{failoverService.Spec.GetHostname()},
				Ports: []*networkingv1alpha3spec.Port{
					{
						Number:   failoverService.Spec.GetPort().GetPort(),
						Protocol: failoverService.Spec.GetPort().GetProtocol(),
						Name:     failoverService.Spec.GetPort().GetProtocol(), // Name the port with the protocol
					},
				},
				Addresses: []string{ip.String()},
				// Treat remote cluster services as part of the service mesh as all clusters in the service mesh share the same root of trust.
				Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
				Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
			},
		}
		serviceEntries = append(serviceEntries, serviceEntry)
	}
	return serviceEntries, multierr.ErrorOrNil()
}

func (t *translator) translateEnvoyFilters(
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	prioritizedMeshServices []*discoveryv1alpha2.MeshService,
	allMeshes v1alpha2sets.MeshSet,
) ([]*networkingv1alpha3.EnvoyFilter, error) {
	var multierr *multierror.Error
	var envoyFilters []*networkingv1alpha3.EnvoyFilter
	for _, meshRef := range failoverService.Spec.GetMeshes() {
		mesh, err := allMeshes.Find(meshRef)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		patches, err := t.buildFailoverEnvoyPatches(failoverService, prioritizedMeshServices, mesh)
		if err != nil {
			return nil, err
		}
		envoyFilter := &networkingv1alpha3.EnvoyFilter{
			// EnvoyFilter must be located in the root config namespace ('istio-system' by default) in order to apply to all workloads in the Mesh.
			ObjectMeta: metav1.ObjectMeta{
				Name:        failoverService.Ref.Name,
				Namespace:   mesh.Spec.GetIstio().Installation.Namespace,
				ClusterName: mesh.Spec.GetIstio().Installation.Cluster,
			},
			Spec: networkingv1alpha3spec.EnvoyFilter{
				ConfigPatches: patches,
			},
		}
		envoyFilters = append(envoyFilters, envoyFilter)
	}
	return envoyFilters, nil
}

func (t *translator) buildFailoverEnvoyPatches(
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	prioritizedServices []*discoveryv1alpha2.MeshService,
	mesh *discoveryv1alpha2.Mesh,
) ([]*networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectPatch, error) {
	var failoverAggregateClusterPatches []*networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectPatch
	failoverServiceClusterString := buildIstioEnvoyClusterName(failoverService.Spec.GetPort().GetPort(), failoverService.Spec.GetHostname())
	envoyFailoverPatch, err := t.buildEnvoyFailoverPatch(
		failoverServiceClusterString,
		mesh.Spec.GetIstio().Installation.Cluster,
		prioritizedServices,
	)
	if err != nil {
		return nil, err
	}
	// EnvoyFilter patches representing the aggregate cluster for the failover service.
	failoverAggregateClusterPatch := []*networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectPatch{
		// Replace the default Envoy configuration for Istio ServiceEntry with custom Envoy failover config
		{
			ApplyTo: networkingv1alpha3spec.EnvoyFilter_CLUSTER,
			Match: &networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectMatch{
				Context: networkingv1alpha3spec.EnvoyFilter_ANY,
				ObjectTypes: &networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectMatch_Cluster{
					Cluster: &networkingv1alpha3spec.EnvoyFilter_ClusterMatch{
						Name: failoverServiceClusterString,
					},
				},
			},
			Patch: &networkingv1alpha3spec.EnvoyFilter_Patch{
				Operation: networkingv1alpha3spec.EnvoyFilter_Patch_REMOVE,
			},
		},
		{
			ApplyTo: networkingv1alpha3spec.EnvoyFilter_CLUSTER,
			Match: &networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectMatch{
				Context: networkingv1alpha3spec.EnvoyFilter_ANY,
				ObjectTypes: &networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectMatch_Cluster{
					Cluster: &networkingv1alpha3spec.EnvoyFilter_ClusterMatch{
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

func (t *translator) buildEnvoyFailoverPatch(
	failoverServiceEnvoyClusterName string,
	failoverServiceCluster string,
	prioritizedServices []*discoveryv1alpha2.MeshService,
) (*networkingv1alpha3spec.EnvoyFilter_Patch, error) {
	aggregateClusterConfig := t.buildEnvoyAggregateClusterConfig(prioritizedServices, failoverServiceCluster)
	aggregateClusterConfigStruct, err := conversion.MessageToStruct(aggregateClusterConfig)
	if err != nil {
		return nil, err
	}
	aggregateCluster, err := ptypes.MarshalAny(&udpa_type_v1.TypedStruct{
		TypeUrl: "type.googleapis.com/envoy.config.cluster.aggregate.v2alpha.ClusterConfig",
		Value:   aggregateClusterConfigStruct,
	})
	if err != nil {
		return nil, err
	}
	envoyCluster := &envoy_api_v2.Cluster{
		Name: failoverServiceEnvoyClusterName,
		ConnectTimeout: &duration.Duration{
			Seconds: 1,
		},
		LbPolicy: envoy_api_v2.Cluster_CLUSTER_PROVIDED,
		ClusterDiscoveryType: &envoy_api_v2.Cluster_ClusterType{
			ClusterType: &envoy_api_v2.Cluster_CustomClusterType{
				Name:        "envoy.clusters.aggregate",
				TypedConfig: aggregateCluster,
			},
		},
	}
	// This is needed because Envoy API's use Golang protobufs whereas Istio API's use Gogo protobufs.
	envoyClusterStruct, err := golangMessageToGogoStruct(envoyCluster)
	if err != nil {
		return nil, err
	}
	return &networkingv1alpha3spec.EnvoyFilter_Patch{
		Operation: networkingv1alpha3spec.EnvoyFilter_Patch_ADD,
		Value:     envoyClusterStruct,
	}, nil
}

// Convert list of MeshServices corresponding to FailoverService.Spec.services to
// an envoy ClusterConfig consisting of the list of Envoy cluster strings.
func (t *translator) buildEnvoyAggregateClusterConfig(
	meshServices []*discoveryv1alpha2.MeshService,
	failoverServiceClusterName string,
) *envoy_config_cluster_aggregate_v2alpha.ClusterConfig {
	var orderedFailoverList []string
	for _, meshService := range meshServices {
		for _, port := range meshService.Spec.GetKubeService().Ports {
			var hostname string
			if meshService.Spec.GetKubeService().Ref.ClusterName == failoverServiceClusterName {
				// Local k8s DNS
				hostname = t.clusterDomains.GetServiceLocalFQDN(meshService)
			} else {
				// Multicluster global DNS
				hostname = t.clusterDomains.GetServiceGlobalFQDN(meshService.Spec.GetKubeService().Ref)
			}
			failoverCluster := buildIstioEnvoyClusterName(port.GetPort(), hostname)
			orderedFailoverList = append(orderedFailoverList, failoverCluster)
		}
	}
	return &envoy_config_cluster_aggregate_v2alpha.ClusterConfig{
		Clusters: orderedFailoverList,
	}
}

func buildIstioEnvoyClusterName(port uint32, hostname string) string {
	return fmt.Sprintf("outbound|%d||%s", port, hostname)
}

func golangMessageToGogoStruct(msg proto.Message) (*gogo_proto_types.Struct, error) {
	if msg == nil {
		return nil, eris.New("nil message")
	}
	// Marshal to bytes using golang protobuf
	buf := &bytes.Buffer{}
	if err := (&jsonpb.Marshaler{OrigName: true}).Marshal(buf, msg); err != nil {
		return nil, err
	}
	// Unmarshal to gogo protobuf Struct using gogo unmarshaller
	pbs := &gogo_proto_types.Struct{}
	if err := gogo_jsonpb.Unmarshal(buf, pbs); err != nil {
		return nil, err
	}
	return pbs, nil
}
