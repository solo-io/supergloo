package check

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/checks"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	checkMarkChar = "\u2705"
	redXChar      = "\u274C"

	// TODO implement kube connectivity check

	managementPlane = checks.Category{
		Name: "Service Mesh Hub",
		Checks: []checks.Check{
			checks.NewDeploymentsCheck(),
		},
	}

	configuration = checks.Category{
		Name: "Management Configuration",
		Checks: []checks.Check{
			checks.NewNetworkingCrdCheck(),
		},
	}

	categories = []checks.Category{
		managementPlane,
		configuration,
	}
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Perform health checks on the Service Mesh Hub system",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			return runChecks(ctx, client, opts.namespace)
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
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace that Service Mesh Hub is installed in")
}

func runChecks(ctx context.Context, client client.Client, installNamespace string) error {
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
