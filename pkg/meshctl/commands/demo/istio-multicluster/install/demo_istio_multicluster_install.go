package install

import (
	"context"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Bootstrap a multicluster Istio demo with Service Mesh Hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return install(ctx)
		},
	}

	return cmd
}

func install(ctx context.Context) error {
	cmd := exec.Command("bash", "./ci/setup-kind.sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return err
}
