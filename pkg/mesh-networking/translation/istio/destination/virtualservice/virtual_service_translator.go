package virtualservice

import (
	"context"
	"reflect"
	"sort"

	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/routeutils"

	"github.com/golang/protobuf/proto"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/utils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/selectorutils"
	skv2sets "github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/fieldutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/pkg/equalityutils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/util/sets"
)

//go:generate mockgen -source ./virtual_service_translator.go -destination mocks/virtual_service_translator.go

// Translator translates a Destination into a VirtualService.
type Translator interface {
	/*
		Translate translates the appropriate VirtualService for the given Destination.
		returns nil if no VirtualService is required for the Destination (i.e. if no VirtualService features are required, such as subsets).

		If sourceMeshInstallation is specified, hostnames in the translated VirtualService will use global FQDNs if the Destination
		exists in a different cluster from the specified mesh (i.e. is a federated Destination). Otherwise, assume translation
		for cluster that the Destination exists in and use local FQDNs.

		Errors caused by invalid user config will be reported using the Reporter.

		Note that the input snapshot DestinationSet contains the given Destination.
	*/
	Translate(
		ctx context.Context,
		in input.LocalSnapshot,
		destination *discoveryv1.Destination,
		sourceMeshInstallation *discoveryv1.MeshInstallation,
		reporter reporting.Reporter,
	) *networkingv1alpha3.VirtualService
}

type translator struct {
	userVirtualServices v1alpha3sets.VirtualServiceSet
	clusterDomains      hostutils.ClusterDomainRegistry
	decoratorFactory    decorators.Factory
}

func NewTranslator(
	userVirtualServices v1alpha3sets.VirtualServiceSet,
	clusterDomains hostutils.ClusterDomainRegistry,
	decoratorFactory decorators.Factory,
) Translator {
	return &translator{
		userVirtualServices: userVirtualServices,
		clusterDomains:      clusterDomains,
		decoratorFactory:    decoratorFactory,
	}
}

