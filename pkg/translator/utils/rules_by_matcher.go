package utils

import (
	"sort"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/hashutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// we create a routing rule for each unique matcher
type RulesByMatcher struct {
	rules map[uint64]v1.RoutingRuleList
}

func NewRulesByMatcher(rules v1.RoutingRuleList) RulesByMatcher {
	rbm := make(map[uint64]v1.RoutingRuleList)
	for _, rule := range rules {
		hash := hashutils.HashAll(
			rule.SourceSelector,
			rule.RequestMatchers,
		)
		rbm[hash] = append(rbm[hash], rule)
	}

	return RulesByMatcher{rules: rbm}
}

func (rbm RulesByMatcher) Sort() []v1.RoutingRuleList {
	var (
		hashes       []uint64
		rulesForHash []v1.RoutingRuleList
	)
	for hash, rules := range rbm.rules {
		hashes = append(hashes, hash)
		rulesForHash = append(rulesForHash, rules)
	}
	sort.SliceStable(rulesForHash, func(i, j int) bool {
		return hashes[i] < hashes[j]
	})
	return rulesForHash
}

type LabelsPortTuple struct {
	Labels map[string]string
	Port   uint32
}

func LabelsAndPortsByHost(upstreams gloov1.UpstreamList) (map[string][]LabelsPortTuple, error) {
	labelsByHost := make(map[string][]LabelsPortTuple)
	for _, us := range upstreams {
		labels := GetLabelsForUpstream(us)
		host, err := GetHostForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting host for upstream")
		}
		port, err := GetPortForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting port for upstream")
		}
		labelsByHost[host] = append(labelsByHost[host], LabelsPortTuple{Labels: labels, Port: port})
	}
	return labelsByHost, nil
}

func RuleAppliesToDestination(destinationHost string, destinationSelector *v1.PodSelector, upstreams gloov1.UpstreamList) (bool, error) {
	if destinationSelector == nil {
		return true, nil
	}
	switch selector := destinationSelector.SelectorType.(type) {
	case *v1.PodSelector_LabelSelector_:
		// true if an upstream exists whose selector falls within the rr's selector
		// and the host in question is that upstream's host
		for _, us := range upstreams {
			hostForUpstream, err := GetHostForUpstream(us)
			if err != nil {
				return false, errors.Wrapf(err, "getting host for upstream")
			}
			// we only care about the host in question
			if destinationHost != hostForUpstream {
				continue
			}

			upstreamLabels := GetLabelsForUpstream(us)
			labelsMatch := labels.SelectorFromSet(selector.LabelSelector.LabelsToMatch).Matches(labels.Set(upstreamLabels))
			if !labelsMatch {
				continue
			}

			// we found an upstream with the correct host and labels
			return true, nil
		}
	case *v1.PodSelector_UpstreamSelector_:
		for _, ref := range selector.UpstreamSelector.Upstreams {
			us, err := upstreams.Find(ref.Strings())
			if err != nil {
				return false, err
			}
			hostForUpstream, err := GetHostForUpstream(us)
			if err != nil {
				return false, errors.Wrapf(err, "getting host for upstream")
			}
			if hostForUpstream == destinationHost {
				return true, nil
			}
		}
	case *v1.PodSelector_NamespaceSelector_:
		for _, us := range upstreams {
			hostForUpstream, err := GetHostForUpstream(us)
			if err != nil {
				return false, errors.Wrapf(err, "getting host for upstream")
			}
			// we only care about the host in question
			if destinationHost != hostForUpstream {
				continue
			}

			var usInSelectedNamespace bool
			for _, ns := range selector.NamespaceSelector.Namespaces {
				namespaceForUpstream := GetNamespaceForUpstream(us)
				if ns == namespaceForUpstream {
					usInSelectedNamespace = true
					break
				}
			}
			if !usInSelectedNamespace {
				continue
			}

			// we found an upstream with the correct host and namespace
			return true, nil

		}
	}
	return false, nil
}
