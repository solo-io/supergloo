package uninstall

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/codegen/helm"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	installhelm "github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Gloo Mesh from the referenced cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.verbose = globalFlags.Verbose
			return uninstall(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	verbose bool
	dryRun  bool

	kubeCfgPath string
	kubeContext string
	namespace   string
	releaseName string
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeCfgPath, &o.kubeContext, flags)
	flags.BoolVarP(&o.dryRun, "dry-run", "d", false, "Output installation manifest")
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace in which to uninstall Gloo Mesh from")
	flags.StringVar(&o.releaseName, "release-name", helm.Chart.Data.Name, "Helm release name")
}

func uninstall(ctx context.Context, opts *options) error {
	if err := (installhelm.Uninstaller{
		KubeConfig:  opts.kubeCfgPath,
		KubeContext: opts.kubeContext,
		Namespace:   opts.namespace,
		ReleaseName: opts.releaseName,
		Verbose:     opts.verbose,
		DryRun:      opts.dryRun,
	}).UninstallChart(ctx); err != nil {
		return eris.Wrap(err, "uninstalling gloo-mesh")
	}

	return nil
}