// Translate a VirtualService for the Destination.
// If sourceMeshInstallation is nil, assume that VirtualService is colocated to the Destination and use local FQDNs.
func (t *translator) Translate(
	_ context.Context,
	in input.LocalSnapshot,
	destination *discoveryv1.Destination,
	sourceMeshInstallation *discoveryv1.MeshInstallation,
	reporter reporting.Reporter,
) *networkingv1alpha3.VirtualService {
	kubeService := destination.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(ilackarms): non kube services currently unsupported
		return nil
	}

	sourceCluster := kubeService.Ref.ClusterName
	if sourceMeshInstallation != nil {
		sourceCluster = sourceMeshInstallation.Cluster
	}

	destinationFQDN := t.clusterDomains.GetDestinationFQDN(sourceCluster, destination.Spec.GetKubeService().Ref)

	virtualService := t.initializeVirtualService(destination, sourceMeshInstallation, destinationFQDN)
	// register the owners of the virtualservice fields
	virtualServiceFields := fieldutils.NewOwnershipRegistry()
	vsDecorators := t.decoratorFactory.MakeDecorators(decorators.Parameters{
		ClusterDomains: t.clusterDomains,
		Snapshot:       in,
	})

	appliedTpsByRequestMatcher := groupAppliedTpsByRequestMatcher(destination.Status.AppliedTrafficPolicies)

	for _, tpsByRequestMatcher := range appliedTpsByRequestMatcher {

		// initialize base route for TP's group by request matcher
		baseRoute := initializeBaseRoute(tpsByRequestMatcher[0].Spec, sourceCluster)

		for _, policy := range tpsByRequestMatcher {
			// nil baseRoute indicates that this cluster is not selected by the WorkloadSelector and thus should not be translated
			if baseRoute == nil {
				continue
			}

			registerField := registerFieldFunc(virtualServiceFields, virtualService, policy.Ref)
			for _, decorator := range vsDecorators {

				if trafficPolicyDecorator, ok := decorator.(decorators.TrafficPolicyVirtualServiceDecorator); ok {
					if err := trafficPolicyDecorator.ApplyTrafficPolicyToVirtualService(
						policy,
						destination,
						sourceMeshInstallation,
						baseRoute,
						registerField,
					); err != nil {
						reporter.ReportTrafficPolicyToDestination(destination, policy.Ref, eris.Wrapf(err, "%v", decorator.DecoratorName()))
					}
				}
			}
		}

		// Avoid appending an HttpRoute that will have no affect, which occurs if no decorators mutate the baseRoute with any non-match config.
		if equalityutils.DeepEqual(initializeBaseRoute(tpsByRequestMatcher[0].Spec, sourceCluster), baseRoute) {
			continue
		}

		// set a default destination for the route (to the target Destination)
		// if a decorator has not already set it
		t.setDefaultDestinationAndPortMatchers(baseRoute, destinationFQDN)

		// construct a copy of a route for each service port
		// required because Istio needs the destination port for every route
		routesPerPort := duplicateRouteForEachPort(baseRoute, kubeService.Ports)

		// split routes with multiple HTTP matchers into one matcher per route for easier route sorting later on
		var routesWithSingleMatcher []*networkingv1alpha3spec.HTTPRoute
		for _, route := range routesPerPort {
			splitRoutes := splitRouteByMatchers(route)
			routesWithSingleMatcher = append(routesWithSingleMatcher, splitRoutes...)
		}

		virtualService.Spec.Http = append(virtualService.Spec.Http, routesWithSingleMatcher...)
	}

	sort.Sort(RoutesBySpecificity(virtualService.Spec.Http))

	if len(virtualService.Spec.Http) == 0 {
		// no need to create this VirtualService as it has no effect
		return nil
	}

	if t.userVirtualServices == nil {
		return virtualService
	}

	// detect and report error on intersecting config if enabled in settings
	if errs := conflictsWithUserVirtualService(
		t.userVirtualServices,
		virtualService,
	); len(errs) > 0 {
		for _, err := range errs {
			for _, policy := range destination.Status.AppliedTrafficPolicies {
				reporter.ReportTrafficPolicyToDestination(destination, policy.Ref, err)
			}
		}
		return nil
	}

	return virtualService
}

// ensure that only a single VirtualService HTTPRoute gets created per TrafficPolicy request matcher
// by first grouping TrafficPolicies by semantically equivalent request matchers
func groupAppliedTpsByRequestMatcher(
	appliedTps []*v1.AppliedTrafficPolicy,
) [][]*v1.AppliedTrafficPolicy {
	var allGroupedTps [][]*v1.AppliedTrafficPolicy

	for _, appliedTp := range appliedTps {
		var grouped bool
		for i, groupedTps := range allGroupedTps {
			// append to existing group
			if requestMatchersEqual(appliedTp.Spec, groupedTps[0].Spec) {
				allGroupedTps[i] = append(groupedTps, appliedTp)
				grouped = true
				break
			}
		}
		// create new group
		if !grouped {
			allGroupedTps = append(allGroupedTps, []*v1.AppliedTrafficPolicy{appliedTp})
		}
	}

	return allGroupedTps
}

func requestMatchersEqual(tp1, tp2 *v1.TrafficPolicySpec) bool {
	return workloadSelectorListsEqual(tp1.GetSourceSelector(), tp2.GetSourceSelector()) &&
		httpRequestMatchersEqual(tp1.GetHttpRequestMatchers(), tp2.GetHttpRequestMatchers())
}

func httpRequestMatchersEqual(matchers1, matchers2 []*v1.HttpMatcher) bool {
	if len(matchers1) != len(matchers2) {
		return false
	}
	for i := range matchers1 {
		if !proto.Equal(matchers1[i], matchers2[i]) {
			return false
		}
	}
	return true
}

