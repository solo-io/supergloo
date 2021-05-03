package routeutils

import (
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

// TranslateRequestMatchers translates request matchers to Istio. Only provide sourceNamespace and sourceLabels for translation of in-mesh VirtualServices.
func TranslateRequestMatchers(
	requestMatchers []*commonv1.HttpMatcher,
	sourceSelectors []*commonv1.WorkloadSelector, // should be nil for gateway
) []*networkingv1alpha3spec.HTTPMatchRequest {
	// Generate HttpMatchRequests for SourceSelector, one per namespace.
	var sourceMatchers []*networkingv1alpha3spec.HTTPMatchRequest
	// Set SourceNamespace and SourceLabels.
	for _, sourceSelector := range sourceSelectors {

		sourceWorkloadMatcher := sourceSelector.GetKubeWorkloadMatcher()
		if len(sourceWorkloadMatcher.GetLabels()) > 0 ||
			len(sourceWorkloadMatcher.GetNamespaces()) > 0 {
			if len(sourceWorkloadMatcher.GetNamespaces()) > 0 {
				for _, namespace := range sourceWorkloadMatcher.GetNamespaces() {
					matchRequest := &networkingv1alpha3spec.HTTPMatchRequest{
						SourceNamespace: namespace,
						SourceLabels:    sourceWorkloadMatcher.GetLabels(),
					}
					sourceMatchers = append(sourceMatchers, matchRequest)
				}
			} else {
				sourceMatchers = append(sourceMatchers, &networkingv1alpha3spec.HTTPMatchRequest{
					SourceLabels: sourceWorkloadMatcher.GetLabels(),
				})
			}
		}
	}
	// If SourceSelector is nil, generate an HttpMatchRequest without SourceSelector match criteria
	if len(sourceMatchers) == 0 {
		sourceMatchers = append(sourceMatchers, &networkingv1alpha3spec.HTTPMatchRequest{})
	}
	if requestMatchers == nil {
		return sourceMatchers
	}

	// If HttpRequestMatchers exist, generate cartesian product of sourceMatchers and httpRequestMatchers.
	var translatedRequestMatchers []*networkingv1alpha3spec.HTTPMatchRequest
	
	// Set QueryParams, Headers, WithoutHeaders, Uri, and Method.
	for _, sourceMatcher := range sourceMatchers {
		for _, matcher := range requestMatchers {
			httpMatcher := &networkingv1alpha3spec.HTTPMatchRequest{
				SourceNamespace: sourceMatcher.GetSourceNamespace(),
				SourceLabels:    sourceMatcher.GetSourceLabels(),
			}
			headerMatchers, inverseHeaderMatchers := translateRequestMatcherHeaders(matcher.GetHeaders())
			uriMatcher := translateRequestMatcherPathSpecifier(matcher)
			var method *networkingv1alpha3spec.StringMatch
			if matcher.GetMethod() != "" {
				method = &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: matcher.GetMethod()}}
			}
			httpMatcher.QueryParams = translateRequestMatcherQueryParams(matcher.GetQueryParameters())
			httpMatcher.Headers = headerMatchers
			httpMatcher.WithoutHeaders = inverseHeaderMatchers
			httpMatcher.Uri = uriMatcher
			httpMatcher.Method = method
			translatedRequestMatchers = append(translatedRequestMatchers, httpMatcher)
		}
	}
	return translatedRequestMatchers
}

func translateRequestMatcherHeaders(matchers []*commonv1.HeaderMatcher) (
	map[string]*networkingv1alpha3spec.StringMatch, map[string]*networkingv1alpha3spec.StringMatch,
) {
	headerMatchers := map[string]*networkingv1alpha3spec.StringMatch{}
	inverseHeaderMatchers := map[string]*networkingv1alpha3spec.StringMatch{}
	var matcherMap map[string]*networkingv1alpha3spec.StringMatch
	if matchers != nil {
		for _, matcher := range matchers {
			matcherMap = headerMatchers
			if matcher.GetInvertMatch() {
				matcherMap = inverseHeaderMatchers
			}
			if matcher.GetRegex() {
				matcherMap[matcher.GetName()] = &networkingv1alpha3spec.StringMatch{
					MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				matcherMap[matcher.GetName()] = &networkingv1alpha3spec.StringMatch{
					MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: matcher.GetValue()},
				}
			}
		}
	}
	// ensure field is set to nil if empty
	if len(headerMatchers) == 0 {
		headerMatchers = nil
	}
	if len(inverseHeaderMatchers) == 0 {
		inverseHeaderMatchers = nil
	}
	return headerMatchers, inverseHeaderMatchers
}

func translateRequestMatcherQueryParams(matchers []*commonv1.HttpMatcher_QueryParameterMatcher) map[string]*networkingv1alpha3spec.StringMatch {
	var translatedQueryParamMatcher map[string]*networkingv1alpha3spec.StringMatch
	if matchers != nil {
		translatedQueryParamMatcher = map[string]*networkingv1alpha3spec.StringMatch{}
		for _, matcher := range matchers {
			if matcher.GetRegex() {
				translatedQueryParamMatcher[matcher.GetName()] = &networkingv1alpha3spec.StringMatch{
					MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				translatedQueryParamMatcher[matcher.GetName()] = &networkingv1alpha3spec.StringMatch{
					MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: matcher.GetValue()},
				}
			}
		}
	}
	return translatedQueryParamMatcher
}

func translateRequestMatcherPathSpecifier(matcher *commonv1.HttpMatcher) *networkingv1alpha3spec.StringMatch {
	if matcher != nil && matcher.GetPathSpecifier() != nil {
		switch pathSpecifierType := matcher.GetPathSpecifier().(type) {
		case *commonv1.HttpMatcher_Exact:
			return &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: pathSpecifierType.Exact}}
		case *commonv1.HttpMatcher_Prefix:
			return &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Prefix{Prefix: pathSpecifierType.Prefix}}
		case *commonv1.HttpMatcher_Regex:
			return &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: pathSpecifierType.Regex}}
		}
	}
	return nil
}
