package routerule

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
)

const (
	TrafficShifting_Rule     = "TrafficShifting"
	FaultInjection_Rule      = "FaultInjection"
	Timeout_Rule             = "Timeout"
	Retries_Rule             = "Retries"
	CorsPolicy_Rule          = "CorsPolicy"
	Mirror_Rule              = "Mirror"
	HeaderManipulaition_Rule = "HeaderManipulaition"
	USE_ALL_ROUTING_RULES    = "UseAllRoutingRules"
)

var RoutingRuleTypes = []string{
	TrafficShifting_Rule,
	FaultInjection_Rule,
	Timeout_Rule,
	Retries_Rule,
	CorsPolicy_Rule,
	Mirror_Rule,
	HeaderManipulaition_Rule,
}
var RoutingRuleDisplayName map[string]string

func init() {
	RoutingRuleDisplayName = make(map[string]string)
	RoutingRuleDisplayName[TrafficShifting_Rule] = "Traffic shifting rule"
	RoutingRuleDisplayName[FaultInjection_Rule] = "Fault injection rule"
	RoutingRuleDisplayName[Timeout_Rule] = "Timeout rule"
	RoutingRuleDisplayName[Retries_Rule] = "Retries rule"
	RoutingRuleDisplayName[CorsPolicy_Rule] = "CORs Policy rule"
	RoutingRuleDisplayName[Mirror_Rule] = "Mirror rule"
	RoutingRuleDisplayName[HeaderManipulaition_Rule] = "Header Manipulation Rule"
}

func GenerateActiveRuleList(id string) []options.MultiselectOptionBool {
	activeRoutingRuleTypes := []options.MultiselectOptionBool{}
	if id == USE_ALL_ROUTING_RULES {
		for _, rrType := range RoutingRuleTypes {
			appendRuleOption(&activeRoutingRuleTypes, rrType)
		}
	} else {
		appendRuleOption(&activeRoutingRuleTypes, id)
	}
	return activeRoutingRuleTypes
}

func appendRuleOption(list *[]options.MultiselectOptionBool, id string) {
	*list = append(*list, options.MultiselectOptionBool{
		ID:          id,
		DisplayName: RoutingRuleDisplayName[id],
		Active:      true,
	})
}
