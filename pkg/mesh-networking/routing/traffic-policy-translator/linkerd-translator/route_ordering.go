package linkerd_translator

import (
	linkerd_config "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
)

type SpecificitySortableRoutes []*linkerd_config.RouteSpec

func (b SpecificitySortableRoutes) Len() int {
	return len(b)
}

func (b SpecificitySortableRoutes) Less(i, j int) bool {
	// if the first Route matches more specific criteria than the second Route,
	// order them such that the first precedes the second (i.e. takes precedence over)
	return isRouteMatcherMoreSpecific(b[i], b[j])
}

func (b SpecificitySortableRoutes) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

/* Order decreasing by specificity of matcher. Specifically this means ordering by the following matcher fields.

1. Fewest number of match conditions
2. Longest path length of first match condition.
	NOTE: this is not a very reliable way to sort regex matchers,
	but it is preferable in order to sort for deterministic behavior
*/
func isRouteMatcherMoreSpecific(routeA, routeB *linkerd_config.RouteSpec) bool {
	// sort by the fewest number of match conditions
	if len(routeA.Condition.Any) != len(routeB.Condition.Any) {
		return len(routeA.Condition.Any) < len(routeB.Condition.Any)
	}

	// sort by the shortest path length in the fist element of the list
	// NOTE: Condition.Any[0] should never be nil as we always create a default matcher if none is provided
	return len(routeA.Condition.Any[0].PathRegex) > len(routeB.Condition.Any[0].PathRegex)
}
