package check

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/checks"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	checkMarkChar = "\u2705"
	redXChar      = "\u274C"
)

func constructChecks(opts *options) []checks.Category {
	managementPlane := checks.Category{
		Name: "Gloo Mesh",
		Checks: []checks.Check{
			checks.NewDeploymentsCheck(),
			checks.NewEnterpriseRegistrationCheck(
				opts.kubeconfig,
				opts.kubecontext,
				opts.localPort,
				opts.remotePort,
			),
		},
	}

	configuration := checks.Category{
		Name: "Management Configuration",
		Checks: []checks.Check{
			checks.NewNetworkingCrdCheck(),
		},
	}

	return []checks.Category{
		managementPlane,
		configuration,
	}
}

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Perform health checks on the Gloo Mesh system",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.meshctlConfigPath != "" {
				config, err := utils.ParseMeshctlConfig(opts.meshctlConfigPath)
				if err == nil {
					if opts.kubeconfig == "" && opts.kubecontext == "" {
						opts.kubeconfig = config.MgmtCluster().KubeConfig
						opts.kubecontext = config.MgmtCluster().KubeContext
					}
				}
			}
			kubeClient, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			meshctlChecks := constructChecks(opts)
			return runChecks(ctx, kubeClient, opts.namespace, meshctlChecks)
		},
	}
	opts.addToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	kubeconfig        string
	kubecontext       string
	meshctlConfigPath string
	namespace         string
	localPort         uint32
	remotePort        uint32
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	utils.AddMeshctlConfigFlags(&o.meshctlConfigPath, flags)
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace that Gloo Mesh is installed in")
	flags.Uint32Var(&o.localPort, "local-port", defaults.MetricsPort, "local port used to open port-forward to enterprise mgmt pod (enterprise only)")
	flags.Uint32Var(&o.remotePort, "remote-port", defaults.MetricsPort, "remote port used to open port-forward to enterprise mgmt pod (enterprise only). set to 0 to disable checks on the mgmt server")
}

func runChecks(ctx context.Context, client client.Client, installNamespace string, categories []checks.Category) error {
	for _, category := range categories {
		fmt.Println(category.Name)
		fmt.Printf(strings.Repeat("-", len(category.Name)+3) + "\n")
		for _, check := range category.Checks {
			failure := check.Run(ctx, client, installNamespace)
			printResult(failure, check.GetDescription())
		}
		fmt.Println()
	}
	return nil
}

func printResult(failure *checks.Failure, description string) {
	if failure != nil {
		fmt.Printf("%s %s\n", redXChar, description)
		for _, err := range failure.Errors {
			fmt.Printf("  - %s\n", err.Error())
		}
	} else {
		fmt.Printf("%s %s\n", checkMarkChar, description)
	}
}
