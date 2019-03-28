package surveyutils

import (
	"context"
	"fmt"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/cli/pkg/options"
)

func SurveyRoutingRule(ctx context.Context, in *options.CreateRoutingRule) error {
	if err := surveySelector(ctx, "source", &in.SourceSelector); err != nil {
		return err
	}
	if err := surveySelector(ctx, "destination", &in.DestinationSelector); err != nil {
		return err
	}

	if err := surveyMatcher(&in.RequestMatchers); err != nil {
		return err
	}

	if err := surveyMesh(ctx, &in.TargetMesh); err != nil {
		return err
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
