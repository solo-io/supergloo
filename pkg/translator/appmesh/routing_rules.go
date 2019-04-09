package appmesh

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/appmesh"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

func processTrafficShiftingRule(upstreams gloov1.UpstreamList, vnodes PodVirtualNode,
	rule *v1.TrafficShifting, out *appmesh.HttpRoute) error {
	if rule == nil || rule.Destinations == nil || len(rule.Destinations.Destinations) == 0 {
		return errors.Errorf("traffic shifting destinations cannot be missing or empty")
	}
	var weightedTargets []*appmesh.WeightedTarget

	var totalWeights uint32
	for _, dest := range rule.Destinations.Destinations {
		totalWeights += dest.Weight
	}

	var totalAppmeshWeights int64
	for i, dest := range rule.Destinations.Destinations {

		if dest.Destination == nil {
			return errors.Errorf("destination %v invalid must provide target upstream", i)
		}

		upstream, err := upstreams.Find(dest.Destination.Upstream.Strings())
		if err != nil {
			return errors.Wrapf(err, "invalid upstream destination")
		}

		vn, err := virtualNodeForUpsteam(upstream, vnodes)
		if err != nil {
			return errors.Wrapf(err, "could not find corresponding virtual node")
		}

		weight := int64(dest.Weight * 100 / totalWeights)
		totalAppmeshWeights += weight

		weightedTargets = append(weightedTargets, &appmesh.WeightedTarget{
			VirtualNode: vn.VirtualNodeName,
			Weight:      &weight,
		})

	}
	// adjust weight in case rounding error occurred
	if weightNeeded := 100 - totalAppmeshWeights; weightNeeded != 0 {
		*weightedTargets[0].Weight += weightNeeded
	}
	out.Action.WeightedTargets = weightedTargets

	return nil
}

func virtualNodeForUpsteam(upstream *gloov1.Upstream, vnodes PodVirtualNode) (*appmesh.VirtualNodeData, error) {
	host, err := utils.GetHostForUpstream(upstream)
	if err != nil {
		return nil, err
	}
	for _, vn := range vnodes {
		if *vn.Spec.ServiceDiscovery.Dns.Hostname == host {
			return vn, nil
		}
	}
	return nil, fmt.Errorf("unable to find vnode for upstream %s.%s", upstream.Metadata.Namespace, upstream.Metadata.Name)
}

func createAppmeshMatcher(rule *v1.RoutingRule) (*appmesh.HttpRouteMatch, error) {
	if len(rule.GetRequestMatchers()) != 1 {
		return nil, fmt.Errorf("appmesh requires exactly one matcher, %d found", len(rule.GetRequestMatchers()))
	}

	matcher := rule.GetRequestMatchers()[0]
	var awsMatcher *appmesh.HttpRouteMatch
	pathSpecifier := matcher.GetPathSpecifier()
	if pathSpecifier == nil {
		return nil, errors.Errorf("path specifier cannot be nil")
	}
	switch matchType := pathSpecifier.(type) {
	case *gloov1.Matcher_Prefix:
		awsMatcher = &appmesh.HttpRouteMatch{
			Prefix: &matchType.Prefix,
		}
	default:
		return nil, errors.Errorf("unsupported matcher type found, %t", matcher.GetPathSpecifier())
	}
	return awsMatcher, nil
}
