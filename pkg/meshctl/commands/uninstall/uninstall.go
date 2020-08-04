package uninstall

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/smh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	UinstallErr = func(err error) error {
		return eris.Wrap(err, "Error installing Service Mesh Hub")
	}
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Service Mesh Hub and clean up any associated resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			return uninstall(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())

	return cmd
}

type options struct {
	kubeCfgPath string
	kubeContext string
	namespace   string
	releaseName string
	verbose     bool
	dryRun      bool
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&o.dryRun, "dry-run", "d", false, "Output installation manifest")
	flags.StringVar(&o.kubeCfgPath, "kubeconfig", "", "path to the kubeconfig from which the master cluster will be accessed")
	flags.StringVar(&o.kubeContext, "kubecontext", "", "name of the kubeconfig context to use for the master cluster")
	flags.StringVar(&o.namespace, "namespace", "", "namespace in which to install Service Mesh Hub")
	flags.StringVar(&o.releaseName, "release-name", "service-mesh-hub", "Helm release name")
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
		KubeConfig:  opts.kubeCfgPath,
		KubeContext: opts.kubeContext,
		Namespace:   opts.namespace,
		ReleaseName: opts.releaseName,
		Verbose:     opts.verbose,
		DryRun:      opts.dryRun,
	}.UninstallServiceMeshHub(
		ctx,
	)
}
