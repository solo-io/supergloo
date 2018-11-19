package install

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/util"
	"github.com/solo-io/supergloo/pkg/constants"
	"github.com/spf13/cobra"
	survey "gopkg.in/AlecAivazis/survey.v1"
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
	pflags.StringVarP(&iop.MeshType, "meshtype", "m", "", "mesh to install: istio, consul, linkerd")
	pflags.StringVarP(&iop.Namespace, "namespace", "n", "", "namespace install mesh into")
	pflags.BoolVar(&iop.Mtls, "mtls", false, "use MTLS")
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
	case "linkerd":
		fmt.Println("ld TODO")
		return
	default:
		// should not get here
		fmt.Println("Please choose a valid mesh")
		return
	}

}

func qualifyFlags(opts *options.Options) error {
	top := opts.Top
	iop := &opts.Install

	// if they are using static mode, they must pass all params
	if top.Static {
		if iop.Namespace == "" {
			return fmt.Errorf("please provide a namespace")
		}
		if iop.MeshType == "" {
			return fmt.Errorf("please provide a mesh type")
		}
	}

	if iop.Namespace == "" {
		namespace, err := chooseNamespace()
		iop.Namespace = namespace
		if err != nil {
			return fmt.Errorf("input error")
		}
	}

	if iop.MeshType == "" {
		chosenMesh, err := chooseMeshType()
		iop.MeshType = chosenMesh
		if err != nil {
			return fmt.Errorf("input error")
		}
	}

	return nil
}

func chooseMeshType() (string, error) {

	question := &survey.Select{
		Message: "Select a mesh type",
		Options: constants.MeshOptions,
	}

	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return "", err
	}

	return choice, nil
}

func chooseNamespace() (string, error) {

	// TODO(mitchdraft) - get from system
	namespaceOptions := []string{"ns1", "ns2", "ns3"}

	question := &survey.Select{
		Message: "Select a namespace",
		Options: namespaceOptions,
	}

	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return "", err
	}

	return choice, nil
}

func installationSummaryMessage(opts *options.Options) {
	fmt.Printf("Installing %v in namespace %v.\n", opts.Install.MeshType, opts.Install.Namespace)
	return
}

func getNewInstallName(opts *options.Options) string {
	return fmt.Sprintf("%v-%v", opts.Install.MeshType, util.RandStringBytes(6))
}
