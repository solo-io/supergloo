package appmesh

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/appmesh"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

func processTrafficShiftingRule(upstreams gloov1.UpstreamList, virtualNodes virtualNodeByHost, rule *v1.TrafficShifting) (*appmesh.HttpRouteAction, uint32, error) {
	if rule == nil || rule.Destinations == nil || len(rule.Destinations.Destinations) == 0 {
		return nil, 0, errors.Errorf("traffic shifting destinations cannot be missing or empty")
	}
	var weightedTargets []*appmesh.WeightedTarget

	var totalWeights uint32
	for _, dest := range rule.Destinations.Destinations {
		totalWeights += dest.Weight
	}

	destinationPorts := make(map[uint32]bool)
	var totalAppmeshWeights int64
	for _, dest := range rule.Destinations.Destinations {

		if dest.Destination == nil {
			return nil, 0, errors.Errorf("weighted destination for routing rule is nil")
		}

		upstream, err := upstreams.Find(dest.Destination.Upstream.Strings())
		if err != nil {
			return nil, 0, errors.Wrapf(err, "cannot route traffic to upstream %s. It either does not exist or it "+
				"does not match any pods injected with the appmesh sidecar", dest.Destination.Upstream.Strings())
		}

		host, err := utils.GetHostForUpstream(upstream)
		if err != nil {
			return nil, 0, err
		}

		// TODO: we need this info to build the listener for the virtual router that contains this route. This should
		//  come from the matcher in the routing rule, but our current matcher does not support ports, so we are forced
		//  to require unique ports on all the weighted destinations (or maybe choose one for the user?)
		port, err := utils.GetPortForUpstream(upstream)
		if err != nil {
			return nil, 0, err
		}
		destinationPorts[port] = true

		virtualNode, ok := virtualNodes[host]
		if !ok {
			// This should never happen, as we ensure that each virtual node has a DNS-discoverable hostname during the
			// initialization of the awsAppMeshConfiguration object.
			return nil, 0, errors.Errorf("could not determine Virtual Node for hostname %s", host)
		}

		weight := int64(dest.Weight * 100 / totalWeights)
		totalAppmeshWeights += weight

		weightedTargets = append(weightedTargets, &appmesh.WeightedTarget{
			VirtualNode: virtualNode.VirtualNodeName,
			Weight:      &weight,
		})
	}

	// See above to-do, we are forced to impose this restriction given our current API
	if len(destinationPorts) > 1 {
		var ports []string
		for port := range destinationPorts {
			ports = append(ports, fmt.Sprint(port))
		}
		return nil, 0, errors.Errorf("multiple ports found for weighted destinations: %s. We currently support splitting traffic "+
			"only between multiple destinations on the same port for AWS AppMesh", strings.Join(ports, ", "))
	}
	var port uint32
	for p := range destinationPorts { // only one port here
		port = p
	}

	// adjust weight in case rounding error occurred
	if weightNeeded := 100 - totalAppmeshWeights; weightNeeded != 0 {
		*weightedTargets[0].Weight += weightNeeded
	}

	return &appmesh.HttpRouteAction{
		WeightedTargets: weightedTargets,
	}, port, nil
}
