package create

import (
	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	apierrs "github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/istio"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func createRoutingRuleCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "routingrule",
		Aliases: []string{"rr"},
		Short:   "Create a routing rule to apply to one or more meshes.",
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
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveyMetadata("Routing Rule", &opts.Metadata); err != nil {
					return err
				}
				if err := surveyutils.SurveyRoutingRule(&opts.CreateRoutingRule); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return createRoutingRule(opts)
		},
	}
	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Metadata)
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)
	addCreateRoutingRuleFlags(cmd.PersistentFlags(), &opts.Create.RoutingRule)
	return cmd
}

func createRoutingRule(opts *options.Options) error {
	// first check if install exists; if so, update and write that object
	in, err := updateDisabledInstall(opts)
	if err != nil {
		return err
	}
	if in == nil {
		in, err = installFromOpts(opts)
		if err != nil {
			return err
		}
	}
	in, err = helpers.MustInstallClient().Write(in, clients.WriteOpts{Ctx: opts.Ctx, OverwriteExisting: true})
	if err != nil {
		return err
	}

	helpers.PrintInstalls(v1.InstallList{in}, opts.OutputType)

	return nil
}

func updateDisabledInstall(opts *options.Options) (*v1.Install, error) {
	existingInstall, err := helpers.MustInstallClient().Read(opts.Install.Metadata.Namespace,
		opts.Install.Metadata.Name, clients.ReadOpts{Ctx: opts.Ctx})
	if err != nil {
		if apierrs.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !existingInstall.Disabled {
		return nil, errors.Errorf("install %v is already installed and enabled", opts.Install.Metadata)
	}
	existingInstall.Disabled = false
	return existingInstall, nil
}

func installFromOpts(opts *options.Options) (*v1.Install, error) {
	if err := validate(opts.Install.InputInstall); err != nil {
		return nil, err
	}
	in := &v1.Install{
		Metadata: opts.Install.Metadata,
		InstallType: &v1.Install_Istio_{
			Istio: &opts.Install.InputInstall.IstioInstall,
		},
	}

	return in, nil
}

func validate(in options.InputInstall) error {
	var validVersion bool
	for _, ver := range []string{
		istio.IstioVersion103,
		istio.IstioVersion105,
	} {
		if in.IstioInstall.IstioVersion == ver {
			validVersion = true
			break
		}
	}
	if !validVersion {
		return errors.Errorf("%v is not a suppported "+
			"istio version", in.IstioInstall.IstioVersion)
	}

	return nil
}

func addCreateRoutingRuleFlags(set *pflag.FlagSet, in *options.RoutingRule) {
	set.StringVar(&in.Name, "name", "", "name for the resource")
	set.StringVar(&in.Namespace, "namespace", "supergloo-system", "namespace for the resource")
}
