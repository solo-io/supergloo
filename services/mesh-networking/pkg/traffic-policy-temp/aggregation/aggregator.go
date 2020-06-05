package traffic_policy_aggregation

import (
	"strings"

	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
)

var (
	TrafficPolicyConflictError = func(s string) error { return eris.Errorf("Found conflicts in TrafficPolicy on the following fields: %s", s) }
)

func NewAggregator(resourceSelector selection.ResourceSelector) Aggregator {
	return &aggregator{
		resourceSelector: resourceSelector,
	}
}

type aggregator struct {
	resourceSelector selection.ResourceSelector
}

func (a *aggregator) FindMergeConflict(
	trafficPolicyToMerge *zephyr_networking_types.TrafficPolicySpec,
	policiesToMergeWith []*zephyr_networking_types.TrafficPolicySpec,
	meshService *zephyr_discovery.MeshService,
) *zephyr_networking_types.TrafficPolicyStatus_ConflictError {
	// The ordering of this list doesn't matter; conflict errors are reported on the list as a whole, and
	// any conflict on a Traffic Policy that is found will also be discovered on all the *other* policies that it conflicts with
	allPoliciesTogether := append([]*zephyr_networking_types.TrafficPolicySpec(nil), policiesToMergeWith...)
	allPoliciesTogether = append(allPoliciesTogether, trafficPolicyToMerge)

	// represents the merged TrafficPolicy
	var mergeableHttpTrafficPolicies []*mergeableHttpTrafficPolicy
	for _, trafficPolicy := range allPoliciesTogether {
		mergedHttpTrafficPolicies, conflictErr := a.mergeHttpTrafficPolicies(trafficPolicy, mergeableHttpTrafficPolicies)
		mergeableHttpTrafficPolicies = mergedHttpTrafficPolicies
		if conflictErr != nil {
			return &zephyr_networking_types.TrafficPolicyStatus_ConflictError{
				ErrorMessage:        conflictErr.Error(),
				ConfigurationTarget: selection.ObjectMetaToResourceRef(meshService.ObjectMeta),
			}
		}
	}
	// convert map[HttpMatcher]TrafficPolicySpec to []TrafficPolicy by consolidating equivalent TrafficPolicySpecs by appending HttpMatchers
	mergedTrafficPolicies := []*zephyr_networking.TrafficPolicy{}
	for _, mergeableHttpTp := range mergeableHttpTrafficPolicies {
		trafficPolicyExists := false
		for _, mergedTrafficPolicy := range mergedTrafficPolicies {
			// if spec containing TrafficPolicy rules already exists, with same Source selectors, just append to its HttpRequestMatchers
			if mergeableHttpTp.SourceSelector.Equal(mergedTrafficPolicy.Spec.GetSourceSelector()) &&
				a.areTrafficPolicyActionsEqual(mergeableHttpTp.TrafficPolicyRoutingConfig, &mergedTrafficPolicy.Spec) {
				a.appendHttpMatcherIfNonNil(mergedTrafficPolicy.Spec.HttpRequestMatchers, mergeableHttpTp.HttpMatcher)
				trafficPolicyExists = true
				break
			}
		}
		// Create new merged TrafficPolicySpec
		if !trafficPolicyExists {
			newMergedTrafficPolicy := &zephyr_networking.TrafficPolicy{
				Spec: *mergeableHttpTp.TrafficPolicyRoutingConfig,
			}
			mergedTrafficPolicies = append(mergedTrafficPolicies, newMergedTrafficPolicy)
			newMergedTrafficPolicy.Spec.SourceSelector = mergeableHttpTp.SourceSelector

			a.appendHttpMatcherIfNonNil(newMergedTrafficPolicy.Spec.HttpRequestMatchers, mergeableHttpTp.HttpMatcher)
		}
	}

	return nil
}

func (a *aggregator) PoliciesForService(
	trafficPolicies []*zephyr_networking.TrafficPolicy,
	meshService *zephyr_discovery.MeshService,
) (results []*zephyr_networking.TrafficPolicy, err error) {
	for _, policyIter := range trafficPolicies {
		policy := policyIter

		// we are only searching across the space of this one service, so if the resulting list is nonempty, this one must be included
		servicesForPolicy, err := a.resourceSelector.FilterMeshServicesByServiceSelector(
			[]*zephyr_discovery.MeshService{meshService},
			policy.Spec.GetDestinationSelector(),
		)
		if err != nil {
			return nil, err
		}

		if len(servicesForPolicy) > 0 {
			results = append(results, policy)
		}
	}

	return results, err
}

/*
	Only attempt to add the HttpMatcher the list if it is non-nil.
	Nil HttpRequestMatchers is a valid state in the Istio API, and therefore must
	be treated as an additional case here
*/
func (*aggregator) appendHttpMatcherIfNonNil(
	list []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher,
	nillableMatcher *zephyr_networking_types.TrafficPolicySpec_HttpMatcher,
) []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher {
	if nillableMatcher != nil {
		return append(list, nillableMatcher)
	}

	return list
}

func (a *aggregator) mergeHttpTrafficPolicies(
	trafficPolicySpec *zephyr_networking_types.TrafficPolicySpec,
	mergeableTrafficPolicies []*mergeableHttpTrafficPolicy,
) ([]*mergeableHttpTrafficPolicy, error) {
	if len(trafficPolicySpec.GetHttpRequestMatchers()) == 0 {
		mergeable, err := a.attemptTrafficPolicyMerge(trafficPolicySpec, mergeableTrafficPolicies, nil)
		if err != nil {
			return nil, err
		}
		if mergeable != nil {
			mergeableTrafficPolicies = append(mergeableTrafficPolicies, mergeable)
		}
		return mergeableTrafficPolicies, nil
	}
	// We choose the N^2 comparison over implementing a Set data structure for HTTPMatchers
	for _, httpMatcher := range trafficPolicySpec.GetHttpRequestMatchers() {
		mergeable, err := a.attemptTrafficPolicyMerge(trafficPolicySpec, mergeableTrafficPolicies, httpMatcher)
		if err != nil {
			return nil, err
		}
		if mergeable != nil {
			mergeableTrafficPolicies = append(mergeableTrafficPolicies, mergeable)
		}
	}
	return mergeableTrafficPolicies, nil
}

