package smi

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/solo-io/supergloo/pkg/api/external/smi/split/v1alpha1"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"

	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

func createRoutingConfig(rules v1.RoutingRuleList, upstreams gloov1.UpstreamList, services kubernetes.ServiceList, resourceErrs reporter.ResourceErrors) RoutingConfig {
	var trafficSplits v1alpha1.TrafficSplitList
	for _, rule := range rules {
		splitsForRule, err := trafficSplitsForRule(rule, upstreams, services)
		if err != nil {
			resourceErrs.AddError(rule, err)
			continue
		}
		trafficSplits = append(trafficSplits, splitsForRule...)
	}
	return RoutingConfig{
		TrafficSplits: trafficSplits,
	}
}

func trafficSplitsForRule(rule *v1.RoutingRule, upstreams gloov1.UpstreamList, services kubernetes.ServiceList) (v1alpha1.TrafficSplitList, error) {
	var trafficSplits v1alpha1.TrafficSplitList
	if rule.Spec == nil {
		return nil, errors.Errorf("rule spec cannot be nil")
	}
	switch ruleType := rule.Spec.RuleType.(type) {
	case *v1.RoutingRuleSpec_TrafficShifting:
		if ruleType.TrafficShifting == nil || ruleType.TrafficShifting.Destinations == nil {
			return nil, errors.Errorf("traffic shifting destinations cannot be nil")
		}
		originalDestinations, err := utils.ServicesForSelector(rule.DestinationSelector, upstreams, services)
		if err != nil {
			return nil, err
		}
		for _, destSvc := range originalDestinations {
			destinationHost := utils.ServiceHost(destSvc.Name, destSvc.Namespace)
			trafficSplits = append(trafficSplits, &v1alpha1.TrafficSplit{
				Metadata: core.Metadata{
					Name:      rule.Metadata.Name + "-" + destinationHost,
					Namespace: rule.Metadata.Namespace,
				},
				Spec: convertTrafficShiftingSpec(destinationHost, ruleType.TrafficShifting),
			})
		}
	}
	return trafficSplits, nil
}

func convertTrafficShiftingSpec(originalDestination string, spec *v1.TrafficShifting) *v1alpha1.TrafficSplitSpec {
	var totalWeights uint32
	for _, dest := range spec.Destinations.Destinations {
		totalWeights += dest.Weight
	}

	var backends []*v1alpha1.TrafficSplitBackend
	for _, dest := range spec.Destinations.Destinations {

		backends = append(backends, &v1alpha1.TrafficSplitBackend{
			Service: utils.ServiceHost(dest.Destination.Upstream.Name, dest.Destination.Upstream.Namespace),
			Weight:  fmt.Sprintf("%vm", dest.Weight*100/totalWeights),
		})
	}
	return &v1alpha1.TrafficSplitSpec{
		Service:  originalDestination,
		Backends: backends,
	}
}
