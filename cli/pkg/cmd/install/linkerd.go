package install

import (
	"github.com/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/constants"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/spf13/cobra"
)

func installLinkerdCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "linkerd",
		Short: "install the Linkerd control plane",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveyMetadata("installation", &opts.Metadata); err != nil {
					return err
				}
				if err := surveyutils.SurveyLinkerdInstall(&opts.Install); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return createLinkerdInstall(opts)
		},
	}
	flagutils.AddLinkerdInstallFlags(cmd.PersistentFlags(), &opts.Install)
	return cmd
}

func createLinkerdInstall(opts *options.Options) error {
	install, err := installLinkerdFromOpts(opts)
	if err != nil {
		return err
	}
	return createInstall(opts, install)
}

func installLinkerdFromOpts(opts *options.Options) (*v1.Install, error) {
	if err := validateLinkerdInstall(opts.Install); err != nil {
		return nil, err
	}
	in := &v1.Install{
		Metadata:              opts.Metadata,
		InstallationNamespace: opts.Install.InstallationNamespace.Linkerd,
		InstallType: &v1.Install_Mesh{
			Mesh: &v1.MeshInstall{
				MeshInstallType: &v1.MeshInstall_LinkerdMesh{
					LinkerdMesh: &opts.Install.LinkerdInstall,
				},
			},
		},
	}

	return in, nil
}

func validateLinkerdInstall(in options.Install) error {
	var validVersion bool
	for _, ver := range constants.SupportedLinkerdVersions {
		if in.LinkerdInstall.LinkerdVersion == ver {
			validVersion = true
			break
		}
	}
	if !validVersion {
		return errors.Errorf("%v is not a supported "+
			"linkerd version", in.LinkerdInstall.LinkerdVersion)
	}

	return nil
}
