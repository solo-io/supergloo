package failoverservice

import (
	"context"
	"fmt"

	"github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficshift"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"

	udpa_type_v1 "github.com/cncf/udpa/go/udpa/type/v1"
	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_config_cluster_aggregate_v2alpha "github.com/envoyproxy/go-control-plane/envoy/config/cluster/aggregate/v2alpha"
	"github.com/envoyproxy/go-control-plane/pkg/conversion"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/protoutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/traffictargetutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/validation/failoverservice"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source ./failover_service_translator.go -destination mocks/failover_service_translator.go

// The FailoverService translator translates a FailoverService for a single Mesh.
type Translator interface {
	// Translate translates the FailoverService into a ServiceEntry representing the new service and an accompanying EnvoyFilter.
	// Output resources will be added to the istio.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.Snapshot,
		mesh *discoveryv1alpha2.Mesh,
		failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
		outputs istio.Builder,
		reporter reporting.Reporter,
	)
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
	outputs istio.Builder,
	reporter reporting.Reporter,
) {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.MeshType)
		return
	}

	// If validation fails, report the errors to the Meshes and do not translate.
	validationErrors := t.validator.Validate(failoverservice.Inputs{
		TrafficTargets: in.TrafficTargets(),
		KubeClusters:   in.KubernetesClusters(),
		Meshes:         in.Meshes(),
		VirtualMeshes:  in.VirtualMeshes(),
	}, failoverService.Spec)
	if validationErrors != nil {
		reporter.ReportFailoverService(failoverService.Ref, validationErrors)
		return
	}

	serviceEntries, destinationRules, envoyFilters := t.translate(failoverService, in.TrafficTargets(), in.Meshes(), reporter)
	outputs.AddServiceEntries(serviceEntries...)
	outputs.AddDestinationRules(destinationRules...)
	outputs.AddEnvoyFilters(envoyFilters...)
}

// Translate FailoverService into ServiceEntry and EnvoyFilter.
func (t *translator) translate(
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	allTrafficTargets v1alpha2sets.TrafficTargetSet,
	allMeshes v1alpha2sets.MeshSet,
	reporter reporting.Reporter,
) (v1alpha3.ServiceEntrySlice, v1alpha3.DestinationRuleSlice, v1alpha3.EnvoyFilterSlice) {
	var serviceEntries []*networkingv1alpha3.ServiceEntry
	var destinationRules []*networkingv1alpha3.DestinationRule
	var envoyFilters []*networkingv1alpha3.EnvoyFilter
	prioritizedTrafficTargets, err := t.collectTrafficTargetsForFailoverService(failoverService.Spec, allTrafficTargets.List())
	if err != nil {
		reporter.ReportFailoverService(failoverService.Ref, []error{err})
		return nil, nil, nil
	}
	if len(prioritizedTrafficTargets) < 1 {
		reporter.ReportFailoverService(failoverService.Ref, []error{eris.New("FailoverService has fewer than one TrafficTarget.")})
		return nil, nil, nil
	}
	for _, meshRef := range failoverService.Spec.Meshes {
		mesh, err := allMeshes.Find(meshRef)
		if err != nil {
			reporter.ReportFailoverService(failoverService.Ref, []error{err})
			continue
		}

		subsets := trafficshift.MakeDestinationRuleSubsetsForFailoverService(
			failoverService,
			allTrafficTargets,
		)

		var errsForMesh *multierror.Error
		serviceEntry, err := t.translateServiceEntry(failoverService, mesh)
		if err != nil {
			errsForMesh = multierror.Append(errsForMesh, err)
		}
		destinationRule, err := t.translateDestinationRule(failoverService, subsets)
		if err != nil {
			errsForMesh = multierror.Append(errsForMesh, err)
		}
		envoyFilter, err := t.translateEnvoyFilter(failoverService, mesh, prioritizedTrafficTargets, subsets)
		if err != nil {
			errsForMesh = multierror.Append(errsForMesh, err)
		}
		errs := errsForMesh.ErrorOrNil()
		if errs != nil {
			reporter.ReportFailoverServiceToMesh(mesh, failoverService.Ref, errs)
			continue
		}

		serviceEntries = append(serviceEntries, serviceEntry)
		destinationRules = append(destinationRules, destinationRule)
		envoyFilters = append(envoyFilters, envoyFilter)
	}
	return serviceEntries, destinationRules, envoyFilters
}

