package install

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: `install a mesh`,
		Long:  `install a mesh.`,
		Run: func(c *cobra.Command, args []string) {
			install(opts)
		},
	}
	iop := &opts.Install
	pflags := cmd.PersistentFlags()
	pflags.StringVarP(&iop.Filename, "filename", "f", "", "filename to create resources from")
	pflags.StringVarP(&iop.MeshType, "meshtype", "m", "", "mesh to install: istio, consul, linkerd")
	pflags.StringVarP(&iop.Namespace, "namespace", "n", "", "namespace to use")
	return cmd
}

func install(opts *options.Options) {

	err := qualifyFlags(opts)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("installing %v in namespace %v from %v\n", opts.Install.MeshType, opts.Install.Namespace, opts.Install.Filename)
}

func qualifyFlags(opts *options.Options) error {
	top := opts.Top
	iop := &opts.Install

	// we always need a filename
	if iop.Filename == "" {
		return fmt.Errorf("please provide a filename")
	}

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

	// TODO(mitchdraft) - get from system/constants
	meshOptions := []string{"istio", "consul", "linkerd"}

	question := &survey.Select{
		Message: "Select a mesh type",
		Options: meshOptions,
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
