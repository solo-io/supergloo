package equality

import (
	"reflect"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"k8s.io/apimachinery/pkg/util/sets"
)

/*
	return true if two TrafficPolicies contain semantically equivalent request matchers,
	which requires equivalence on the SourceSelector and HttpRequestMatcher fields.
*/
func TrafficPolicyMatchersEqual(tp1 *v1alpha2.TrafficPolicySpec, tp2 *v1alpha2.TrafficPolicySpec) bool {
	return workloadSelectorListsEqual(tp1.GetSourceSelector(), tp2.GetSourceSelector()) &&
		httpMatcherListsEqual(tp1.GetHttpRequestMatchers(), tp2.HttpRequestMatchers)
}

func workloadSelectorsEqual(ws1 *v1alpha2.WorkloadSelector, ws2 *v1alpha2.WorkloadSelector) bool {
	return reflect.DeepEqual(ws1.Labels, ws2.Labels) &&
		sets.NewString(ws1.Namespaces...).Equal(sets.NewString(ws2.Namespaces...)) &&
		sets.NewString(ws1.Clusters...).Equal(sets.NewString(ws2.Clusters...))
}

// return true if two lists of WorkloadSelectors are semantically equivalent, abstracting away order
func workloadSelectorListsEqual(wsList1 []*v1alpha2.WorkloadSelector, wsList2 []*v1alpha2.WorkloadSelector) bool {
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

func headerMatchersEqual(
	headerMatcher1 *v1alpha2.TrafficPolicySpec_HeaderMatcher,
	headerMatcher2 *v1alpha2.TrafficPolicySpec_HeaderMatcher,
) bool {
	return headerMatcher1.GetName() == headerMatcher2.GetName() &&
		headerMatcher1.GetValue() == headerMatcher2.GetValue() &&
		headerMatcher1.GetRegex() == headerMatcher2.GetRegex() &&
		headerMatcher1.GetInvertMatch() == headerMatcher2.GetInvertMatch()
}

// return true if two lists of TrafficPolicySpec_HeaderMatcher are semantically equivalent, abstracting away order
func headerMatcherListsEqual(
	headerMatcherList1 []*v1alpha2.TrafficPolicySpec_HeaderMatcher,
	headerMatcherList2 []*v1alpha2.TrafficPolicySpec_HeaderMatcher,
) bool {
	if len(headerMatcherList1) != len(headerMatcherList2) {
		return false
	}
	matchedHm2 := sets.NewInt()
	for _, hm1 := range headerMatcherList1 {
		var matched bool
		for i, hm2 := range headerMatcherList2 {
			if matchedHm2.Has(i) {
				continue
			}
			if headerMatchersEqual(hm1, hm2) {
				matchedHm2.Insert(i)
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

func queryParamMatchersEqual(
	queryParamMatcher1 *v1alpha2.TrafficPolicySpec_QueryParameterMatcher,
	queryParamMatcher2 *v1alpha2.TrafficPolicySpec_QueryParameterMatcher,
) bool {
	return queryParamMatcher1.GetName() == queryParamMatcher2.GetName() &&
		queryParamMatcher1.GetValue() == queryParamMatcher2.GetValue() &&
		queryParamMatcher1.GetRegex() == queryParamMatcher2.GetRegex()
}

// return true if two lists of TrafficPolicySpec_QueryParameterMatcher are semantically equivalent, abstracting away order
func queryParamMatcherListsEqual(
	queryParamMatcherList1 []*v1alpha2.TrafficPolicySpec_QueryParameterMatcher,
	queryParamMatcherList2 []*v1alpha2.TrafficPolicySpec_QueryParameterMatcher,
) bool {
	if len(queryParamMatcherList1) != len(queryParamMatcherList2) {
		return false
	}
	matchedQp2 := sets.NewInt()
	for _, qp1 := range queryParamMatcherList1 {
		var matched bool
		for i, qp2 := range queryParamMatcherList2 {
			if matchedQp2.Has(i) {
				continue
			}
			if queryParamMatchersEqual(qp1, qp2) {
				matchedQp2.Insert(i)
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

func httpMatchersEqual(matcher1 *v1alpha2.TrafficPolicySpec_HttpMatcher, matcher2 *v1alpha2.TrafficPolicySpec_HttpMatcher) bool {
	return matcher1.GetPrefix() == matcher2.GetPrefix() &&
		matcher1.GetExact() == matcher2.GetExact() &&
		matcher1.GetRegex() == matcher2.GetRegex() &&
		matcher1.GetMethod().GetMethod() == matcher2.GetMethod().GetMethod() &&
		headerMatcherListsEqual(matcher1.GetHeaders(), matcher2.GetHeaders()) &&
		queryParamMatcherListsEqual(matcher1.GetQueryParameters(), matcher2.GetQueryParameters())
}

// return true if two lists of TrafficPolicySpec_HttpMatcher are semantically equivalent, abstracting away order
func httpMatcherListsEqual(
	httpMatcherList1 []*v1alpha2.TrafficPolicySpec_HttpMatcher,
	httpMatcherList2 []*v1alpha2.TrafficPolicySpec_HttpMatcher,
) bool {
	if len(httpMatcherList1) != len(httpMatcherList2) {
		return false
	}
	matchedHttp2 := sets.NewInt()
	for _, http1 := range httpMatcherList1 {
		var matched bool
		for i, http2 := range httpMatcherList2 {
			if matchedHttp2.Has(i) {
				continue
			}
			if httpMatchersEqual(http1, http2) {
				matchedHttp2.Insert(i)
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
