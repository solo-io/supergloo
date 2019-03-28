package apply

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	"github.com/ghodss/yaml"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/solo-io/go-utils/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/spf13/cobra"
)

func applySecurityRuleCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "securityrule",
		Aliases: []string{"rr"},
		Short:   "Apply a security rule to one or more meshes.",
		Long: `
Each Security Rule applies an HTTP security feature to a mesh.

Security rules implement the following semantics:

RULE:
  FOR all HTTP Requests:
  - FROM these **source pods**
  - TO these **destination pods**
  - MATCHING these **request matchers**
  APPLY this rule
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveyMetadata("Security Rule", &opts.Metadata); err != nil {
					return err
				}
				if err := surveyutils.SurveySecurityRule(opts.Ctx, &opts.CreateSecurityRule); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return applySecurityRule(opts)
		},
	}
	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Metadata)
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)
	flagutils.AddCreateSecurityRuleFlags(cmd.PersistentFlags(), &opts.CreateSecurityRule)
	flagutils.AddKubeYamlFlag(cmd.PersistentFlags(), opts)

	return cmd
}

func applySecurityRule(opts *options.Options) error {
	in, err := securityRuleFromOpts(opts)
	if err != nil {
		return err
	}
	rr := clients.MustSecurityRuleClient()
	existing, err := rr.Read(in.Metadata.Namespace, in.Metadata.Name, skclients.ReadOpts{Ctx: opts.Ctx})
	if err == nil {
		// perform update
		in.Metadata.ResourceVersion = existing.Metadata.ResourceVersion
	}

	if opts.PrintKubeYaml {
		raw, err := yaml.Marshal(v1.SecurityRuleCrd.KubeResource(in))
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

	helpers.PrintSecurityRules(v1.SecurityRuleList{in}, opts.OutputType)

	return nil
}

func securityRuleFromOpts(opts *options.Options) (*v1.SecurityRule, error) {
	ss, err := convertSelector(opts.CreateSecurityRule.SourceSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid source selector")
	}

	ds, err := convertSelector(opts.CreateSecurityRule.DestinationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid destination selector")
	}

	if opts.Metadata.Name == "" {
		return nil, errors.Errorf("name cannot be empty, provide with --name flag")
	}
	if opts.CreateSecurityRule.TargetMesh.Name == "" || opts.CreateSecurityRule.TargetMesh.Namespace == "" {
		return nil, errors.Errorf("target mesh must be specified, provide with --target-mesh flag")
	}

	ref := core.ResourceRef(opts.CreateSecurityRule.TargetMesh)

	_, meshNotFoundErr := clients.MustMeshClient().Read(ref.Namespace, ref.Name, skclients.ReadOpts{Ctx: opts.Ctx})
	if meshNotFoundErr != nil {
		return nil, meshNotFoundErr
	}

	in := &v1.SecurityRule{
		Metadata:            opts.Metadata,
		TargetMesh:          &ref,
		SourceSelector:      ss,
		DestinationSelector: ds,
		AllowedMethods:      opts.CreateSecurityRule.AllowedMethods,
		AllowedPaths:        opts.CreateSecurityRule.AllowedPaths,
	}

	return in, nil
}
