package preprocess

import (
	"context"
	"sort"

	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_v1alpha1_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	networking_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	networking_errors "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/errors"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/keys"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type trafficPolicyMerger struct {
	meshServiceSelector MeshServiceSelector
	meshClient          discovery_core.MeshClient
	trafficPolicyClient networking_core.TrafficPolicyClient
}

func NewTrafficPolicyMerger(
	meshServiceSelector MeshServiceSelector,
	meshClient discovery_core.MeshClient,
	trafficPolicyClient networking_core.TrafficPolicyClient,
) TrafficPolicyMerger {
	return &trafficPolicyMerger{
		meshServiceSelector: meshServiceSelector,
		meshClient:          meshClient,
		trafficPolicyClient: trafficPolicyClient,
	}
}

func (t *trafficPolicyMerger) MergeTrafficPoliciesForMeshServices(
	ctx context.Context,
	meshServices []*discovery_v1alpha1.MeshService,
) (map[keys.MeshServiceMultiClusterKey][]*networking_v1alpha1.TrafficPolicy, error) {
	trafficPoliciesByMeshService, err := t.getTrafficPoliciesByMeshService(ctx, meshServices)
	if err != nil {
		return nil, err
	}
	policiesByMeshService, err := mergeTrafficPoliciesByMeshService(trafficPoliciesByMeshService)
	if err != nil {
		return nil, err
	}
	return policiesByMeshService, nil
}

// Given a set of MeshServices, fetch applicable TrafficPolicies for each
func (t *trafficPolicyMerger) getTrafficPoliciesByMeshService(
	ctx context.Context,
	meshServices []*discovery_v1alpha1.MeshService,
) (map[keys.MeshServiceMultiClusterKey][]*networking_v1alpha1.TrafficPolicy, error) {
	trafficPoliciesByMeshService := map[keys.MeshServiceMultiClusterKey][]*networking_v1alpha1.TrafficPolicy{}
	// initialize map with given meshServices
	for _, meshService := range meshServices {
		meshServiceKey, err := buildKeyForMeshService(ctx, t.meshClient, meshService)
		if err != nil {
			return nil, err
		}
		trafficPoliciesByMeshService[*meshServiceKey] = []*networking_v1alpha1.TrafficPolicy{}
	}
	// List all TrafficPolicies, for each add to map if it applies to any MeshService in meshServices
	trafficPolicyList, err := t.trafficPolicyClient.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, trafficPolicy := range trafficPolicyList.Items {
		// shadow trafficPolicy to avoid overwriting memory referenced by &trafficPolicy on each iteration
		trafficPolicy := trafficPolicy
		meshServicesForTP, err := t.meshServiceSelector.GetMatchingMeshServices(
			ctx,
			trafficPolicy.Spec.GetDestinationSelector(),
		)
		if err != nil {
			return nil, err
		}
		for _, meshServiceForTP := range meshServicesForTP {
			meshServiceKey, err := buildKeyForMeshService(ctx, t.meshClient, meshServiceForTP)
			if err != nil {
				return nil, err
			}
			if trafficPolicies, ok := trafficPoliciesByMeshService[*meshServiceKey]; ok {
				trafficPolicies = append(trafficPolicies, &trafficPolicy)
				trafficPoliciesByMeshService[*meshServiceKey] = trafficPolicies
			}
		}
	}
	return trafficPoliciesByMeshService, nil
}

