package linkerd_translator

import (
	"context"
	"fmt"
	"sort"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	linkerd_config "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	linkerd_client "github.com/solo-io/service-mesh-hub/pkg/clients/linkerd/v1alpha2"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator"
)

const (
	TranslatorId = "linkerd-translator"
)

var (
	SubsetsNotSupportedErr = func(dest *networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination) error {
		return eris.Errorf("Subsets are currently not supported, found one on destination: %+v", dest)
	}

	TrafficShiftRedefinedErr = func(meshService *discovery_v1alpha1.MeshService, trafficPoliciesWithTrafficShifts []core_types.ResourceRef) error {
		return eris.Errorf("multiple traffic policies with traffic shifts (%+v) defined for a single mesh service (%s)", trafficPoliciesWithTrafficShifts, meshService.Name)
	}
)

type LinkerdTranslator traffic_policy_translator.TrafficPolicyMeshTranslator

func NewLinkerdTrafficPolicyTranslator(
	dynamicClientGetter mc_manager.DynamicClientGetter,
	meshClient zephyr_discovery.MeshClient,
	serviceProfileClientFactory linkerd_client.ServiceProfileClientFactory,
) LinkerdTranslator {
	return &linkerdTrafficPolicyTranslator{
		dynamicClientGetter:         dynamicClientGetter,
		meshClient:                  meshClient,
		serviceProfileClientFactory: serviceProfileClientFactory,
	}
}

type linkerdTrafficPolicyTranslator struct {
	dynamicClientGetter         mc_manager.DynamicClientGetter
	meshClient                  zephyr_discovery.MeshClient
	serviceProfileClientFactory linkerd_client.ServiceProfileClientFactory
}

func (i *linkerdTrafficPolicyTranslator) Name() string {
	return TranslatorId
}

/*
	Translate a TrafficPolicy into the following Linkerd specific configuration:
	https://linkerd.io/docs/concepts/traffic-management/

	1. ServiceProfile - routing rules (e.g. retries, fault injection, traffic shifts)
*/
func (i *linkerdTrafficPolicyTranslator) TranslateTrafficPolicy(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
	mesh *discovery_v1alpha1.Mesh,
	mergedTrafficPolicies []*networking_v1alpha1.TrafficPolicy,
) *networking_types.TrafficPolicyStatus_TranslatorError {
	if mesh.Spec.GetLinkerd() == nil {
		return nil
	}
	serviceProfileClient, err := i.fetchClientsForMeshService(ctx, meshService)
	if err != nil {
		return i.errorToStatus(err)
	}

	translatorError := i.ensureServiceProfile(ctx, mesh, meshService, mergedTrafficPolicies, serviceProfileClient)
	if translatorError != nil {
		return translatorError
	}
	return nil
}

// get ServiceProfile and ServiceProfile clients for MeshService's cluster
func (i *linkerdTrafficPolicyTranslator) fetchClientsForMeshService(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
) (linkerd_client.ServiceProfileClient, error) {
	mesh, err := i.getMesh(ctx, meshService)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := i.dynamicClientGetter.GetClientForCluster(mesh.Spec.GetCluster().GetName())
	if err != nil {
		return nil, err
	}
	return i.serviceProfileClientFactory(dynamicClient), nil
}

func (i *linkerdTrafficPolicyTranslator) getMesh(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
) (*discovery_v1alpha1.Mesh, error) {
	mesh, err := i.meshClient.Get(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
	if err != nil {
		return nil, err
	}
	return mesh, nil
}

func (i *linkerdTrafficPolicyTranslator) ensureServiceProfile(
	ctx context.Context,
	mesh *discovery_v1alpha1.Mesh,
	meshService *discovery_v1alpha1.MeshService,
	mergedTrafficPolicies []*networking_v1alpha1.TrafficPolicy,
	serviceProfileClient linkerd_client.ServiceProfileClient,
) *networking_types.TrafficPolicyStatus_TranslatorError {
	computedServiceProfile, err := i.translateIntoServiceProfile(ctx, mesh, meshService, mergedTrafficPolicies)
	if err != nil {
		return i.errorToStatus(err)
	}
	// if service profile has no routes,
	// no reason to write it
	if len(computedServiceProfile.Spec.Routes) == 0 {
		return nil
	}

	// Upsert computed ServiceProfile
	err = serviceProfileClient.UpsertSpec(ctx, computedServiceProfile)
	if err != nil {
		return i.errorToStatus(err)
	}
	return nil
}

func (i *linkerdTrafficPolicyTranslator) translateIntoServiceProfile(
	ctx context.Context,
	mesh *discovery_v1alpha1.Mesh,
	meshService *discovery_v1alpha1.MeshService,
	trafficPolicies []*networking_v1alpha1.TrafficPolicy,
) (*linkerd_config.ServiceProfile, error) {

	var errs error

	dstOverrides, err := i.makeWeightedDestinations(mesh, meshService, trafficPolicies)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	serviceProfile := &linkerd_config.ServiceProfile{
		ObjectMeta: serviceProfileMeta(meshService.Spec.GetKubeService(), mesh.Spec.GetLinkerd()),
		Spec: linkerd_config.ServiceProfileSpec{
			Routes:       makeRoutes(ctx, trafficPolicies),
			DstOverrides: dstOverrides,
		},
	}

	return serviceProfile, errs
}

// get the service profile meta for the service
func serviceProfileMeta(svc *types.MeshServiceSpec_KubeService, linkerd *types.MeshSpec_LinkerdMesh) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%v.%v.%v", svc.GetRef().GetName(), svc.GetRef().GetNamespace(), linkerd.GetClusterDomain()),
		Namespace: svc.GetRef().GetNamespace(),
	}
}

