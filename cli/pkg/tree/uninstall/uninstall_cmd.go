package uninstall

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UninstallCommand *cobra.Command

var (
	UninstallSet = wire.NewSet(
		UninstallCmd,
	)

	FailedToSetUpUninstallClient = func(err error) error {
		return eris.Wrap(err, "Unexpected error while setting up Helm uninstall client")
	}
	FailedToRemoveNamespace = func(err error, namespace string) error {
		return eris.Wrapf(err, "Failed to remove namespace %s", namespace)
	}
	FailedToDeRegisterCluster = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to de-register cluster %s", clusterName)
	}

	// helm isn't using Eris, so we do the best we can with their error message
	ReleaseNotFoundHelmErrorMessage = "release: not found"
)

func UninstallCmd(
	ctx context.Context,
	out io.Writer,
	opts *options.Options,
	kubeClientsFactory common.KubeClientsFactory,
	kubeLoader common_config.KubeLoader,
) UninstallCommand {
	cmd := &cobra.Command{
		Use:   cliconstants.UninstallCommand.Use,
		Short: cliconstants.UninstallCommand.Short,
		RunE: func(cmd *cobra.Command, args []string) error {
			uninstallErrorOccured := false

			masterCfg, masterKubeClients, err := buildMasterKubeClients(opts, kubeLoader, kubeClientsFactory)
			if err != nil {
				return err
			}

			// do this first- cleaning up the management plane is the highest priority
			err = cleanUpManagementPlaneComponents(out, masterKubeClients, opts)
			if err != nil {
				if strings.Contains(err.Error(), ReleaseNotFoundHelmErrorMessage) {
					fmt.Fprintf(out, "Management plane components are not running here...\n")
				} else {
					uninstallErrorOccured = true
					fmt.Fprintf(out, "Management plane components not removed - Continuing...\n\t(%s)\n", err.Error())
				}
			}

			// find all the kube clusters that we need to de-register
			kubeClusters, err := masterKubeClients.KubeClusterClient.ListKubernetesCluster(ctx, client.InNamespace(opts.Root.WriteNamespace))

			// List will only return a NoMatch error if the CRD is not registered
			// if there are no resources, then List returns an object with an empty .Items field and a nil error
			if (err == nil && len(kubeClusters.Items) == 0) || meta.IsNoMatchError(err) || errors.IsNotFound(err) {
				fmt.Fprintf(out, "No clusters to deregister...\n")
			} else if err != nil {
				fmt.Fprintf(out, "Failed to find registered clusters - Continuing...\n\t(%s)\n", err.Error())

				// failed to find the clusters, but continue through to the other steps, making this one a no-op
				kubeClusters = &zephyr_discovery.KubernetesClusterList{}
				uninstallErrorOccured = true
			} else {
				// can only do this step if we definitely have kube clusters to read from
				err = deregisterClusters(ctx, out, kubeClusters, masterKubeClients)
				if err != nil {
					uninstallErrorOccured = true
					fmt.Fprintf(out, "Failed to de-register all clusters - Continuing...\n\t(%s)\n", err.Error())
				}
			}

			// optionally clean up the management plane namespace
			err = removeManagementPlaneNamespace(ctx, out, masterKubeClients, opts)
			if err != nil {
				uninstallErrorOccured = true
				fmt.Fprintf(out, "Failed to remove management plane namespace - Continuing...\n\t(%s)\n", err.Error())
			}

			fmt.Println("About to remove zephyr CRDs from management plane")

			// remove all SMH CRDs from management plane cluster
			deletedCrds, err := masterKubeClients.UninstallClients.CrdRemover.RemoveZephyrCrds(ctx, "management plane cluster", masterCfg)
			if err == nil && deletedCrds {
				fmt.Fprintf(out, "Service Mesh Hub CRDs have been de-registered from the management plane...\n")
			} else if err == nil && !deletedCrds {
				fmt.Fprintf(out, "No CRDs to remove from the management plane cluster...\n")
			} else {
				uninstallErrorOccured = true
				fmt.Fprintf(out, "Failed to remove CRDs from management plane - Continuing...\n\t(%s)\n", err.Error())
			}

			messageSuffix := ""
			if uninstallErrorOccured {
				messageSuffix = " with errors"
			}

			fmt.Fprintf(out, "\nService Mesh Hub has been uninstalled%s\n", messageSuffix)
			if uninstallErrorOccured {
				return eris.New("errors occurred")
			}

			return nil
		},
	}

	options.AddUninstallFlags(cmd, opts)

	// This is due to  a limitation of cobra; when `meshctl check` fails, we want to
	// have the exit code of the process be nonzero. That's done in cobra by returning
	// an error out of the command handler. But by default, that causes the usage message
	// to be printed. So instead we opt, for this command, to not display usage when
	// an error is returned from normal operation. This does not prevent usage from being
	// printed by --help, however. Also note that we can't just os.Exit(1) in the handler
	// because that would kill ginkgo in a test bed environment.
	cmd.SilenceUsage = true

	return cmd
}

