package cleanup

import (
	"context"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up bootstrapped local resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cleanup(ctx)
		},
	}

	return cmd
}

func cleanup(ctx context.Context) error {
	cmd := exec.Command("bash", "./ci/setup-kind.sh", "cleanup")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return err
}
