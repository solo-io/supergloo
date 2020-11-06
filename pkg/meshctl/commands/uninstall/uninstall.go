package uninstall

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/codegen/helm"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/smh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Service Mesh Hub from the referenced cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return uninstall(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	kubeconfig  string
	kubecontext string
	namespace   string
	releaseName string
	verbose     bool
	dryRun      bool
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.BoolVarP(&o.dryRun, "dry-run", "d", false, "Output installation manifest")
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace in which to install Service Mesh Hub")
	flags.StringVar(&o.releaseName, "release-name", helm.Chart.Data.Name, "Helm release name")
	flags.BoolVarP(&o.verbose, "verbose", "v", false, "Enable verbose output")
}

func uninstall(ctx context.Context, opts *options) error {
	if err := uninstallServiceMeshHub(ctx, opts); err != nil {
		return eris.Wrap(err, "uninstalling service-mesh-hub")
	}
	return nil
}

func uninstallServiceMeshHub(ctx context.Context, opts *options) error {
	return smh.Uninstaller{
		KubeConfig:  opts.kubeconfig,
		KubeContext: opts.kubecontext,
		Namespace:   opts.namespace,
		ReleaseName: opts.releaseName,
		Verbose:     opts.verbose,
		DryRun:      opts.dryRun,
	}.UninstallServiceMeshHub(
		ctx,
	)
}
