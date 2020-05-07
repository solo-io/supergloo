package install

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/helmutil"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/semver"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register"
	"github.com/solo-io/service-mesh-hub/pkg/factories"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

type InstallCommand *cobra.Command

var (
	InstallErr = func(err error) error {
		return eris.Wrap(err, "Error installing Service Mesh Hub")
	}
	InvalidVersionErr = func(version string) error {
		return eris.Errorf(
			"Invalid version: %s. For a list of supported versions, "+
				"see the releases page: https://github.com/solo-io/service-mesh-hub/releases", version)
	}
	CannotRegisterInDryRunMode = eris.New("Cannot register the management plane cluster when --dry-run is specified")

	InstallSet = wire.NewSet(
		InstallCmd,
	)
	PreInstallMessage  = "Starting Service Mesh Hub installation...\n"
	PostInstallMessage = "Service Mesh Hub successfully installed!\n"
)

func HelmInstallerProvider(kubeClient kubernetes.Interface) factories.HelmerInstallerFactory {
	return factories.NewHelmInstallerFactory(kubeClient.CoreV1().Namespaces(), os.Stdout)
}

func InstallCmd(
	ctx context.Context,
	opts *options.Options,
	kubeClientsFactory common.KubeClientsFactory,
	clientFactory common.ClientsFactory,
	kubeLoader common_config.KubeLoader,
	out io.Writer,
) InstallCommand {
	cmd := &cobra.Command{
		Use:     cliconstants.InstallCommand.Use,
		Short:   cliconstants.InstallCommand.Short,
		PreRunE: validateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.SmhInstall.DryRun {
				// if we're running in dry-run mode, we may be piping into `kubectl apply -f -`, so don't print any of our own messages
				out = ioutil.Discard
			}

			cfg, err := kubeLoader.GetRestConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext)
			if err != nil {
				return err
			}
			kubeClients, err := kubeClientsFactory(cfg, opts.Root.WriteNamespace)
			if err != nil {
				return err
			}
			chartUri, err := helmutil.GetChartUri(opts.SmhInstall.HelmChartOverride, opts.SmhInstall.Version)
			if err != nil {
				return InstallErr(err)
			}
			helmClient := kubeClients.HelmClientFileConfigFactory(opts.Root.KubeConfig, opts.Root.KubeContext)
			if err := kubeClients.HelmInstallerFactory(helmClient).Install(&types.InstallerConfig{
				DryRun:             opts.SmhInstall.DryRun,
				Verbose:            opts.Root.Verbose,
				InstallNamespace:   opts.Root.WriteNamespace,
				CreateNamespace:    opts.SmhInstall.CreateNamespace,
				ReleaseName:        opts.SmhInstall.HelmReleaseName,
				ReleaseUri:         chartUri,
				ValuesFiles:        opts.SmhInstall.HelmChartValueFileNames,
				PreInstallMessage:  PreInstallMessage,
				PostInstallMessage: PostInstallMessage,
			}); err != nil {
				return InstallErr(err)
			}

			fmt.Fprintf(out, "Service Mesh Hub has been installed to namespace %s\n", opts.Root.WriteNamespace)

			// Register has been set by the user, therefore we will attempt to register the current cluster
			if opts.SmhInstall.Register && !opts.SmhInstall.DryRun {
				// need to fetch raw config in order to get current context to pass to registration client
				raw, err := kubeLoader.GetRawConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext)
				if err != nil {
					return err
				}
				opts.Cluster.Register = options.Register{
					RemoteClusterName:    opts.SmhInstall.ClusterName,
					RemoteWriteNamespace: opts.Root.WriteNamespace,
					RemoteContext:        raw.CurrentContext,
					RemoteKubeConfig:     opts.Root.KubeConfig,
				}
				err = register.RegisterCluster(
					ctx,
					kubeClientsFactory,
					clientFactory,
					opts,
					kubeLoader,
				)
				if err != nil {
					fmt.Printf("Error registering cluster %s: %+v", opts.SmhInstall.ClusterName, err)
				} else {
					fmt.Printf("Successfully registered cluster %s.", opts.SmhInstall.ClusterName)
				}
				return err
			} else if opts.SmhInstall.Register {
				return CannotRegisterInDryRunMode
			}
			return nil
		},
	}
	options.AddInstallFlags(cmd, opts)
	return cmd
}

func validateArgs(cmd *cobra.Command, _ []string) error {
	// validate version, prefix with 'v' if not already
	version, _ := cmd.Flags().GetString("version")
	if version != "" {
		if !semver.ValidReleaseSemver(version) {
			return InvalidVersionErr(version)
		}
	}
	return nil
}