// return true if workload selectors' labels and namespaces are equivalent, ignore clusters
func workloadSelectorsEqual(ws1, ws2 *commonv1.WorkloadSelector) bool {
	return reflect.DeepEqual(ws1.GetKubeWorkloadMatcher().Labels, ws2.GetKubeWorkloadMatcher().Labels) &&
		sets.NewString(ws1.GetKubeWorkloadMatcher().Namespaces...).Equal(sets.NewString(ws2.GetKubeWorkloadMatcher().Namespaces...))
}

// return true if two lists of WorkloadSelectors are semantically equivalent, abstracting away order
func workloadSelectorListsEqual(wsList1, wsList2 []*commonv1.WorkloadSelector) bool {
	if len(wsList1) != len(wsList2) {
		return false
	}
	matchedWs2 := sets.NewInt()
	for _, ws1 := range wsList1 {
		var matched bool
		for i, ws2 := range wsList2 {
			if matchedWs2.Has(i) {
				continue
			}
			if workloadSelectorsEqual(ws1, ws2) {
				matchedWs2.Insert(i)
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// construct the callback for registering fields in the virtual service
func registerFieldFunc(
	virtualServiceFields fieldutils.FieldOwnershipRegistry,
	virtualService *networkingv1alpha3.VirtualService,
	policyRef ezkube.ResourceId,
) decorators.RegisterField {
	return func(fieldPtr, val interface{}) error {
		fieldVal := reflect.ValueOf(fieldPtr).Elem().Interface()

		if equalityutils.DeepEqual(fieldVal, val) {
			return nil
		}
		if err := virtualServiceFields.RegisterFieldOwnership(
			virtualService,
			fieldPtr,
			[]ezkube.ResourceId{policyRef},
			&v1.TrafficPolicy{},
			0, // TODO(ilackarms): priority
		); err != nil {
			return err
		}
		return nil
	}
}

func (t *translator) initializeVirtualService(
	destination *discoveryv1.Destination,
	sourceMeshInstallation *discoveryv1.MeshInstallation,
	destinationFQDN string,
) *networkingv1alpha3.VirtualService {
	var meta metav1.ObjectMeta
	if sourceMeshInstallation != nil {
		meta = metautils.FederatedObjectMeta(
			destination.Spec.GetKubeService().Ref,
			sourceMeshInstallation,
			destination.Annotations,
		)
	} else {
		meta = metautils.TranslatedObjectMeta(
			destination.Spec.GetKubeService().Ref,
			destination.Annotations,
		)
	}

	return &networkingv1alpha3.VirtualService{
		ObjectMeta: meta,
		Spec: networkingv1alpha3spec.VirtualService{
			Hosts: []string{destinationFQDN},
		},
	}
}

// Returns nil to prevent translating the trafficPolicy if the sourceClusterName is not selected by the WorkloadSelector
func initializeBaseRoute(trafficPolicy *v1.TrafficPolicySpec, sourceClusterName string) *networkingv1alpha3spec.HTTPRoute {
	if !selectorutils.WorkloadSelectorContainsCluster(trafficPolicy.SourceSelector, sourceClusterName) {
		return nil
	}
	return &networkingv1alpha3spec.HTTPRoute{
		Match: translateRequestMatchers(trafficPolicy),
	}
}

func (t *translator) setDefaultDestinationAndPortMatchers(
	baseRoute *networkingv1alpha3spec.HTTPRoute,
	destinationFQDN string,
) {
	// need a default matcher for the destination, which
	// gets populated later in duplicateRouteForEachPort()
	if len(baseRoute.Match) == 0 {
		baseRoute.Match = []*networkingv1alpha3spec.HTTPMatchRequest{{}}
	}

	// if a route destination is already set, we don't need to modify the route
	if baseRoute.Route != nil {
		return
	}

	baseRoute.Route = []*networkingv1alpha3spec.HTTPRouteDestination{{
		Destination: &networkingv1alpha3spec.Destination{
			Host: destinationFQDN,
		},
	}}
}

// construct a copy of a route for each service port
// required because Istio needs the destination port for every route
// if the service has multiple service ports defined
func duplicateRouteForEachPort(
	baseRoute *networkingv1alpha3spec.HTTPRoute,
	ports []*discoveryv1.DestinationSpec_KubeService_KubeServicePort,
) []*networkingv1alpha3spec.HTTPRoute {
	var routesWithPort []*networkingv1alpha3spec.HTTPRoute
	for _, port := range ports {
		// create a separate set of matchers for each port on the destination service
		var matchersWithPort []*networkingv1alpha3spec.HTTPMatchRequest

		for _, matcher := range baseRoute.Match {
			matcher := matcher.DeepCopy()
			matcher.Port = port.GetPort()
			matchersWithPort = append(matchersWithPort, matcher)
		}

		var destinationsWithPort []*networkingv1alpha3spec.HTTPRouteDestination

		for _, destination := range baseRoute.Route {
			destination := destination.DeepCopy()

			// don't overwrite ports that were derived from traffic shift
			if destination.GetDestination().GetPort().GetNumber() == 0 {
				destination.Destination.Port = &networkingv1alpha3spec.PortSelector{
					Number: port.GetPort(),
				}
			}

			destinationsWithPort = append(destinationsWithPort, destination)
		}

		routeWithPort := baseRoute.DeepCopy()
		routeWithPort.Match = matchersWithPort
		routeWithPort.Route = destinationsWithPort
		routesWithPort = append(routesWithPort, routeWithPort)
	}

	return routesWithPort
}

func splitRouteByMatchers(baseRoute *networkingv1alpha3spec.HTTPRoute) []*networkingv1alpha3spec.HTTPRoute {
	if len(baseRoute.Match) < 1 {
		return []*networkingv1alpha3spec.HTTPRoute{baseRoute}
	}

	var singleMatcherRoutes []*networkingv1alpha3spec.HTTPRoute
	for _, match := range baseRoute.Match {
		singleMatcherRoute := baseRoute.DeepCopy()
		singleMatcherRoute.Match = []*networkingv1alpha3spec.HTTPMatchRequest{match}
		singleMatcherRoutes = append(singleMatcherRoutes, singleMatcherRoute)
	}
	return singleMatcherRoutes
}

func translateRequestMatchers(
	trafficPolicy *v1.TrafficPolicySpec,
) []*networkingv1alpha3spec.HTTPMatchRequest {
	return routeutils.TranslateRequestMatchers(
		trafficPolicy.HttpRequestMatchers,
		trafficPolicy.SourceSelector,
	)
}

// Return errors for each user-supplied VirtualService that applies to the same hostname as the translated VirtualService
func conflictsWithUserVirtualService(
	userVirtualServices v1alpha3sets.VirtualServiceSet,
	translatedVirtualService *networkingv1alpha3.VirtualService,
) []error {
	// For each user VS, check whether any hosts match any hosts from translated VS
	var errs []error

	// virtual services from RemoteSnapshot only contain non-translated objects
	userVirtualServices.List(func(vs *networkingv1alpha3.VirtualService) (_ bool) {
		// different cluster, no conflict
		if vs.ClusterName != translatedVirtualService.ClusterName {
			return
		}

		// check if common hostnames exist
		commonHostnames := utils.CommonHostnames(vs.Spec.Hosts, translatedVirtualService.Spec.Hosts)
		if len(commonHostnames) > 0 {
			errs = append(
				errs,
				eris.Errorf("Unable to translate AppliedTrafficPolicies to VirtualService, applies to hosts %+v that are already configured by the existing VirtualService %s", commonHostnames, skv2sets.Key(vs)),
			)
		}
		return
	})

	return errs
}
