package get

import (
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
				if err := surveyutils.SurveyMetadata("secret", &opts.Metadata); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return cmd
}