/*
	Collect, in priority order as declared in the FailoverService, the relevant TrafficTargets.
	The first TrafficTarget is guaranteed to be the FailoverService's target service.
	If a TrafficTarget cannot be found, return an error
*/
func (t *translator) collectTrafficTargetsForFailoverService(
	failoverServiceSpec *v1alpha2.FailoverServiceSpec,
	allTrafficTargets []*discoveryv1alpha2.TrafficTarget,
) ([]*discoveryv1alpha2.TrafficTarget, error) {
	var prioritizedTrafficTargets []*discoveryv1alpha2.TrafficTarget
	for _, typedServiceRef := range failoverServiceSpec.BackingServices {
		// TODO(harveyxia) add support for non-k8s services
		serviceRef := typedServiceRef.GetKubeService()
		var matchingTrafficTarget *discoveryv1alpha2.TrafficTarget
		for _, trafficTarget := range allTrafficTargets {
			if !ezkube.ClusterRefsMatch(serviceRef, trafficTarget.Spec.GetKubeService().Ref) {
				continue
			}
			matchingTrafficTarget = trafficTarget
		}
		if matchingTrafficTarget == nil {
			// Should never happen because it would be caught in validation.
			return nil, failoverservice.BackingServiceNotFound(serviceRef)
		}
		prioritizedTrafficTargets = append(prioritizedTrafficTargets, matchingTrafficTarget)
	}
	return prioritizedTrafficTargets, nil
}

func (t *translator) translateServiceEntry(
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	mesh *discoveryv1alpha2.Mesh,
) (*networkingv1alpha3.ServiceEntry, error) {
	ip, err := traffictargetutils.ConstructUniqueIpForFailoverService(failoverService.Ref)
	if err != nil {
		return nil, err
	}
	// Locate ServiceEntry in default SMH namespace—by default a ServiceEntry is exported to all namespaces.
	failoverServicePort := failoverService.Spec.GetPort()
	serviceEntry := &networkingv1alpha3.ServiceEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:        failoverService.Ref.Name,
			Namespace:   defaults.GetPodNamespace(),
			ClusterName: mesh.Spec.GetIstio().Installation.Cluster,
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: networkingv1alpha3spec.ServiceEntry{
			Hosts: []string{failoverService.Spec.GetHostname()},
			Ports: []*networkingv1alpha3spec.Port{
				{
					Number:   failoverServicePort.GetNumber(),
					Protocol: failoverServicePort.GetProtocol(),
					Name:     failoverServicePort.GetProtocol(), // Name the port with the protocol
				},
			},
			Addresses: []string{ip.String()},
			// Treat remote cluster services as part of the service mesh as all clusters in the service mesh share the same root of trust.
			Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
			Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
		},
	}
	return serviceEntry, nil
}

func (t *translator) translateDestinationRule(
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	subsets []*networkingv1alpha3spec.Subset,
) (*networkingv1alpha3.DestinationRule, error) {
	// No need to output a DestinationRule if there are no subsets.
	if len(subsets) < 1 {
		return nil, nil
	}

	meta := metautils.TranslatedObjectMeta(
		&metav1.ObjectMeta{
			Name:      failoverService.Ref.Name,
			Namespace: failoverService.Ref.Namespace,
		},
		nil,
	)

	return &networkingv1alpha3.DestinationRule{
		ObjectMeta: meta,
		Spec: networkingv1alpha3spec.DestinationRule{
			Host: failoverService.Spec.Hostname,
			TrafficPolicy: &networkingv1alpha3spec.TrafficPolicy{
				Tls: &networkingv1alpha3spec.ClientTLSSettings{
					// TODO(ilackarms): currently we set all DRs to mTLS
					// in the future we'll want to make this configurable
					// https://github.com/solo-io/service-mesh-hub/issues/790
					Mode: networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
				},
			},
			Subsets: subsets,
		},
	}, nil
}