/*
	Merge algorithm:
		1. Sort TrafficPolicies by creation time ascending (to ensure determinism when surfacing conflicting TrafficPolicies)
		2. Merge non-conflicting policies.
		3. If conflict encountered, do not apply any of its configuration, update conflicting TrafficPolicy with CONFLICT status and continue.
*/
func mergeTrafficPoliciesByMeshService(
	trafficPoliciesByMeshService map[keys.MeshServiceMultiClusterKey][]*networking_v1alpha1.TrafficPolicy,
) (map[keys.MeshServiceMultiClusterKey][]*networking_v1alpha1.TrafficPolicy, error) {
	for meshServiceKey, trafficPolicies := range trafficPoliciesByMeshService {
		sort.Slice(trafficPolicies, func(a, b int) bool {
			timestampA := trafficPolicies[a].GetCreationTimestamp()
			timestampB := trafficPolicies[b].GetCreationTimestamp()
			return timestampA.Before(&timestampB)
		})
		// represents the merged TrafficPolicy
		var trafficPolicyHTTPMatcherPairs []*HttpMatcherTrafficPolicyPair
		for _, trafficPolicy := range trafficPolicies {
			// ignore TrafficPolicy with bad statuses
			if trafficPolicy.Status.GetComputedStatus().GetStatus() == core_types.ComputedStatus_ACCEPTED {
				updatedTrafficPolicyHTTPMatcherPairs, conflictErr := mergeHttpTrafficPolicies(trafficPolicy, trafficPolicyHTTPMatcherPairs)
				trafficPolicyHTTPMatcherPairs = updatedTrafficPolicyHTTPMatcherPairs
				if conflictErr != nil {
					return nil, conflictErr
				}
			}
		}
		// convert map[HttpMatcher]TrafficPolicySpec to []TrafficPolicy by consolidating equivalent TrafficPolicySpecs by appending HttpMatchers
		mergedTrafficPolicies := []*networking_v1alpha1.TrafficPolicy{}
		for _, trafficPolicyHTTPMatcherPair := range trafficPolicyHTTPMatcherPairs {
			trafficPolicyExists := false
			for _, mergedTrafficPolicy := range mergedTrafficPolicies {
				// TrafficPolicySpec already exists, so just append to its HttpRequestMatchers
				if areTrafficPolicyActionsEqual(trafficPolicyHTTPMatcherPair.TrafficPolicySpec, &mergedTrafficPolicy.Spec) {
					mergedTrafficPolicy.Spec.HttpRequestMatchers =
						append(mergedTrafficPolicy.Spec.GetHttpRequestMatchers(), trafficPolicyHTTPMatcherPair.HttpMatcher)
					trafficPolicyExists = true
					break
				}
			}
			// Create new merged TrafficPolicySpec
			if !trafficPolicyExists {
				newMergedTrafficPolicy := &networking_v1alpha1.TrafficPolicy{
					Spec: *trafficPolicyHTTPMatcherPair.TrafficPolicySpec,
				}
				mergedTrafficPolicies = append(mergedTrafficPolicies, newMergedTrafficPolicy)
				newMergedTrafficPolicy.Spec.HttpRequestMatchers =
					[]*networking_v1alpha1_types.HttpMatcher{trafficPolicyHTTPMatcherPair.HttpMatcher}
			}
		}
		trafficPoliciesByMeshService[meshServiceKey] = mergedTrafficPolicies
	}
	return trafficPoliciesByMeshService, nil
}

/*
	Merge traffic policy into existing set of TrafficPolicy rules (represented as trafficPolicyHTTPMatcherPairs)
	Two HTTP rules "conflict" if and only if they share an HttpMatcher (equality here is defined as an exact match for each HttpMatcher field)
	and there exists a TrafficPolicy rule field that does not equal.
*/
func mergeHttpTrafficPolicies(
	trafficPolicy *networking_v1alpha1.TrafficPolicy,
	trafficPolicyHTTPMatcherPairs []*HttpMatcherTrafficPolicyPair,
) ([]*HttpMatcherTrafficPolicyPair, error) {
	// We choose the N^2 comparison over implementing a Set data structure for HTTPMatchers
	for _, httpMatcher := range trafficPolicy.Spec.GetHttpRequestMatchers() {
		httpMatcherExists := false
		for _, trafficPolicyHTTPMatcherPair := range trafficPolicyHTTPMatcherPairs {
			if httpMatcher.Equal(trafficPolicyHTTPMatcherPair.HttpMatcher) {
				mergedTrafficPolicySpec, err := mergeTrafficPolicySpec(trafficPolicyHTTPMatcherPair.TrafficPolicySpec, &trafficPolicy.Spec)
				if err != nil {
					return nil, err
				}
				// update existing TrafficPolicy with merged spec
				trafficPolicyHTTPMatcherPair.TrafficPolicySpec = mergedTrafficPolicySpec
				httpMatcherExists = true
				break
			}
		}
		if !httpMatcherExists {
			// copy all spec fields except HttpMatchers and Destination rules
			newTPSpec, err := mergeTrafficPolicySpec(&networking_v1alpha1_types.TrafficPolicySpec{}, &trafficPolicy.Spec)
			if err != nil {
				return nil, err
			}
			trafficPolicyHTTPMatcherPairs = append(trafficPolicyHTTPMatcherPairs, &HttpMatcherTrafficPolicyPair{
				HttpMatcher:       httpMatcher,
				TrafficPolicySpec: newTPSpec,
			})
		}
	}
	return trafficPolicyHTTPMatcherPairs, nil
}