func buildMasterKubeClients(opts *options.Options, kubeLoader common_config.KubeLoader, kubeClientsFactory common.KubeClientsFactory) (*rest.Config, *common.KubeClients, error) {
	masterCfg, err := kubeLoader.GetRestConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext)
	if err != nil {
		return nil, nil, common.FailedLoadingMasterConfig(err)
	}
	masterKubeClients, err := kubeClientsFactory(masterCfg, opts.Root.WriteNamespace)
	if err != nil {
		return nil, nil, common.FailedLoadingMasterConfig(err)
	}
	return masterCfg, masterKubeClients, nil
}

func cleanUpManagementPlaneComponents(out io.Writer, masterKubeClients *common.KubeClients, opts *options.Options) error {
	uninstaller, err := masterKubeClients.HelmClient.NewUninstall(opts.Root.KubeContext, opts.Root.KubeContext, opts.Root.WriteNamespace)
	if err != nil {
		return FailedToSetUpUninstallClient(err)
	}
	_, err = uninstaller.Run(opts.SmhUninstall.ReleaseName)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Service Mesh Hub management plane components have been removed...\n")

	return nil
}

func deregisterClusters(ctx context.Context, out io.Writer, kubeClusters *zephyr_discovery.KubernetesClusterList, masterKubeClients *common.KubeClients) error {
	if len(kubeClusters.Items) == 0 {
		// don't want to print anything out in this case
		return nil
	}

	fmt.Fprintf(out, "Starting to de-register %d cluster(s). This may take a moment...\n", len(kubeClusters.Items))

	for _, kubeCluster := range kubeClusters.Items {
		err := masterKubeClients.ClusterDeregistrationClient.Run(ctx, &kubeCluster)
		if err != nil {
			return FailedToDeRegisterCluster(err, kubeCluster.GetName())
		}
	}

	fmt.Fprintf(out, "All clusters have been de-registered...\n")
	return nil
}

func removeManagementPlaneNamespace(ctx context.Context, out io.Writer, masterKubeClients *common.KubeClients, opts *options.Options) error {
	if opts.SmhUninstall.RemoveNamespace {

		ns, err := masterKubeClients.NamespaceClient.GetNamespace(ctx, client.ObjectKey{Name: opts.Root.WriteNamespace})

		// if the namespace is already gone then we shouldn't report an error
		if errors.IsNotFound(err) {
			return nil
		} else if err != nil {
			return FailedToRemoveNamespace(err, opts.Root.WriteNamespace)
		}

		if err = masterKubeClients.NamespaceClient.DeleteNamespace(ctx, client.ObjectKey{Name: ns.GetName()}); err != nil {
			return FailedToRemoveNamespace(err, opts.Root.WriteNamespace)
		}

		fmt.Fprintf(out, "Service Mesh Hub management plane namespace has been removed...\n")
	}

	return nil
}
