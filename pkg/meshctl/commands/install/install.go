package install

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/common/version"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/smh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const serviceMeshHubChartUriTemplate = "https://storage.googleapis.com/service-mesh-hub/service-mesh-hub/service-mesh-hub-%s.tgz"

var (
	InstallErr = func(err error) error {
		return eris.Wrap(err, "Error installing Service Mesh Hub")
	}
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Service Mesh Hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return install(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())

	return cmd
}

type options struct {
	dryRun bool

	kubeCfgPath     string
	kubeContext     string
	namespace       string
	chartPath       string
	chartValuesFile string
	releaseName     string
	version         string
	register        bool
	clusterName     string
	verbose         bool
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&o.dryRun, "dry-run", "d", false, "Output installation manifest")
	flags.StringVar(&o.kubeCfgPath, "kubeconfig", "", "path to the kubeconfig from which the master cluster will be accessed")
	flags.StringVar(&o.kubeContext, "kubecontext", "", "name of the kubeconfig context to use for the master cluster")
	flags.StringVar(&o.namespace, "namespace", "", "namespace in which to install Service Mesh Hub")
	flags.StringVar(&o.chartPath, "chart-file", "", "Path to a local Helm chart for installing Service Mesh Hub. If unset, this command will install Service Mesh Hub from the publicly released Helm chart.")
	flags.StringVarP(&o.chartValuesFile, "chart-values-file", "", "", "File containing value overrides for the Service Mesh Hub Helm chart")
	flags.StringVar(&o.releaseName, "release-name", "service-mesh-hub", "Helm release name")
	flags.StringVar(&o.version, "version", "", "Version to install, defaults to latest if omitted")
	flags.BoolVarP(&o.register, "register", "r", false, "Register the management plane cluster")
	flags.StringVar(&o.clusterName, "cluster-name", "management-cluster",
		"Name with which to register the management-plane cluster in Service Mesh Hub, only applies if --register is also set")
	flags.BoolVarP(&o.verbose, "verbose", "v", false, "Enable verbose output")
}

func install(ctx context.Context, opts *options) error {
	// User-specified chartPath takes precedence over specified version.
	smhChartUri := opts.chartPath
	smhVersion := opts.version
	if opts.version == "" {
		smhVersion = version.Version
	}
	if smhChartUri == "" {
		smhChartUri = fmt.Sprintf(serviceMeshHubChartUriTemplate, smhVersion)
	}

	err := smh.Installer{
		HelmChartPath:  smhChartUri,
		HelmValuesPath: opts.chartValuesFile,
		KubeConfig:     opts.kubeCfgPath,
		KubeContext:    opts.kubeContext,
		Namespace:      opts.namespace,
		ReleaseName:    opts.releaseName,
		Verbose:        opts.verbose,
		DryRun:         opts.dryRun,
	}.InstallServiceMeshHub(
		ctx,
	)

	if err != nil {
		return eris.Wrap(err, "installing service-mesh-hub")
	}

	// TODO(harveyxia) register cluster

	return nil
}
