package bug_report

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug/bug-report/pkg/bugreport"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"istio.io/pkg/log"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	return bugreport.Cmd(log.DefaultOptions())
}