// For fields that exist in that but not this, merge into this.
// Return conflict error if any field exists in both this and that TrafficPolicySpec
func mergeTrafficPolicySpec(
	this *networking_v1alpha1_types.TrafficPolicySpec,
	that *networking_v1alpha1_types.TrafficPolicySpec,
) (*networking_v1alpha1_types.TrafficPolicySpec, error) {
	if this.GetTrafficShift() == nil {
		this.TrafficShift = that.TrafficShift
	} else if that.GetTrafficShift() != nil && !this.GetTrafficShift().Equal(that.GetTrafficShift()) {
		return nil, networking_errors.TrafficPolicyConflictError
	}
	if this.GetFaultInjection() == nil {
		this.FaultInjection = that.FaultInjection
	} else if that.GetFaultInjection() != nil && !this.GetFaultInjection().Equal(that.GetFaultInjection()) {
		return nil, networking_errors.TrafficPolicyConflictError
	}
	if this.GetRequestTimeout() == nil {
		this.RequestTimeout = that.RequestTimeout
	} else if that.GetRequestTimeout() != nil && !this.GetRequestTimeout().Equal(that.GetRequestTimeout()) {
		return nil, networking_errors.TrafficPolicyConflictError
	}
	if this.GetRetries() == nil {
		this.Retries = that.Retries
	} else if that.GetRetries() != nil && !this.GetRetries().Equal(that.GetRetries()) {
		return nil, networking_errors.TrafficPolicyConflictError
	}
	if this.GetCorsPolicy() == nil {
		this.CorsPolicy = that.CorsPolicy
	} else if that.GetCorsPolicy() != nil && !this.GetCorsPolicy().Equal(that.GetCorsPolicy()) {
		return nil, networking_errors.TrafficPolicyConflictError
	}
	if this.GetMirror() == nil {
		this.Mirror = that.Mirror
	} else if that.GetMirror() != nil && !this.GetMirror().Equal(that.GetMirror()) {
		return nil, networking_errors.TrafficPolicyConflictError
	}
	if this.GetHeaderManipulation() == nil {
		this.HeaderManipulation = that.HeaderManipulation
	} else if that.GetHeaderManipulation() != nil && !this.GetHeaderManipulation().Equal(that.GetHeaderManipulation()) {
		return nil, networking_errors.TrafficPolicyConflictError
	}
	return this, nil
}

// Return true if all fields except DestinationSelector and HttpRequestMatchers are equal
func areTrafficPolicyActionsEqual(
	this *networking_v1alpha1_types.TrafficPolicySpec,
	that *networking_v1alpha1_types.TrafficPolicySpec,
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

func buildKeyForMeshService(
	ctx context.Context,
	meshClient discovery_core.MeshClient,
	meshService *discovery_v1alpha1.MeshService,
) (*keys.MeshServiceMultiClusterKey, error) {
	mesh, err := meshClient.Get(ctx, client.ObjectKey{
		Name:      meshService.Spec.GetMesh().GetName(),
		Namespace: meshService.Spec.GetMesh().GetNamespace()})
	if err != nil {
		return nil, err
	}
	return &keys.MeshServiceMultiClusterKey{
		Name:        meshService.GetName(),
		Namespace:   meshService.GetNamespace(),
		ClusterName: mesh.Spec.GetCluster().GetName(),
	}, nil
}

type HttpMatcherTrafficPolicyPair struct {
	HttpMatcher       *networking_v1alpha1_types.HttpMatcher
	TrafficPolicySpec *networking_v1alpha1_types.TrafficPolicySpec
}