func makeRoutes(
	ctx context.Context,
	trafficPolicies []*networking_v1alpha1.TrafficPolicy,
) []*linkerd_config.RouteSpec {
	var routes []*linkerd_config.RouteSpec
	for _, policy := range trafficPolicies {

		if !hasRelevantConfig(policy) {
			contextutils.LoggerFrom(ctx).Warnf("policy %v applies to linkerd service, but has no relevant linkerd config", policy.Name)
			continue
		}

		matchers := policy.Spec.GetHttpRequestMatchers()
		if len(matchers) == 0 {
			// use catch-all matcher
			matchers = []*networking_types.TrafficPolicySpec_HttpMatcher{{
				PathSpecifier: &networking_types.TrafficPolicySpec_HttpMatcher_Prefix{Prefix: "/"},
			}}
		}
		var timeout string
		if reqTimeout := policy.Spec.GetRequestTimeout(); reqTimeout != nil {
			timeout = reqTimeout.String()
		}

		route := &linkerd_config.RouteSpec{
			Name:      policy.Name + "." + policy.Namespace,
			Condition: getMatchCondition(ctx, matchers),
			Timeout:   timeout,
		}
		routes = append(routes, route)
	}
	r := SpecificitySortableRoutes(routes)
	sort.Sort(r)
	return r
}

// returns true if the traffic policy has config relevant to Linkerd
func hasRelevantConfig(tp *networking_v1alpha1.TrafficPolicy) bool {
	return tp.Spec.RequestTimeout != nil ||
		tp.Spec.TrafficShift != nil ||
		len(tp.Spec.HttpRequestMatchers) != 0
}

func getMatchCondition(ctx context.Context, matches []*networking_types.TrafficPolicySpec_HttpMatcher) *linkerd_config.RequestMatch {
	var conditions []*linkerd_config.RequestMatch
	for _, match := range matches {
		conditions = append(conditions, getMatch(ctx, match))
	}
	return &linkerd_config.RequestMatch{
		Any: conditions,
	}
}

func getMatch(ctx context.Context, match *networking_types.TrafficPolicySpec_HttpMatcher) *linkerd_config.RequestMatch {
	pathRegex := func() string {
		switch match.GetPathSpecifier().(type) {
		case *networking_types.TrafficPolicySpec_HttpMatcher_Exact:
			return match.GetExact()
		case *networking_types.TrafficPolicySpec_HttpMatcher_Regex:
			return match.GetRegex()
		case *networking_types.TrafficPolicySpec_HttpMatcher_Prefix:
			return match.GetPrefix() + ".*"
		}
		contextutils.LoggerFrom(ctx).DPanicf("path specifier not implemented")
		return ""
	}()

	var method string
	if match.GetMethod() != nil {
		method = match.GetMethod().GetMethod().String()
	}

	return &linkerd_config.RequestMatch{
		PathRegex: pathRegex,
		Method:    method,
	}
}

// For each Destination, create an Linkerd HTTPRouteDestination
func (i *linkerdTrafficPolicyTranslator) makeWeightedDestinations(
	mesh *discovery_v1alpha1.Mesh,
	meshService *discovery_v1alpha1.MeshService,
	policies []*networking_v1alpha1.TrafficPolicy,
) ([]*linkerd_config.WeightedDst, error) {

	var trafficShift *networking_types.TrafficPolicySpec_MultiDestination
	var policiesWithTrafficShifts []core_types.ResourceRef
	for _, policy := range policies {
		if policy.Spec.TrafficShift != nil {
			policiesWithTrafficShifts = append(policiesWithTrafficShifts, core_types.ResourceRef{
				Name:      policy.Name,
				Namespace: policy.Namespace,
				Cluster:   policy.ClusterName,
			})
			trafficShift = policy.Spec.TrafficShift
		}
	}
	if len(policiesWithTrafficShifts) > 1 {
		return nil, TrafficShiftRedefinedErr(meshService, policiesWithTrafficShifts)
	}
	var translatedRouteDestinations []*linkerd_config.WeightedDst
	for _, dest := range trafficShift.GetDestinations() {
		hostname := buildServiceHostname(dest.Destination, mesh.Spec.GetLinkerd().GetClusterDomain(), dest.Port)
		if dest.Subset != nil {
			// subsets are currently unsupported, so return a status error to invalidate the TrafficPolicy
			return nil, SubsetsNotSupportedErr(dest)
		}
		httpRouteDestination := &linkerd_config.WeightedDst{
			Authority: hostname,
			Weight:    convertWeightQuantity(trafficShift.GetDestinations(), dest.GetWeight()),
		}
		translatedRouteDestinations = append(translatedRouteDestinations, httpRouteDestination)
	}
	return translatedRouteDestinations, nil
}

// convert a destination weight to a kube resource.Quantity
func convertWeightQuantity(destinations []*networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination, relativeWeight uint32) resource.Quantity {
	var total uint32
	for _, dest := range destinations {
		total += dest.GetWeight()
	}
	milliWeights := int64(relativeWeight * 1000 / total)
	return *resource.NewScaledQuantity(milliWeights, resource.Milli)
}

func (i *linkerdTrafficPolicyTranslator) errorToStatus(err error) *networking_types.TrafficPolicyStatus_TranslatorError {
	return &networking_types.TrafficPolicyStatus_TranslatorError{
		TranslatorId: TranslatorId,
		ErrorMessage: err.Error(),
	}
}

func buildServiceHostname(kubeSvc *core_types.ResourceRef, clusterDnsName string, port uint32) string {
	hostname := fmt.Sprintf("%v.%v.svc.%v",
		kubeSvc.GetName(),
		kubeSvc.GetNamespace(),
		clusterDnsName,
	)
	if port != 0 {
		hostname = fmt.Sprintf("%v:%v", hostname, port)
	}
	return hostname
}
