package install

import (
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

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
				if err := surveyutils.SurveyGlooInstall(opts.Ctx, &opts.Install); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return createGlooInstall(opts)
		},
	}
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
				IngressInstallType: &v1.MeshIngressInstall_Gloo{
					Gloo: &opts.Install.GlooIngressInstall,
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

	if len(opts.Install.MeshIngress.Meshes) > 0 && !opts.Interactive {
		meshClient := clients.MustMeshClient()
		meshes, err := meshClient.List("", skclients.ListOpts{})
		if err != nil {
			return err
		}
		for _, v := range opts.Install.MeshIngress.Meshes {
			if _, err := meshes.Find(v.Namespace, v.Name); err != nil {
				return errors.Wrapf(err, "mesh resource %s is not a valid mesh", v.Key())
			}
			opts.Install.GlooIngressInstall.Meshes = append(opts.Install.GlooIngressInstall.Meshes, &v)
		}
	}

	return nil
}
