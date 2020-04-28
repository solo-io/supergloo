package linkerd_translator

import (
	"context"
	"fmt"
	"sort"

	mc_manager "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/k8s"
	"k8s.io/apimachinery/pkg/api/resource"

	smi_config "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1"
	smi_networking "github.com/solo-io/service-mesh-hub/pkg/api/smi/split/v1alpha1"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	linkerd_config "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	linkerd_client "github.com/solo-io/service-mesh-hub/pkg/api/linkerd/v1alpha2"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator"
)

const (
	TranslatorId = "linkerd-translator"
)

var (
	SubsetsNotSupportedErr = func(dest *zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination) error {
		return eris.Errorf("Subsets are currently not supported, found one on destination: %+v", dest)
	}

	CrossNamespaceSplitNotSupportedErr = eris.Errorf("Traffic Shifts are currently not supported across namespaces in Linkerd")

	TrafficShiftRedefinedErr = func(meshService *zephyr_discovery.MeshService, trafficPoliciesWithTrafficShifts []zephyr_core_types.ResourceRef) error {
		return eris.Errorf("multiple traffic policies with traffic shifts (%+v) defined for a single mesh service (%s)", trafficPoliciesWithTrafficShifts, meshService.Name)
	}
)

type LinkerdTranslator traffic_policy_translator.TrafficPolicyMeshTranslator

func NewLinkerdTrafficPolicyTranslator(
	dynamicClientGetter mc_manager.DynamicClientGetter,
	meshClient zephyr_discovery.MeshClient,
	serviceProfileClientFactory linkerd_client.ServiceProfileClientFactory,
	trafficSplitClientFactory smi_networking.TrafficSplitClientFactory,
) LinkerdTranslator {
	return &linkerdTrafficPolicyTranslator{
		dynamicClientGetter:         dynamicClientGetter,
		meshClient:                  meshClient,
		serviceProfileClientFactory: serviceProfileClientFactory,
		trafficSplitClientFactory:   trafficSplitClientFactory,
	}
}

type linkerdTrafficPolicyTranslator struct {
	dynamicClientGetter         mc_manager.DynamicClientGetter
	meshClient                  zephyr_discovery.MeshClient
	serviceProfileClientFactory linkerd_client.ServiceProfileClientFactory
	trafficSplitClientFactory   smi_networking.TrafficSplitClientFactory
}

func (i *linkerdTrafficPolicyTranslator) Name() string {
	return TranslatorId
}

/*
	Translate a TrafficPolicy into the following Linkerd specific configuration:
	https://linkerd.io/docs/concepts/traffic-management/

	1. ServiceProfile - routing rules (retries, timeouts)
	2. TrafficSplit - routing rules (traffic shifts)
*/
func (i *linkerdTrafficPolicyTranslator) TranslateTrafficPolicy(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
	mesh *zephyr_discovery.Mesh,
	mergedTrafficPolicies []*zephyr_networking.TrafficPolicy,
) *zephyr_networking_types.TrafficPolicyStatus_TranslatorError {
	if mesh.Spec.GetLinkerd() == nil {
		return nil
	}
	serviceProfileClient, trafficSplitClient, err := i.fetchClientsForMeshService(ctx, meshService)
	if err != nil {
		return errorToStatus(err)
	}

	translatorErrs := i.ensureServiceProfile(ctx, mesh, meshService, mergedTrafficPolicies, serviceProfileClient)
	if trafficSplitError := i.ensureTrafficSplit(ctx, mesh, meshService, mergedTrafficPolicies, trafficSplitClient); trafficSplitError != nil {
		translatorErrs = multierror.Append(translatorErrs, trafficSplitError)
	}

	if translatorErrs != nil {
		return errorToStatus(translatorErrs)
	}
	return nil
}

// get ServiceProfile and ServiceProfile clients for MeshService's cluster
func (i *linkerdTrafficPolicyTranslator) fetchClientsForMeshService(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
) (linkerd_client.ServiceProfileClient, smi_networking.TrafficSplitClient, error) {
	mesh, err := i.getMesh(ctx, meshService)
	if err != nil {
		return nil, nil, err
	}
	dynamicClient, err := i.dynamicClientGetter.GetClientForCluster(ctx, mesh.Spec.GetCluster().GetName())
	if err != nil {
		return nil, nil, err
	}
	return i.serviceProfileClientFactory(dynamicClient), i.trafficSplitClientFactory(dynamicClient), nil
}

func (i *linkerdTrafficPolicyTranslator) getMesh(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
) (*zephyr_discovery.Mesh, error) {
	mesh, err := i.meshClient.GetMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
	if err != nil {
		return nil, err
	}
	return mesh, nil
}

func (i *linkerdTrafficPolicyTranslator) ensureServiceProfile(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
	meshService *zephyr_discovery.MeshService,
	mergedTrafficPolicies []*zephyr_networking.TrafficPolicy,
	serviceProfileClient linkerd_client.ServiceProfileClient,
) error {
	computedServiceProfile := i.translateIntoServiceProfile(ctx, mesh, meshService, mergedTrafficPolicies)

	// if service profile has no routes,
	// no reason to write it
	if len(computedServiceProfile.Spec.Routes) == 0 {
		return nil
	}

	// Upsert computed ServiceProfile
	err := serviceProfileClient.UpsertServiceProfileSpec(ctx, computedServiceProfile)
	if err != nil {
		return err
	}
	return nil
}