func (a *aggregator) attemptTrafficPolicyMerge(
	trafficPolicySpec *zephyr_networking_types.TrafficPolicySpec,
	mergeableTrafficPolicies []*mergeableHttpTrafficPolicy,
	httpMatcher *zephyr_networking_types.TrafficPolicySpec_HttpMatcher,
) (*mergeableHttpTrafficPolicy, error) {
	for _, mergeableTp := range mergeableTrafficPolicies {
		// attempt merging if Source selector and HttpMatcher are equal
		if trafficPolicySpec.SourceSelector.Equal(mergeableTp.SourceSelector) && httpMatcher.Equal(mergeableTp.HttpMatcher) {
			mergedTrafficPolicySpec, err := a.mergeTrafficPolicySpec(mergeableTp.TrafficPolicyRoutingConfig, trafficPolicySpec)
			if err != nil {
				return nil, err
			}
			// update existing TrafficPolicy with merged spec
			mergeableTp.TrafficPolicyRoutingConfig = mergedTrafficPolicySpec
			return nil, nil
		}
	}

	// copy all spec fields except HttpMatchers and Destination rules
	newTPSpec, err := a.mergeTrafficPolicySpec(&zephyr_networking_types.TrafficPolicySpec{}, trafficPolicySpec)
	if err != nil {
		return nil, err
	}
	return &mergeableHttpTrafficPolicy{
		HttpMatcher:                httpMatcher,
		SourceSelector:             trafficPolicySpec.SourceSelector,
		TrafficPolicyRoutingConfig: newTPSpec,
	}, nil
}

func (a *aggregator) mergeTrafficPolicySpec(
	this *zephyr_networking_types.TrafficPolicySpec,
	that *zephyr_networking_types.TrafficPolicySpec,
) (*zephyr_networking_types.TrafficPolicySpec, error) {
	var conflicts []string
	if this.GetTrafficShift() == nil {
		this.TrafficShift = that.TrafficShift
	} else if that.GetTrafficShift() != nil && !this.GetTrafficShift().Equal(that.GetTrafficShift()) {
		conflicts = append(conflicts, "TrafficShift")
	}
	if this.GetFaultInjection() == nil {
		this.FaultInjection = that.FaultInjection
	} else if that.GetFaultInjection() != nil && !this.GetFaultInjection().Equal(that.GetFaultInjection()) {
		conflicts = append(conflicts, "FaultInjection")
	}
	if this.GetRequestTimeout() == nil {
		this.RequestTimeout = that.RequestTimeout
	} else if that.GetRequestTimeout() != nil && !this.GetRequestTimeout().Equal(that.GetRequestTimeout()) {
		conflicts = append(conflicts, "RequestTimeout")
	}
	if this.GetRetries() == nil {
		this.Retries = that.Retries
	} else if that.GetRetries() != nil && !this.GetRetries().Equal(that.GetRetries()) {
		conflicts = append(conflicts, "Retries")
	}
	if this.GetCorsPolicy() == nil {
		this.CorsPolicy = that.CorsPolicy
	} else if that.GetCorsPolicy() != nil && !this.GetCorsPolicy().Equal(that.GetCorsPolicy()) {
		conflicts = append(conflicts, "CorsPolicy")
	}
	if this.GetMirror() == nil {
		this.Mirror = that.Mirror
	} else if that.GetMirror() != nil && !this.GetMirror().Equal(that.GetMirror()) {
		conflicts = append(conflicts, "Mirror")
	}
	if this.GetHeaderManipulation() == nil {
		this.HeaderManipulation = that.HeaderManipulation
	} else if that.GetHeaderManipulation() != nil && !this.GetHeaderManipulation().Equal(that.GetHeaderManipulation()) {
		conflicts = append(conflicts, "HeaderManipulation")
	}
	if len(conflicts) != 0 {
		return nil, TrafficPolicyConflictError(strings.Join(conflicts, ", "))
	}
	return this, nil
}

// Return true if all fields except DestinationSelector and HttpRequestMatchers are equal
func (a *aggregator) areTrafficPolicyActionsEqual(
	this *zephyr_networking_types.TrafficPolicySpec,
	that *zephyr_networking_types.TrafficPolicySpec,
) bool {
	if !this.GetTrafficShift().Equal(that.GetTrafficShift()) {
		return false
	}
	if !this.GetFaultInjection().Equal(that.GetFaultInjection()) {
		return false
	}
	if !this.GetRequestTimeout().Equal(that.GetRequestTimeout()) {
		return false
	}
	if !this.GetRetries().Equal(that.GetRetries()) {
		return false
	}
	if !this.GetCorsPolicy().Equal(that.GetCorsPolicy()) {
		return false
	}
	if !this.GetMirror().Equal(that.GetMirror()) {
		return false
	}
	if !this.GetHeaderManipulation().Equal(that.GetHeaderManipulation()) {
		return false
	}
	return true
}

type mergeableHttpTrafficPolicy struct {
	HttpMatcher                *zephyr_networking_types.TrafficPolicySpec_HttpMatcher
	SourceSelector             *zephyr_core_types.WorkloadSelector
	TrafficPolicyRoutingConfig *zephyr_networking_types.TrafficPolicySpec
}
