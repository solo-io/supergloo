package demo

import (
	"fmt"
	"os"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/demo/script"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/rand"
)

type DemoCommand *cobra.Command

var DemoSet = wire.NewSet(
	DemoRootCmd,
)

func DemoRootCmd(opts *options.Options, commandLineRunner CommandLineRunner) DemoCommand {
	demo := &cobra.Command{
		Use:   cliconstants.DemoCommand.Use,
		Short: cliconstants.DemoCommand.Short,
		RunE:  common.NonTerminalCommand(cliconstants.DemoCommand.Use),
	}

	options.AddDemoFlags(demo, opts)

	demo.AddCommand(
		basicHubInstallCmd(opts, commandLineRunner),
		istioSingleCluster(opts, commandLineRunner),
		istioMultiCluster(opts, commandLineRunner),
		linkerdSingleCluster(opts, commandLineRunner),
	)

	return demo
}

func istioSingleCluster(opts *options.Options, commandLineRunner CommandLineRunner) *cobra.Command {
	istioSingleClusterCmd := &cobra.Command{
		Use:   "istio",
		Short: "Install Service Mesh Hub to a cluster, and install/register Istio to that same cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initServiceMeshHubAndCluster(opts.Demo, commandLineRunner, true); err != nil {
				return err
			}

			if err := runBash(commandLineRunner, fmt.Sprintf(script.InstallIstio, os.Args[0])); err != nil {
				return err
			}

			return nil
		},
	}

	return istioSingleClusterCmd
}

func linkerdSingleCluster(opts *options.Options, commandLineRunner CommandLineRunner) *cobra.Command {
	linkerd := &cobra.Command{
		Use:   "linkerd",
		Short: "Install Service Mesh Hub to a cluster, then install/register a Linkerd mesh on that cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initServiceMeshHubAndCluster(opts.Demo, commandLineRunner, true); err != nil {
				return err
			}

			return runBash(commandLineRunner, script.InstallLinkerd)
		},
	}

	return linkerd
}

func istioMultiCluster(opts *options.Options, commandLineRunner CommandLineRunner) *cobra.Command {
	istioMultiClusterCmd := &cobra.Command{
		Use:   "istio-multicluster",
		Short: "Install Service Mesh Hub to a cluster, then install/register two Istio installations: one to the management plane cluster, and one to a separate cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !opts.Demo.UseKind && opts.Demo.IstioMulticluster.RemoteContextName == "" {
				return eris.New("Must provide --remote-context-name when not speicfying --use-kind")
			}

			managementPlaneClusterName := opts.Demo.ClusterName
			if managementPlaneClusterName == "" {
				managementPlaneClusterName = "management-plane-" + rand.String(5)
			}
			managementPlaneOpts := options.Demo{
				DemoLabel:   opts.Demo.DemoLabel,
				UseKind:     opts.Demo.UseKind,
				ClusterName: managementPlaneClusterName,
				DevMode:     opts.Demo.DevMode,
				ContextName: opts.Demo.ContextName,
			}
			if err := initServiceMeshHubAndCluster(managementPlaneOpts, commandLineRunner, true); err != nil {
				return err
			}

			if err := runBash(commandLineRunner, fmt.Sprintf(script.InstallIstio, os.Args[0])); err != nil {
				return err
			}

			remoteClusterName := opts.Demo.IstioMulticluster.RemoteClusterName
			if remoteClusterName == "" {
				remoteClusterName = "remote-cluster-" + rand.String(5)
			}

			remoteContext := opts.Demo.IstioMulticluster.RemoteContextName
			if remoteContext == "" {
				remoteContext = "kind-" + remoteClusterName
			}

			if opts.Demo.UseKind {
				err := runBash(commandLineRunner, fmt.Sprintf(script.InitKind, remoteClusterName))
				if err != nil {
					return err
				}
			}

			if err := runBash(commandLineRunner, fmt.Sprintf(script.SwitchKubeContext, remoteContext)); err != nil {
				return err
			}

			if err := runBash(commandLineRunner, fmt.Sprintf(script.InstallIstio, os.Args[0])); err != nil {
				return err
			}

			err := runBash(commandLineRunner, fmt.Sprintf(script.SwitchKubeContext, "kind-"+managementPlaneClusterName))
			if err != nil {
				return err
			}

			csrFlag := ""
			if opts.Demo.DevMode {
				csrFlag = "true"
			}
			kindFlag := ""
			if opts.Demo.UseKind {
				kindFlag = "true"
			}
			if err := runBash(commandLineRunner, fmt.Sprintf(script.RegisterCluster, os.Args[0], remoteClusterName, remoteContext, csrFlag, kindFlag)); err != nil {
				return err
			}

			if opts.Demo.DemoLabel != "" {
				if err := runBash(commandLineRunner, fmt.Sprintf(script.LabelNamespace, "service-mesh-hub", "solo.io/hub-demo", opts.Demo.DemoLabel)); err != nil {
					return err
				}
			}

			return nil
		},
	}

	options.AddMulticlusterDemoFlags(istioMultiClusterCmd, opts)

	return istioMultiClusterCmd
}

func basicHubInstallCmd(opts *options.Options, commandLineRunner CommandLineRunner) *cobra.Command {
	basicHubInstall := &cobra.Command{
		Use:   "basic-hub-install",
		Short: "Get Service Mesh Hub installed to a cluster, and register that cluster for management through the Hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return initServiceMeshHubAndCluster(opts.Demo, commandLineRunner, true)
		},
	}

	return basicHubInstall
}

// expected to leave the kubeconfig's currentContext in a meaningful place for future operations
func initServiceMeshHubAndCluster(demoOpts options.Demo, commandLineRunner CommandLineRunner, installServiceMeshHub bool) error {
	clusterName := demoOpts.ClusterName
	if clusterName == "" {
		clusterName = "hub-demo-" + rand.String(5)
	}

	contextName := demoOpts.ContextName
	if contextName == "" {
		cfg, err := kubeutils.GetKubeConfig("", "")
		if err != nil {
			return err
		}
		contextName = cfg.CurrentContext
	}

	if demoOpts.UseKind {
		err := runBash(commandLineRunner, fmt.Sprintf(script.InitKind, clusterName))
		if err != nil {
			return err
		}

		contextName = "kind-" + clusterName

		err = runBash(commandLineRunner, fmt.Sprintf(script.SwitchKubeContext, contextName))
		if err != nil {
			return err
		}
	}

	if installServiceMeshHub {
		if err := commandLineRunner.Run("bash", "-c", fmt.Sprintf(script.InstallServiceMeshHub, os.Args[0])); err != nil {
			return err
		}
	}

	csrFlag := ""
	if demoOpts.DevMode {
		csrFlag = "true"
	}
	kindFlag := ""
	if demoOpts.UseKind {
		kindFlag = "true"
	}
	if err := runBash(commandLineRunner, fmt.Sprintf(script.RegisterCluster, os.Args[0], clusterName, contextName, csrFlag, kindFlag)); err != nil {
		return err
	}

	if demoOpts.DemoLabel != "" {
		if err := runBash(commandLineRunner, fmt.Sprintf(script.LabelNamespace, "service-mesh-hub", "solo.io/hub-demo", demoOpts.DemoLabel)); err != nil {
			return err
		}
	}
	return nil
}

func runBash(commandLineRunner CommandLineRunner, command string) error {
	return commandLineRunner.Run("bash", "-c", command)
}
