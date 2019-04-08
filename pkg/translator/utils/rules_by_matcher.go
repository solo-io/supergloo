package utils

import (
	"sort"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/hashutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
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

type PortsForHost map[string][]uint32

func GetPortsForHost(upstreams gloov1.UpstreamList) (PortsForHost, error) {
	portsByHost := make(PortsForHost)
	for _, us := range upstreams {
		host, err := GetHostForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting host for upstream")
		}
		port, err := GetPortForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting port for upstream")
		}
		portsByHost[host] = append(portsByHost[host], port)
	}
	return portsByHost, nil
}
