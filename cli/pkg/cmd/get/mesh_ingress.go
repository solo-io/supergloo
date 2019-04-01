package get

import (
	"fmt"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/go-utils/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	"github.com/spf13/cobra"
)

func getMeshIngressCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mesh-ingress",
		Aliases: []string{"mi"},
		Short:   "retrieve information regarding an installed mesh-ingress",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				meshIngress, err := surveyutils.SurveyMeshIngress(opts.Ctx)
				if err != nil {
					return err
				}
				opts.Metadata.Namespace = meshIngress.Namespace
				opts.Metadata.Name = meshIngress.Name
			}
			return nil
		},
	}

	cmd.AddCommand(urlCmd(opts))
	return cmd
}

func urlCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "url",
		Aliases: []string{"u"},
		Short:   "get proxy url for mesh ingress",
		RunE: func(cmd *cobra.Command, args []string) error {
			miClient := clients.MustMeshIngressClient()
			mi, err := miClient.Read(opts.Metadata.Namespace, opts.Metadata.Name, skclients.ReadOpts{})
			if err != nil {
				return errors.Errorf("could not retrieve mesh ingress %s", opts.Metadata.Ref().Key())
			}
			url, err := cliutil.GetIngressHost(&opts.GetMeshIngress.Proxy, mi.InstallationNamespace)
			if err != nil {
				return err
			}
			fmt.Println(url)
			return nil
		},
	}

	flagutils.AddProxyFlags(cmd.PersistentFlags(), &opts.GetMeshIngress)

	return cmd
}
