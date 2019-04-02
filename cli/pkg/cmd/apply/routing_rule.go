package apply

import (
	"context"
	"fmt"
	"sort"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	"github.com/ghodss/yaml"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var routingRuleTypes = []routingRuleSpecCommand{
	{
		use:   "trafficshifting",
		alias: "ts",
		short: "apply a traffic shifting rule",
		long: `Traffic Shifting rules are used to divert HTTP requests sent within the mesh from their original destinations. 
This can be used to force traffic to be sent to a specific subset of a service, a different service entirely, and/or 
be load-balanced by weight across a variety of destinations`,
		specSurveyFunc: surveyutils.SurveyTrafficShiftingSpec,
		addFlagsFunc:   flagutils.AddTrafficShiftingFlags,
		convertSpecFunc: func(in options.RoutingRuleSpec) (*v1.RoutingRuleSpec, error) {
			if in.TrafficShifting.Destinations == nil || len(in.TrafficShifting.Destinations.Destinations) == 0 {
				return nil, errors.Errorf("must provide at least 1 destination")
			}
			return &v1.RoutingRuleSpec{
				RuleType: &v1.RoutingRuleSpec_TrafficShifting{
					TrafficShifting: &v1.TrafficShifting{
						Destinations: in.TrafficShifting.Destinations,
					},
				},
			}, nil
		},
	},
	faultInjectionSpecCommand,
}

func applyRoutingRuleCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "routingrule",
		Aliases: []string{"rr"},
		Short:   "Apply a routing rule to one or more meshes.",
		Long: `
Each Routing Rule applies an HTTP routing feature to a mesh.

Routing rules implement the following semantics:

RULE:
  FOR all HTTP Requests:
  - FROM these **source pods**
  - TO these **destination pods**
  - MATCHING these **request matchers**
  APPLY this rule
`,
	}
	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Metadata)
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)
	flagutils.AddCreateRoutingRuleFlags(cmd.PersistentFlags(), &opts.CreateRoutingRule)
	flagutils.AddKubeYamlFlag(cmd.PersistentFlags(), opts)

	for _, rrType := range routingRuleTypes {
		cmd.AddCommand(createRoutingRuleSubcmd(rrType, opts))
	}

	return cmd
}

type routingRuleSpecCommand struct {
	use             string
	alias           string
	short           string
	long            string
	subCmds         []routingRuleSpecCommand
	mutateOpts      func(opts *options.Options)
	specSurveyFunc  func(ctx context.Context, in *options.CreateRoutingRule) error
	addFlagsFunc    func(set *pflag.FlagSet, in *options.RoutingRuleSpec)
	convertSpecFunc func(in options.RoutingRuleSpec) (*v1.RoutingRuleSpec, error)
}

func createRoutingRuleSubcmd(subCmd routingRuleSpecCommand, opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     subCmd.use,
		Aliases: []string{subCmd.alias},
		Short:   subCmd.alias,
		Long:    subCmd.long,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if subCmd.mutateOpts != nil {
				subCmd.mutateOpts(opts)
			}
			if opts.Interactive {
				if err := surveyutils.SurveyMetadata("Routing Rule", &opts.Metadata); err != nil {
					return err
				}
				if err := surveyutils.SurveyRoutingRule(opts.Ctx, &opts.CreateRoutingRule); err != nil {
					return err
				}
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := subCmd.specSurveyFunc(opts.Ctx, &opts.CreateRoutingRule); err != nil {
					return err
				}
			}
			return nil
		},
	}
	if subCmd.convertSpecFunc != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			spec, err := subCmd.convertSpecFunc(opts.CreateRoutingRule.RoutingRuleSpec)
			if err != nil {
				return err
			}
			return applyRoutingRule(opts, spec)
		}
	}
	if subCmd.addFlagsFunc != nil {
		subCmd.addFlagsFunc(cmd.PersistentFlags(), &opts.CreateRoutingRule.RoutingRuleSpec)
	}

	for _, v := range subCmd.subCmds {
		cmd.AddCommand(createRoutingRuleSubcmd(v, opts))
	}
	return cmd
}

