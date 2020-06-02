package describe

import (
	"context"
	"io"
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/describe/description"
	"github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

type DescribeCommand *cobra.Command

var (
	DescribeSet = wire.NewSet(
		DescribeCmd,
	)

	validResources = sets.NewString("service", "workload")
)

func DescribeCmd(
	ctx context.Context,
	kubeLoader kubeconfig.KubeLoader,
	kubeClientsFactory common.KubeClientsFactory,
	printers common.Printers,
	opts *options.Options,
	out io.Writer,
) DescribeCommand {
	describeCommand := cliconstants.DescribeCommand(validResources.List())
	cmd := &cobra.Command{
		Use:   describeCommand.Use,
		Short: describeCommand.Short,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			masterCfg, err := kubeLoader.GetRestConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext)
			if err != nil {
				return err
			}
			masterKubeClients, err := kubeClientsFactory(masterCfg, opts.Root.WriteNamespace)
			if err != nil {
				return err
			}
			resourceDescriber := masterKubeClients.ResourceDescriber

			resourceType := args[0]
			if !validResources.Has(resourceType) {
				return eris.Errorf("Unsupported resource type: %s - expected one of [%s]", resourceType, strings.Join(validResources.List(), ", "))
			}
			resourceName, err := description.ParseResourceName(args[1])
			if err != nil {
				return err
			}
			var explorationResult *description.DescriptionResult
			var printMode table_printing.PrintMode
			switch resourceType {
			case validResources.List()[0]:
				if explorationResult, err = resourceDescriber.DescribeService(ctx, *resourceName); err != nil {
					return err
				}
				printMode = table_printing.ServicePrintMode
			case validResources.List()[1]:
				if explorationResult, err = resourceDescriber.DescribeWorkload(ctx, *resourceName); err != nil {
					return err
				}
				printMode = table_printing.WorkloadPrintMode
			default:
				err = eris.Errorf("Unrecognized resource type: %s", resourceType)
			}
			if err != nil {
				return err
			}
			if opts.Describe.Policies == description.AccessPolicies {
				printers.AccessControlPolicyPrinter.Print(out, printMode, explorationResult.Policies.AccessControlPolicies)
			} else if opts.Describe.Policies == description.TrafficPolicies {
				printers.TrafficPolicyPrinter.Print(out, printMode, explorationResult.Policies.TrafficPolicies)
			} else {
				printers.AccessControlPolicyPrinter.Print(out, printMode, explorationResult.Policies.AccessControlPolicies)
				printers.TrafficPolicyPrinter.Print(out, printMode, explorationResult.Policies.TrafficPolicies)
			}
			return nil
		},
	}
	options.AddDescribeResourceFlags(cmd, opts, description.AllPolicies,
		[]string{description.AllPolicies, description.AccessPolicies, description.TrafficPolicies})
	return cmd
}
