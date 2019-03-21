package install

import (
	"github.com/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/spf13/cobra"
)

func installGlooCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gloo",
		Short: "gloo installation",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveyMetadata("installation", &opts.Metadata); err != nil {
					return err
				}
				if err := surveyutils.SurveyGlooInstall(&opts.Install); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return createGlooInstall(opts)
		},
	}
	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Metadata)
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)
	flagutils.AddGlooIngressInstallFlags(cmd.PersistentFlags(), &opts.Install)
	return cmd
}

func createGlooInstall(opts *options.Options) error {
	install, err := installGlooFromOpts(opts)
	if err != nil {
		return err
	}
	return createInstall(opts, install)
}

func installGlooFromOpts(opts *options.Options) (*v1.Install, error) {
	if err := validateGlooInstall(opts); err != nil {
		return nil, err
	}
	in := &v1.Install{
		Metadata:              opts.Metadata,
		InstallationNamespace: opts.Install.InstallationNamespace.Gloo,
		InstallType: &v1.Install_Ingress{
			Ingress: &v1.MeshIngressInstall{
				InstallType: &v1.MeshIngressInstall_Gloo{
					Gloo: &v1.GlooInstall{
						GlooVersion: opts.Install.GlooIngressInstall.GlooVersion,
						Meshes:      []string{"istio"},
					},
				},
			},
		},
	}

	return in, nil
}

func validateGlooInstall(opts *options.Options) error {
	var err error
	version := opts.Install.GlooIngressInstall.GlooVersion
	if version == "latest" {
		version, err = helpers.GetLatestVersion(opts.Ctx, "gloo")
		if err != nil {
			return errors.Wrapf(err, "unable to get latest release version from gloo")
		} else {
			opts.Install.GlooIngressInstall.GlooVersion = version
		}
	} else {
		version, err = helpers.IsValidVersion(opts.Ctx, "gloo", opts.Install.GlooIngressInstall.GlooVersion)
		if err != nil {
			return errors.Wrapf(err, "%v is not a supported gloo version", opts.Install.GlooIngressInstall.GlooVersion)
		} else {
			opts.Install.GlooIngressInstall.GlooVersion = version
		}
	}

	return nil
}