func (t *translator) translateEnvoyFilter(
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	mesh *discoveryv1alpha2.Mesh,
	prioritizedTrafficTargets []*discoveryv1alpha2.TrafficTarget,
	subsets []*networkingv1alpha3spec.Subset,
) (*networkingv1alpha3.EnvoyFilter, error) {
	patches, err := t.buildFailoverEnvoyPatches(failoverService, prioritizedTrafficTargets, mesh, subsets)
	if err != nil {
		return nil, err
	}
	istioInstallation := mesh.Spec.GetIstio().Installation
	envoyFilter := &networkingv1alpha3.EnvoyFilter{
		// EnvoyFilter must be located in the root config namespace ('istio-system' by default) in order to apply to all workloads in the Mesh.
		ObjectMeta: metav1.ObjectMeta{
			Name:        failoverService.Ref.Name,
			Namespace:   istioInstallation.Namespace,
			ClusterName: istioInstallation.Cluster,
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: networkingv1alpha3spec.EnvoyFilter{
			ConfigPatches: patches,
		},
	}
	return envoyFilter, nil
}

func (t *translator) buildFailoverEnvoyPatches(
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	prioritizedServices []*discoveryv1alpha2.TrafficTarget,
	mesh *discoveryv1alpha2.Mesh,
	subsets []*networkingv1alpha3spec.Subset,
) ([]*networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectPatch, error) {
	var failoverAggregateClusterPatches []*networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectPatch

	// Append nil for no-subset case
	subsets = append(subsets, nil)

	for _, subset := range subsets {
		subsetName := ""
		if subset != nil {
			subsetName = subset.Name
		}
		failoverServiceClusterName := buildIstioEnvoyClusterName(
			failoverService.Spec.GetPort().GetNumber(),
			subsetName,
			failoverService.Spec.GetHostname(),
		)
		envoyFailoverPatch, err := t.buildEnvoyFailoverPatch(
			failoverServiceClusterName,
			mesh.Spec.GetIstio().Installation.Cluster,
			prioritizedServices,
			subsetName,
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
							Name: failoverServiceClusterName,
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
							Name: failoverServiceClusterName,
						},
					},
				},
				Patch: envoyFailoverPatch,
			},
		}
		failoverAggregateClusterPatches = append(failoverAggregateClusterPatches, failoverAggregateClusterPatch...)
	}

	return failoverAggregateClusterPatches, nil
}

func (t *translator) buildEnvoyFailoverPatch(
	failoverServiceEnvoyClusterName string,
	failoverServiceCluster string,
	prioritizedServices []*discoveryv1alpha2.TrafficTarget,
	subsetName string,
) (*networkingv1alpha3spec.EnvoyFilter_Patch, error) {
	aggregateClusterConfig := t.buildEnvoyAggregateClusterConfig(prioritizedServices, failoverServiceCluster, subsetName)
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
	envoyClusterStruct, err := protoutils.GolangMessageToGogoStruct(envoyCluster)
	if err != nil {
		return nil, err
	}
	return &networkingv1alpha3spec.EnvoyFilter_Patch{
		Operation: networkingv1alpha3spec.EnvoyFilter_Patch_ADD,
		Value:     envoyClusterStruct,
	}, nil
}

// Convert list of TrafficTargets corresponding to FailoverService.Spec.services to
// an envoy ClusterConfig consisting of the list of Envoy cluster strings.
func (t *translator) buildEnvoyAggregateClusterConfig(
	trafficTargets []*discoveryv1alpha2.TrafficTarget,
	failoverServiceClusterName string,
	subsetName string,
) *envoy_config_cluster_aggregate_v2alpha.ClusterConfig {
	var orderedFailoverList []string
	for _, trafficTarget := range trafficTargets {
		kubeService := trafficTarget.Spec.GetKubeService()
		for _, port := range kubeService.Ports {
			var hostname string
			if kubeService.Ref.ClusterName == failoverServiceClusterName {
				// Local k8s DNS
				hostname = t.clusterDomains.GetServiceLocalFQDN(kubeService.Ref)
			} else {
				// Multicluster global DNS
				hostname = t.clusterDomains.GetServiceGlobalFQDN(kubeService.Ref)
			}
			failoverCluster := buildIstioEnvoyClusterName(port.GetPort(), subsetName, hostname)
			orderedFailoverList = append(orderedFailoverList, failoverCluster)
		}
	}
	return &envoy_config_cluster_aggregate_v2alpha.ClusterConfig{
		Clusters: orderedFailoverList,
	}
}

func buildIstioEnvoyClusterName(port uint32, subsetName string, hostname string) string {
	return fmt.Sprintf("outbound|%d|%s|%s", port, subsetName, hostname)
}
