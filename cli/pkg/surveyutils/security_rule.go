package surveyutils

import (
	"context"
	"strings"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/supergloo/cli/pkg/options"
)

func SurveySecurityRule(ctx context.Context, in *options.CreateSecurityRule) error {
	if err := surveySelector(ctx, "source", &in.SourceSelector); err != nil {
		return err
	}
	if err := surveySelector(ctx, "destination", &in.DestinationSelector); err != nil {
		return err
	}
	if err := surveyMesh(ctx, &in.TargetMesh); err != nil {
		return err
	}
	if err := surveyAllowedMethods(ctx, in); err != nil {
		return err
	}
	if err := surveyAllowedPaths(ctx, in); err != nil {
		return err
	}

	return nil
}

func surveyAllowedMethods(ctx context.Context, in *options.CreateSecurityRule) error {
	var methods string
	if err := cliutil.GetStringInput("enter a comma-separated list of HTTP methods to allow for "+
		"this rule, e.g.: GET,POST,PATCH (leave empty to allow all):", &methods); err != nil {
		return err
	}

	in.AllowedMethods = strings.Split(methods, ",")

	return nil
}

func surveyAllowedPaths(ctx context.Context, in *options.CreateSecurityRule) error {
	var paths string
	if err := cliutil.GetStringInput("enter a comma-separated list of HTTP paths to allow for "+
		"this rule, e.g.: /api,/admin,/auth (leave empty to allow all):", &paths); err != nil {
		return err
	}

	in.AllowedPaths = strings.Split(paths, ",")

	return nil
}
