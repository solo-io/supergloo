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
				opts.GetMeshIngress.Target = options.ResourceRefValue(meshIngress)
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
			target := opts.GetMeshIngress.Target
			mi, err := miClient.Read(target.Namespace, target.Name, skclients.ReadOpts{})
			if err != nil {
				return errors.Errorf("could not retrieve mesh ingress %s.%s", target.Namespace, target.Name)
			}
			proxy := &opts.GetMeshIngress.Proxy
			url, err := cliutil.GetIngressHost(proxy, mi.InstallationNamespace)
			if err != nil {
				return err
			}

			if proxy.Port == "http" || proxy.Port == "https" {
				fmt.Printf("%v://%v\n", proxy.Port, url)
			} else {
				fmt.Printf("%v\n", url)
			}

			return nil
		},
	}

	flagutils.AddProxyFlags(cmd.PersistentFlags(), &opts.GetMeshIngress)

	return cmd
}
