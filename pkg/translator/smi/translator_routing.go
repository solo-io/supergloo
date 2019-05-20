package smi

import (
	"sort"

	splitv1alpha1 "github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/apis/split/v1alpha1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/supergloo/api/external/smi/split"
	"github.com/solo-io/supergloo/pkg/api/external/smi/split/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"

	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

type RoutingConfig struct {
	TrafficSplits v1alpha1.TrafficSplitList
}

func (c *RoutingConfig) Sort() {
	sort.SliceStable(c.TrafficSplits, func(i, j int) bool {
		return c.TrafficSplits[i].GetMetadata().Less(c.TrafficSplits[j].GetMetadata())
	})
}

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
			spec, err := convertTrafficShiftingSpec(destinationHost, ruleType.TrafficShifting, upstreams)
			if err != nil {
				return nil, err
			}
			trafficSplits = append(trafficSplits, &v1alpha1.TrafficSplit{
				TrafficSplit: split.TrafficSplit{
					ObjectMeta: kubeutils.ToKubeMeta(core.Metadata{
						Name:      rule.Metadata.Name + "-" + destinationHost,
						Namespace: rule.Metadata.Namespace,
					}),
					Spec: *spec,
				},
			})
		}
	}
	return trafficSplits, nil
}

func convertTrafficShiftingSpec(originalDestination string, spec *v1.TrafficShifting, upstreams gloov1.UpstreamList) (*splitv1alpha1.TrafficSplitSpec, error) {
	var totalWeights uint32
	for _, dest := range spec.Destinations.Destinations {
		totalWeights += dest.Weight
	}

	var backends []splitv1alpha1.TrafficSplitBackend
	remainingWeight := uint32(1000)
	for i, dest := range spec.Destinations.Destinations {
		weightMilli := dest.Weight * 1000 / totalWeights
		remainingWeight -= weightMilli
		if i == len(spec.Destinations.Destinations)-1 {
			weightMilli += remainingWeight // ensure we always get 1000 total
		}
		us, err := upstreams.Find(dest.Destination.Upstream.Strings())
		if err != nil {
			return nil, err
		}
		kubeSpec, err := utils.GetUpstreamKubeSpec(us)
		if err != nil {
			return nil, err
		}
		backends = append(backends, splitv1alpha1.TrafficSplitBackend{
			Service: utils.ServiceHost(kubeSpec.ServiceName, kubeSpec.ServiceNamespace),
			Weight:  *resource.NewMilliQuantity(int64(weightMilli), resource.DecimalSI),
		})
	}
	return &splitv1alpha1.TrafficSplitSpec{
		Service:  originalDestination,
		Backends: backends,
	}, nil
}