func applyRoutingRule(opts *options.Options, spec *v1.RoutingRuleSpec) error {
	in, err := routingRuleFromOpts(opts)
	if err != nil {
		return err
	}
	in.Spec = spec

	rr := clients.MustRoutingRuleClient()
	existing, err := rr.Read(in.Metadata.Namespace, in.Metadata.Name, skclients.ReadOpts{Ctx: opts.Ctx})
	if err == nil {
		// perform update
		in.Metadata.ResourceVersion = existing.Metadata.ResourceVersion
	}

	if opts.PrintKubeYaml {
		raw, err := yaml.Marshal(v1.RoutingRuleCrd.KubeResource(in))
		if err != nil {
			return err
		}
		fmt.Println(string(raw))
		return nil
	}

	in, err = rr.Write(in, skclients.WriteOpts{Ctx: opts.Ctx, OverwriteExisting: true})
	if err != nil {
		return err
	}

	helpers.PrintRoutingRules(v1.RoutingRuleList{in}, opts.OutputType)

	return nil
}

func routingRuleFromOpts(opts *options.Options) (*v1.RoutingRule, error) {
	matchers, err := convertMatchers(opts.CreateRoutingRule.RequestMatchers)
	if err != nil {
		return nil, err
	}

	ss, err := ConvertSelector(opts.CreateRoutingRule.SourceSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid source selector")
	}

	ds, err := ConvertSelector(opts.CreateRoutingRule.DestinationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid destination selector")
	}

	if opts.Metadata.Name == "" {
		return nil, errors.Errorf("name cannot be empty, provide with --name flag")
	}
	if opts.CreateRoutingRule.TargetMesh.Name == "" || opts.CreateRoutingRule.TargetMesh.Namespace == "" {
		return nil, errors.Errorf("target mesh must be specified, provide with --target-mesh flag")
	}

	ref := core.ResourceRef(opts.CreateRoutingRule.TargetMesh)

	_, meshNotFoundErr := clients.MustMeshClient().Read(ref.Namespace, ref.Name, skclients.ReadOpts{Ctx: opts.Ctx})
	if meshNotFoundErr != nil {
		return nil, meshNotFoundErr
	}

	in := &v1.RoutingRule{
		Metadata:            opts.Metadata,
		TargetMesh:          &ref,
		SourceSelector:      ss,
		DestinationSelector: ds,
		RequestMatchers:     matchers,
	}

	return in, nil
}

func convertMatchers(in options.RequestMatchersValue) ([]*gloov1.Matcher, error) {
	var matchers []*gloov1.Matcher
	for _, match := range in {
		converted, err := matcherFromInput(match)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, converted)
	}
	return matchers, nil
}

var invalidPathsErr = errors.Errorf("can only set one of path-regex, path-prefix, or path-exact")

func matcherFromInput(input options.RequestMatcher) (*gloov1.Matcher, error) {
	m := &gloov1.Matcher{}
	switch {
	case input.PathExact != "":
		if input.PathRegex != "" || input.PathPrefix != "" {
			return nil, invalidPathsErr
		}
		m.PathSpecifier = &gloov1.Matcher_Exact{
			Exact: input.PathExact,
		}
	case input.PathRegex != "":
		if input.PathExact != "" || input.PathPrefix != "" {
			return nil, invalidPathsErr
		}
		m.PathSpecifier = &gloov1.Matcher_Regex{
			Regex: input.PathRegex,
		}
	case input.PathPrefix != "":
		if input.PathExact != "" || input.PathRegex != "" {
			return nil, invalidPathsErr
		}
		m.PathSpecifier = &gloov1.Matcher_Prefix{
			Prefix: input.PathPrefix,
		}
	default:
		return nil, errors.Errorf("must provide path prefix, path exact, or path regex for route matcher")
	}
	if len(input.Methods) > 0 {
		m.Methods = input.Methods
	}
	for k, v := range input.HeaderMatcher {
		m.Headers = append(m.Headers, &gloov1.HeaderMatcher{
			Name:  k,
			Value: v,
			Regex: true,
		})
	}
	sort.SliceStable(m.Headers, func(i, j int) bool {
		return m.Headers[i].Name < m.Headers[j].Name
	})
	return m, nil
}
