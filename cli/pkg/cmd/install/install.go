package install

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	filename  string
	meshType  string
	namespace string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: `install a mesh`,
		Long:  `install a mesh.`,
		Run: func(c *cobra.Command, args []string) {
			install()
		},
	}
	pflags := cmd.PersistentFlags()
	pflags.StringVarP(&filename, "filename", "f", "", "filename to create resources from")
	pflags.StringVarP(&meshType, "meshtype", "m", "", "mesh to install: istio, consul, linkerd")
	pflags.StringVarP(&namespace, "namespace", "n", "", "namespace to use")
	return cmd
}

func install() {
	if namespace == "" {
		fmt.Println("Please provide a namespace")
		return
	}
	if meshType == "" {
		fmt.Println("Please provide a mesh type")
		return
	}
	if filename == "" {
		fmt.Println("Please provide a filename")
		return
	}
	fmt.Printf("installing %v in namespace %v from %v\n", meshType, namespace, filename)
}
