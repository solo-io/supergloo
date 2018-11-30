package install

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/common"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/nsutil"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/constants"
	"gopkg.in/AlecAivazis/survey.v1"
)

func installationSummaryMessage(opts *options.Options, name string) {
	if opts.Install.Namespace == "" {
		fmt.Printf("Installing %v with name %s.\n", opts.Install.MeshType, name)
	} else {
		fmt.Printf("Installing %v in namespace %v with name %s.\n", opts.Install.MeshType, opts.Install.Namespace, name)
	}

	if opts.Install.Mtls {
		fmt.Printf("mTLS active.\n")
	}
	return
}

// there is a very small chance that names may overlap
// kubernetes has good messaging for name collisions so just print the error
func getNewInstallName(opts *options.Options) string {
	return fmt.Sprintf("%v-%v", opts.Install.MeshType, common.RandStringBytes(6))
}

func getMetadataFromOpts(opts *options.Options) core.Metadata {
	return core.Metadata{
		Name: getNewInstallName(opts),
		// TODO(mitchdraft) get mesh storage namespace from flag
		Namespace: constants.SuperglooNamespace,
	}
}

func getEncryptionFromOpts(opts *options.Options) *v1.Encryption {
	if opts.Install.Mtls {
		return &v1.Encryption{
			TlsEnabled: opts.Install.Mtls,
		}
	}
	return &v1.Encryption{}
}

func qualifyFlags(opts *options.Options) error {
	top := opts.Top
	iop := &opts.Install

	// if they are using static mode, they must pass all params
	if top.Static {
		if iop.MeshType == "" {
			return fmt.Errorf("please provide a mesh type")
		}
		return nil
	}

	if iop.MeshType == "" {
		chosenMesh, err := chooseMeshType()
		iop.MeshType = chosenMesh
		if err != nil {
			return fmt.Errorf("input error")
		}
	}

	if iop.MeshType == common.AppMesh {
		if top.Static {
			if iop.AwsRegion == "" {
				return fmt.Errorf("please specify an aws region")
			}
			if iop.AwsSecretRef.Name == "" || iop.AwsSecretRef.Namespace == "" {
				return fmt.Errorf("please specify an aws secret")
			}
		}

		if iop.AwsRegion == "" {
			awsRegion, err := common.GetString("Select an aws region")
			if err != nil {
				return fmt.Errorf("input error")
			}
			iop.AwsRegion = awsRegion
		}

		if err := nsutil.EnsureCommonResource("awssecret", "awssecret", &iop.AwsSecretRef, opts); err != nil {
			return err
		}

		// short-circuit, none of the remaining options are used for AppMesh
		return nil
	}

	if top.Static {
		if iop.Namespace == "" {
			return fmt.Errorf("please provide a namespace")
		}
		// user does not need to pass a custom secret
		// if they do, they must pass both the name and namespace
		if iop.SecretRef.Namespace != "" && iop.SecretRef.Name == "" {
			return fmt.Errorf("please specify a secret name")
		}
		if iop.SecretRef.Name != "" && iop.SecretRef.Namespace == "" {
			return fmt.Errorf("please specify a secret namespace")
		}
	}

	if iop.Namespace == "" {
		namespace, err := common.ChooseNamespace(opts, "Select an installation namespace")
		if err != nil {
			return fmt.Errorf("input error")
		}
		iop.Namespace = namespace
	}

	if common.Contains([]string{common.Istio, common.Linkerd2}, iop.MeshType) {
		watchNamespaces, err := chooseWatchNamespaces(opts)
		if err != nil {
			return fmt.Errorf("input error")
		}
		iop.WatchNamespaces = watchNamespaces
	}

	chosenMtls, err := common.ChooseBool("use mTLS?")
	iop.Mtls = chosenMtls
	if err != nil {
		return fmt.Errorf("input error")
	}

	if iop.Mtls {
		useCustomSecret, err := common.ChooseBool("use custom secret?")
		if err != nil {
			return fmt.Errorf("input error")
		}
		if useCustomSecret {
			if err := nsutil.EnsureCommonResource("secret", "secret", &iop.SecretRef, opts); err != nil {
				return err
			}
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
