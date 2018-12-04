package uninstall

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/common"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: `Uninstall a mesh`,
		Long:  `Uninstall a mesh.`,
		Run: func(c *cobra.Command, args []string) {
			uninstall(opts)
		},
	}

	uop := &opts.Uninstall
	pflags := cmd.PersistentFlags()

	pflags.StringVarP(&uop.MeshType, "meshtype", "m", "", "mesh to uninstall: istio, consul, linkerd2")
	pflags.StringVarP(&uop.MeshNames, "meshnames", "n", "", "list of comma separated names")
	pflags.BoolVar(&uop.All, "all", false, "uninstall all")
	return cmd
}

func uninstall(opts *options.Options) {

	installClient, err := common.GetInstallClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = validateArgs(opts, installClient)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}
