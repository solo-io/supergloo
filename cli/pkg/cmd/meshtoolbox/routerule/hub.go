package routerule

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	superglooV1 "github.com/solo-io/supergloo/pkg/api/v1"
	"gopkg.in/AlecAivazis/survey.v1"
)

func AssembleRoutingRule(ruleTypeID string, activeRuleTypes *[]options.MultiselectOptionBool, opts *options.Options) error {

	if err := EnsureMinimumRequiredParams(opts); err != nil {
		return err
	}

	rrOpts := &(opts.Create).InputRoutingRule
	rrOpts.ActiveTypes = GenerateActiveRuleList(ruleTypeID)

	// if they are using the full "create" workflow the user first specifies
	// which rules to apply
	if ruleTypeID == USE_ALL_ROUTING_RULES {
		if err := EnsureActiveRoutingRuleTypes(&rrOpts.ActiveTypes, opts.Top.Static); err != nil {
			return err
		}
	}

	// Initialize the root of our RoutingRule with the minimal required params
	// TODO(mitchdraft) move these fields out s.t. they are populated by the ensure methods
	opts.MeshTool.RoutingRule = superglooV1.RoutingRule{
		Metadata: core.Metadata{
			Name:      rrOpts.RouteName,
			Namespace: rrOpts.TargetMesh.Namespace,
		},
		TargetMesh:      &rrOpts.TargetMesh,
		Sources:         opts.MeshTool.RoutingRule.Sources,
		Destinations:    opts.MeshTool.RoutingRule.Destinations,
		RequestMatchers: opts.MeshTool.RoutingRule.RequestMatchers,
	}

	for _, rrType := range *activeRuleTypes {
		if rrType.Active {
			if err := applyRule(rrType.ID, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

// TODO(mitchdraft) add the rest of the routing rules here
func applyRule(id string, opts *options.Options) error {
	irOpts := opts.Create.InputRoutingRule
	switch id {
	case TrafficShifting_Rule:
		return EnsureTrafficShifting(&irOpts.TrafficShifting, opts)
	case Timeout_Rule:
		return EnsureTimeout(opts)
	case Retries_Rule:
		return EnsureRetry(&irOpts.Retry, opts)
	case FaultInjection_Rule:
		return EnsureFault(&irOpts.FaultInjection, opts)
	case CorsPolicy_Rule:
		return EnsureCors(&irOpts.Cors, opts)
	case Mirror_Rule:
		return EnsureMirror(&irOpts.Mirror, opts)
	case HeaderManipulaition_Rule:
		return EnsureHeaderManipulation(&irOpts.HeaderManipulation, opts)
	default:
		return fmt.Errorf("Unknown routing rule type %v", id)
	}
	return nil
}

func EnsureActiveRoutingRuleTypes(active *[]options.MultiselectOptionBool, staticMode bool) error {
	if staticMode {
		// this function is irrelevant in static mode
		return nil
	}
	return selectRoutingRules(active)
}

func selectRoutingRules(list *[]options.MultiselectOptionBool) error {
	var optionsList []string
	for i, l := range *list {
		// construct the options
		optionsList = append(optionsList, fmt.Sprintf("%v. %v", i, l.DisplayName))
		// set the starting value to false
		// must use long form to edit the list
		(*list)[i].Active = false
	}
	question := &survey.MultiSelect{
		Message: fmt.Sprintf("Select which rules to apply"),
		Options: optionsList,
	}

	var choice []string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return err
	}

	for _, c := range choice {
		// extract index from user choice
		parts := strings.SplitN(c, ".", 2)
		index, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}
		(*list)[index].Active = true
	}

	return nil
}
