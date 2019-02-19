package istio

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/istio/plugins"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

type Translator interface {
	Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (*MeshConfig, reporter.ResourceErrors, error)
}

// A container for the entire set of config for a single istio mesh
type MeshConfig struct {
	DesinationRules v1alpha3.DestinationRuleList
	VirtualServices v1alpha3.VirtualServiceList
}

func (c *MeshConfig) Sort() {
	sort.SliceStable(c.DesinationRules, func(i, j int) bool {
		return c.DesinationRules[i].Metadata.Less(c.DesinationRules[j].Metadata)
	})
	sort.SliceStable(c.VirtualServices, func(i, j int) bool {
		return c.VirtualServices[i].Metadata.Less(c.VirtualServices[j].Metadata)
	})
}

// first create all desintation rules for all subsets of each upstream
// then we need to apply the ISTIO_MUTUAL policy depending on
// whether mtls is enabled

type translator struct {
	writeNamespace string
	plugins        []plugins.Plugin
}

func NewTranslator(writeNamespace string, plugins []plugins.Plugin) Translator {
	return &translator{writeNamespace: writeNamespace, plugins: plugins}
}

// we create a routing rule for each unique matcher
type rulesByMatcher struct {
	rules map[uint64]v1.RoutingRuleList
}

func newRulesByMatcher(rules v1.RoutingRuleList) rulesByMatcher {
	rbm := make(map[uint64]v1.RoutingRuleList)
	for _, rule := range rules {
		hash := hashutils.HashAll(
			rule.SourceSelector,
			rule.RequestMatchers,
		)
		rbm[hash] = append(rbm[hash], rule)
	}

	return rulesByMatcher{rules: rbm}
}

func (rbm rulesByMatcher) sort() []v1.RoutingRuleList {
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

type labelsPortTuple struct {
	labels map[string]string
	port   uint32
}

// produces a complete istio config
func (t *translator) Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (*MeshConfig, reporter.ResourceErrors, error) {
	meshes := snapshot.Meshes.List()
	meshGroups := snapshot.Meshgroups.List()
	upstreams := snapshot.Upstreams.List()
	routingRules := snapshot.Routingrules.List()
	encryptionRules := snapshot.Ecryptionrules.List()

	resourceErrs := make(reporter.ResourceErrors)
	resourceErrs.Accept(meshes.AsInputResources()...)
	resourceErrs.Accept(meshGroups.AsInputResources()...)
	resourceErrs.Accept(routingRules.AsInputResources()...)
	resourceErrs.Accept(encryptionRules.AsInputResources()...)

	params := plugins.Params{
		Ctx:      ctx,
		Snapshot: snapshot,
	}
	destinationHostsPortsAndLabels, err := labelsAndPortsByHost(upstreams)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "internal error: getting ports and labels from upstreams")
	}
	var destinationRules v1alpha3.DestinationRuleList
	var virtualServices v1alpha3.VirtualServiceList
	for destinationHost, destinationPortAndLabelSets := range destinationHostsPortsAndLabels {
		var labelSets []map[string]string
		for _, set := range destinationPortAndLabelSets {
			labelSets = append(labelSets, set.labels)
		}

		dr := t.makeDestinatioRuleForHost(ctx,
			params,
			destinationHost,
			labelSets,
			encryptionRules,
			resourceErrs,
		)
		destinationRules = append(destinationRules, dr)

		vs := t.makeVirtualServiceForHost(ctx,
			params,
			destinationHost,
			destinationPortAndLabelSets,
			routingRules,
			upstreams,
			resourceErrs,
		)

		virtualServices = append(virtualServices, vs)
	}

	meshConfig := &MeshConfig{
		VirtualServices: virtualServices,
		DesinationRules: destinationRules,
	}
	meshConfig.Sort()

	return meshConfig, resourceErrs, nil
}

func labelsAndPortsByHost(upstreams gloov1.UpstreamList) (map[string][]labelsPortTuple, error) {
	labelsByHost := make(map[string][]labelsPortTuple)
	for _, us := range upstreams {
		labels := utils.GetLabelsForUpstream(us)
		host, err := utils.GetHostForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting host for upstream")
		}
		port, err := utils.GetPortForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting port for upstream")
		}
		labelsByHost[host] = append(labelsByHost[host], labelsPortTuple{labels: labels, port: port})
	}
	return labelsByHost, nil
}
