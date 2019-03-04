package surveyutils

import (
	"context"
	"fmt"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/vektah/gqlgen/neelance/errors"
)

func SurveyRoutingRule(ctx context.Context, in *options.CreateRoutingRule) error {
	if err := surveySelector(ctx, "source", &in.SourceSelector); err != nil {
		return err
	}
	if err := surveySelector(ctx, "source", &in.DestinationSelector); err != nil {
		return err
	}

	if err := surveySpec(&in.RoutingRuleSpec); err != nil {
		return err
	}

	return nil
}

func surveySelector(ctx context.Context, selectorName string, in *options.Selector) error {
	if err := cliutil.GetBoolInput(fmt.Sprintf("create a %v selector for this rule? ", selectorName), &in.Enabled); err != nil {
		return err
	}
	if !in.Enabled {
		return nil
	}

	if err := cliutil.ChooseFromList("what kind of selector would you like to create? ", &in.SelectorType, []string{
		options.SelectorType_Labels,
		options.SelectorType_Upstream,
		options.SelectorType_Namespace,
	}); err != nil {
		return err
	}
	switch in.SelectorType {
	case options.SelectorType_Labels:
		fmt.Println("add key-value pairs to the selector:")
		if err := SurveyMapStringString(&in.SelectedLabels); err != nil {
			return err
		}
	case options.SelectorType_Upstream:
		fmt.Println("choose upstreams for this selector")
		ups, err := SurveyUpstreams(ctx)
		if err != nil {
			return err
		}
		in.SelectedUpstreams = ups
	case options.SelectorType_Namespace:
		fmt.Println("choose namespaces for this selector")
		nss, err := SurveyNamespaces()
		if err != nil {
			return err
		}
		in.SelectedNamespaces = nss
	default:
		return errors.Errorf("%v is not a known selector type", in.SelectorType)
	}

	return nil
}

func surveySpec(spec *options.RoutingRuleSpec) error {
	return nil
}
