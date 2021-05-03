package uninstall

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Gloo Mesh from the referenced cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Verbose = globalFlags.Verbose
			return Uninstall(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

type Options struct {
	Verbose bool
	DryRun  bool

	KubeCfgPath string
	KubeContext string
	Namespace   string
	ReleaseName string
}

func (o *Options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.KubeCfgPath, &o.KubeContext, flags)
	flags.BoolVarP(&o.DryRun, "dry-run", "d", false, "Output installation manifest")
	flags.StringVar(&o.Namespace, "Namespace", defaults.DefaultPodNamespace, "Namespace in which to uninstall Gloo Mesh from")
	flags.StringVar(&o.ReleaseName, "release-name", gloomesh.GlooMeshReleaseName, "Helm release name")
}

func Uninstall(ctx context.Context, opts *Options) error {
	logrus.Info("Uninstalling Helm chart")
	if err := (helm.Uninstaller{
		KubeConfig:  opts.KubeCfgPath,
		KubeContext: opts.KubeContext,
		Namespace:   opts.Namespace,
		ReleaseName: opts.ReleaseName,
		Verbose:     opts.Verbose,
		DryRun:      opts.DryRun,
	}).UninstallChart(ctx); err != nil {
		return eris.Wrap(err, "uninstalling gloo-mesh")
	}

	return nil
}
