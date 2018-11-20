package install

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: `Install a mesh`,
		Long:  `Install a mesh.`,
		Run: func(c *cobra.Command, args []string) {
			install(opts)
		},
	}
	iop := &opts.Install
	pflags := cmd.PersistentFlags()
	// TODO(mitchdraft) - remove filename or apply it to something
	pflags.StringVarP(&iop.Filename, "filename", "f", "", "filename to create resources from")
	pflags.StringVarP(&iop.MeshType, "meshtype", "m", "", "mesh to install: istio, consul, linkerd2")
	pflags.StringVarP(&iop.Namespace, "namespace", "n", "", "namespace install mesh into")
	pflags.BoolVar(&iop.Mtls, "mtls", false, "use MTLS")
	pflags.StringVar(&iop.SecretRef.Name, "secret.name", "", "name of the MTLS secret")
	pflags.StringVar(&iop.SecretRef.Namespace, "secret.namespace", "", "namespace of the MTLS secret")
	return cmd
}

func install(opts *options.Options) {

	err := qualifyFlags(opts)
	if err != nil {
		fmt.Println(err)
		return
	}

	iop := &opts.Install
	switch iop.MeshType {
	case "consul":
		installConsul(opts)
		return
	case "istio":
		fmt.Println("istio TODO")
		return
	case "linkerd2":
		fmt.Println("ld TODO")
		return
	default:
		// should not get here
		fmt.Println("Please choose a valid mesh")
		return
	}

}
