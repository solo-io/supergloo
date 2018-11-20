package install

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/util"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/constants"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

func getInstallClient() (v1.InstallClient, error) {
	cfg, err := kubeutils.GetConfig("", "")
	cache := kube.NewKubeCache()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return v1.NewInstallClient(&factory.KubeResourceClientFactory{
		Crd:         v1.InstallCrd,
		Cfg:         cfg,
		SharedCache: cache,
	})
}

func installationSummaryMessage(opts *options.Options) {
	fmt.Printf("Installing %v in namespace %v.\n", opts.Install.MeshType, opts.Install.Namespace)
	if opts.Install.Mtls {
		fmt.Printf("MTLS active with secret %v from namespace %v.\n", opts.Install.SecretRef.Name, opts.Install.SecretRef.Namespace)

	}
	return
}

// there is a very small chance that names may overlap
// kubernetes has good messaging for name collisions so just print the error
func getNewInstallName(opts *options.Options) string {
	return fmt.Sprintf("%v-%v", opts.Install.MeshType, util.RandStringBytes(6))
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
		if iop.Mtls {
			if iop.SecretRef.Name == "" {
				return fmt.Errorf("please specify a secret name to use MTLS")
			}
			if iop.SecretRef.Namespace == "" {
				return fmt.Errorf("please specify a secret namespace to use MTLS")
			}
		}
		return nil
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

	chosenMtls, err := chooseMtls()
	iop.Mtls = chosenMtls
	if err != nil {
		return fmt.Errorf("input error")
	}

	if iop.Mtls {
		chosenSecretNamespace, err := chooseSecretNamespace()
		if err != nil {
			return fmt.Errorf("input error")
		}
		iop.SecretRef.Namespace = chosenSecretNamespace

		chosenSecretName, err := chooseSecretName()
		if err != nil {
			return fmt.Errorf("input error")
		}
		iop.SecretRef.Name = chosenSecretName
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

func chooseMtls() (bool, error) {

	options := []string{"yes", "no"}

	question := &survey.Select{
		Message: "use MTLS?",
		Options: options,
	}

	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return false, err
	}

	if choice == "yes" {
		return true, nil
	}

	return false, nil
}

func chooseSecretNamespace() (string, error) {

	// TODO(mitchdraft) - get from system
	// AND restrict these to the NS that have secrets
	namespaceOptions := []string{"ns1", "ns2", "ns3"}

	question := &survey.Select{
		Message: "Select a secret namespace",
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

func chooseSecretName() (string, error) {

	// TODO(mitchdraft) - get from system
	nameOptions := []string{"verySecret", "sssshhhhh!!", "notSoSecret"}

	question := &survey.Select{
		Message: "Select a secret namespace",
		Options: nameOptions,
	}

	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return "", err
	}

	return choice, nil
}
