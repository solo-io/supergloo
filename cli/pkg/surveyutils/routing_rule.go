package surveyutils

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/errors"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/supergloo/cli/pkg/options"
)

func SurveyRoutingRule(ctx context.Context, in *options.CreateRoutingRule) error {
	if err := surveySelector(ctx, "source", &in.SourceSelector); err != nil {
		return err
	}
	if err := surveySelector(ctx, "source", &in.DestinationSelector); err != nil {
		return err
	}

	if err := surveyMatcher(&in.RequestMatchers); err != nil {
		return err
	}

	return nil
}

const (
	SelectorType_Labels    = "Label Selector"
	SelectorType_Namespace = "Namespace Selector"
	SelectorType_Upstream  = "Upstream Selector"
)

func surveySelector(ctx context.Context, selectorName string, in *options.Selector) error {
	var enabled bool
	if err := cliutil.GetBoolInput(fmt.Sprintf("create a %v selector for this rule? ", selectorName), &enabled); err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	var selectorType string
	if err := cliutil.ChooseFromList("what kind of selector would you like to create? ", &selectorType, []string{
		SelectorType_Labels,
		SelectorType_Upstream,
		SelectorType_Namespace,
	}); err != nil {
		return err
	}
	switch selectorType {
	case SelectorType_Labels:
		fmt.Println("add key-value pairs to the selector:")
		m := map[string]string(in.SelectedLabels)
		if err := SurveyMapStringString(&m); err != nil {
			return err
		}
		in.SelectedLabels = options.MapStringStringValue(m)
	case SelectorType_Upstream:
		fmt.Println("choose upstreams for this selector")
		ups, err := SurveyUpstreams(ctx)
		if err != nil {
			return err
		}
		in.SelectedUpstreams = ups
	case SelectorType_Namespace:
		fmt.Println("choose namespaces for this selector")
		nss, err := SurveyNamespaces()
		if err != nil {
			return err
		}
		in.SelectedNamespaces = nss
	default:
		return errors.Errorf("%v is not a known selector type", selectorType)
	}

	return nil
}

const (
	pathMatch_Prefix = "prefix"
	pathMatch_Regex  = "regex"
	pathMatch_Exact  = "exact"
)

var pathMatchOptions = []string{
	pathMatch_Prefix,
	pathMatch_Regex,
	pathMatch_Exact,
}

func surveyMatcher(matchers *options.RequestMatchersValue) error {
	for {
		var addMatcher bool
		if err := cliutil.GetBoolInput("add a request matcher for this rule?", &addMatcher); err != nil {
			return err
		}
		if !addMatcher {
			return nil
		}

		var match options.RequestMatcher
		var pathType string
		if err := cliutil.ChooseFromList(
			"Choose a path match type: ",
			&pathType,
			pathMatchOptions,
		); err != nil {
			return err
		}
		if pathType == "" {
			return errors.Errorf("must specify one of %v", pathMatchOptions)
		}

		var pathMatch string
		if err := cliutil.GetStringInputDefault(
			fmt.Sprintf("What path %v should we match? ", pathType),
			&pathMatch,
			"/",
		); err != nil {
			return err
		}

		switch pathType {
		case pathMatch_Exact:
			match.PathExact = pathMatch
		case pathMatch_Regex:
			match.PathRegex = pathMatch
		case pathMatch_Prefix:
			match.PathPrefix = pathMatch
		default:
			return errors.Errorf("must specify one of %v", pathMatchOptions)
		}

		fmt.Println("add key-value pairs to match header values")
		if err := SurveyMapStringString(&match.HeaderMatcher); err != nil {
			return err
		}

		if err := cliutil.GetStringSliceInput(
			fmt.Sprintf("HTTP Method to match for this route (empty to skip)? %v", match.Methods),
			&match.Methods,
		); err != nil {
			return err
		}

		*matchers = append(*matchers, match)
	}
}
