package cleanup

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rotisserie/eris"

	"github.com/gobuffalo/packr"
	"github.com/spf13/cobra"
)

func Command(managementCluster string, remoteCluster string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up bootstrapped local resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cleanup(managementCluster, remoteCluster)
		},
	}

	cmd.SilenceUsage = true
	return cmd
}

func cleanup(managementCluster string, remoteCluster string) error {
	fmt.Println("Cleaning up clusters")

	box := packr.NewBox("./scripts")
	script, err := box.FindString("delete_clusters.sh")
	if err != nil {
		return eris.Wrap(err, "Error loading script")
	}

	cmd := exec.Command("bash", "-c", script, managementCluster, remoteCluster)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}
