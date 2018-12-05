package initsupergloo

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: `Initialize supergloo`,
		Long:  `Initialize supergloo`,
		Run: func(c *cobra.Command, args []string) {
			initsupergloo()
		},
	}
	return cmd
}

func initsupergloo() {

	kubectl := "kubectl"
	args := []string{"apply", "-f", "https://raw.githubusercontent.com/solo-io/supergloo/master/hack/install/supergloo.yaml"}

	cmd := exec.Command(kubectl, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		fmt.Println(err)
		return
	}

}