func (i *linkerdTrafficPolicyTranslator) translateIntoServiceProfile(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
	meshService *zephyr_discovery.MeshService,
	trafficPolicies []*zephyr_networking.TrafficPolicy,
) *linkerd_config.ServiceProfile {

	serviceProfile := &linkerd_config.ServiceProfile{
		ObjectMeta: metaForMeshService(meshService.Spec.GetKubeService(), mesh.Spec.GetLinkerd()),
		Spec: linkerd_config.ServiceProfileSpec{
			Routes: makeRoutes(ctx, trafficPolicies),
		},
	}

	return serviceProfile
}

func makeRoutes(
	ctx context.Context,
	trafficPolicies []*zephyr_networking.TrafficPolicy,
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
			matchers = []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{{
				PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{Prefix: "/"},
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
func hasRelevantConfig(tp *zephyr_networking.TrafficPolicy) bool {
	return tp.Spec.RequestTimeout != nil ||
		tp.Spec.TrafficShift != nil ||
		len(tp.Spec.HttpRequestMatchers) != 0
}

func getMatchCondition(ctx context.Context, matches []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher) *linkerd_config.RequestMatch {
	var conditions []*linkerd_config.RequestMatch
	for _, match := range matches {
		conditions = append(conditions, getMatch(ctx, match))
	}
	return &linkerd_config.RequestMatch{
		Any: conditions,
	}
}

func getMatch(ctx context.Context, match *zephyr_networking_types.TrafficPolicySpec_HttpMatcher) *linkerd_config.RequestMatch {
	pathRegex := func() string {
		switch match.GetPathSpecifier().(type) {
		case *zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact:
			return match.GetExact()
		case *zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Regex:
			return match.GetRegex()
		case *zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix:
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

func errorToStatus(err error) *zephyr_networking_types.TrafficPolicyStatus_TranslatorError {
	return &zephyr_networking_types.TrafficPolicyStatus_TranslatorError{
		TranslatorId: TranslatorId,
		ErrorMessage: err.Error(),
	}
}

func (i *linkerdTrafficPolicyTranslator) ensureTrafficSplit(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
	meshService *zephyr_discovery.MeshService,
	mergedTrafficPolicies []*zephyr_networking.TrafficPolicy,
	trafficSplitClient smi_networking.TrafficSplitClient,
) error {
	computedTrafficSplit, err := i.translateIntoTrafficSplit(mesh, meshService, mergedTrafficPolicies)
	if err != nil {
		return err
	}
	// if traffic split has no splits,
	// no reason to write it
	if len(computedTrafficSplit.Spec.Backends) == 0 {
		return nil
	}

	// Upsert computed TrafficSplit
	err = trafficSplitClient.UpsertTrafficSplitSpec(ctx, computedTrafficSplit)
	if err != nil {
		return err
	}
	return nil
}

func (i *linkerdTrafficPolicyTranslator) translateIntoTrafficSplit(
	mesh *zephyr_discovery.Mesh,
	meshService *zephyr_discovery.MeshService,
	trafficPolicies []*zephyr_networking.TrafficPolicy,
) (*smi_config.TrafficSplit, error) {

	var errs error

	backends, err := i.makeBackends(meshService, trafficPolicies)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	trafficSplit := &smi_config.TrafficSplit{
		ObjectMeta: metaForMeshService(meshService.Spec.GetKubeService(), mesh.Spec.GetLinkerd()),
		Spec: smi_config.TrafficSplitSpec{
			Service:  meshService.Spec.KubeService.GetRef().GetName(),
			Backends: backends,
		},
	}

	return trafficSplit, errs
}

// For each Destination, create an Linkerd HTTPRouteDestination
func (i *linkerdTrafficPolicyTranslator) makeBackends(
	meshService *zephyr_discovery.MeshService,
	policies []*zephyr_networking.TrafficPolicy,
) ([]smi_config.TrafficSplitBackend, error) {

	var trafficShift *zephyr_networking_types.TrafficPolicySpec_MultiDestination
	var policiesWithTrafficShifts []zephyr_core_types.ResourceRef
	for _, policy := range policies {
		if policy.Spec.TrafficShift != nil {
			policiesWithTrafficShifts = append(policiesWithTrafficShifts, zephyr_core_types.ResourceRef{
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
	var translatedRouteDestinations []smi_config.TrafficSplitBackend
	for _, dest := range trafficShift.GetDestinations() {
		if dest.Destination.GetNamespace() != meshService.Spec.KubeService.GetRef().GetNamespace() {
			return nil, CrossNamespaceSplitNotSupportedErr
		}
		if dest.Subset != nil {
			// subsets are currently unsupported, so return a status error to invalidate the TrafficPolicy
			return nil, SubsetsNotSupportedErr(dest)
		}
		httpRouteDestination := smi_config.TrafficSplitBackend{
			Service: dest.Destination.Name,
			Weight:  convertWeightQuantity(trafficShift.GetDestinations(), dest.GetWeight()),
		}
		translatedRouteDestinations = append(translatedRouteDestinations, httpRouteDestination)
	}
	return translatedRouteDestinations, nil
}

// get the service profile meta for the service
func metaForMeshService(svc *types.MeshServiceSpec_KubeService, linkerd *types.MeshSpec_LinkerdMesh) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%v.%v.%v", svc.GetRef().GetName(), svc.GetRef().GetNamespace(), linkerd.GetClusterDomain()),
		Namespace: svc.GetRef().GetNamespace(),
	}
}

// convert a destination weight to a kube resource.Quantity
func convertWeightQuantity(destinations []*zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination, relativeWeight uint32) *resource.Quantity {
	var total uint32
	for _, dest := range destinations {
		total += dest.GetWeight()
	}
	milliWeights := int64(relativeWeight * 1000 / total)
	return resource.NewScaledQuantity(milliWeights, resource.Milli)
}
