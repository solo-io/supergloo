package istio

import (
	"strings"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

// creates a destination rule for every host
// then a subset for every unique set of labels therein
func destinationRulesFromUpstreams(writeNamespace string, upstreams gloov1.UpstreamList) (v1alpha3.DestinationRuleList, error) {
	var destinationRules v1alpha3.DestinationRuleList
	labelsByHost := make(map[string][]map[string]string)
	for _, us := range upstreams {
		labels := utils.GetLabelsForUpstream(us)
		host, err := utils.GetHostForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting host for upstream")
		}
		labelsByHost[host] = append(labelsByHost[host], labels)
	}
	for host, labelSets := range labelsByHost {
		var subsets []*v1alpha3.Subset
		for _, labels := range labelSets {
			if len(labels) == 0 {
				continue
			}
			subsets = append(subsets, &v1alpha3.Subset{
				Name:   subsetName(labels),
				Labels: labels,
			})
		}
		var trafficPolicy *v1alpha3.TrafficPolicy
		destinationRules = append(destinationRules, &v1alpha3.DestinationRule{
			Metadata: core.Metadata{
				Namespace: writeNamespace,
				Name:      host,
			},
			Host:          host,
			TrafficPolicy: trafficPolicy,
			Subsets:       subsets,
		})
	}

	return destinationRules.Sort(), nil
}

func subsetName(labels map[string]string) string {
	keys, values := stringutils.KeysAndValues(labels)
	name := ""
	for i := range keys {
		name += keys[i] + "-" + values[i] + "-"
	}
	name = strings.TrimSuffix(name, "-")
	return sanitizeName(name)
}

func sanitizeName(name string) string {
	name = strings.Replace(name, ".", "-", -1)
	name = strings.Replace(name, "[", "", -1)
	name = strings.Replace(name, "]", "", -1)
	name = strings.Replace(name, ":", "-", -1)
	name = strings.Replace(name, " ", "-", -1)
	name = strings.Replace(name, "\n", "", -1)
	return name
}
