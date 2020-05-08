package check

import (
	"context"
	"io"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/status"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	"github.com/spf13/cobra"
)

var (
	UnrecognizedPrintFormat = func(format string) error {
		return eris.Errorf("Unrecognized print format: '%s'", format)
	}
)

type CheckCommand *cobra.Command

var CheckSet = wire.NewSet(
	status.NewPrettyPrinter,
	status.NewJsonPrinter,
	CheckCmd,
)

var (
	FailedToSetUpClients = func(err error) error {
		return eris.Wrapf(err, "Failed to set up meshctl check clients")
	}
	prettyOutput       = "pretty"
	validOutputFormats = []string{prettyOutput, resource_printing.JSONFormat.String()}
)

func CheckCmd(
	ctx context.Context,
	out io.Writer,
	opts *options.Options,
	kubeClientsFactory common.KubeClientsFactory,
	clientsFactory common.ClientsFactory,
	kubeLoader kubeconfig.KubeLoader,
	prettyPrinter status.PrettyPrinter,
	jsonPrinter status.JsonPrinter,
) CheckCommand {
	cmd := &cobra.Command{
		Use:   cliconstants.CheckCommand.Use,
		Short: cliconstants.CheckCommand.Short,
		RunE: func(_ *cobra.Command, _ []string) error {
			var statusPrinter status.StatusPrinter
			switch opts.Check.OutputFormat {
			case prettyOutput:
				statusPrinter = prettyPrinter
			case resource_printing.JSONFormat.String():
				statusPrinter = jsonPrinter
			default:
				return UnrecognizedPrintFormat(opts.Check.OutputFormat)
			}

			cfg, err := kubeLoader.GetRestConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext)
			if err != nil {
				return err
			}
			kubeClients, err := kubeClientsFactory(cfg, opts.Root.WriteNamespace)
			if err != nil {
				return err
			}

			clients, err := clientsFactory(opts)
			if err != nil {
				return FailedToSetUpClients(err)
			}

			statusClient := clients.StatusClientFactory(kubeClients.HealthCheckClients)
			statusReport := statusClient.Check(ctx, opts.Root.WriteNamespace, clients.HealthCheckSuite)

			statusPrinter.Print(out, statusReport)

			if statusReport.Success {
				return nil
			} else {
				// be sure that the exit code is non-zero
				return eris.New("check failed")
			}
		},
	}

	options.AddCheckFlags(cmd, opts, prettyOutput, validOutputFormats)

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
