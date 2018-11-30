package install

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/common"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/pkg/api/v1"
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
	pflags.StringVarP(&iop.MeshType, "meshtype", "m", "", "mesh to install: istio, consul, linkerd2, appmesh")
	pflags.StringVarP(&iop.Namespace, "namespace", "n", "", "namespace to install mesh into")
	pflags.BoolVar(&iop.Mtls, "mtls", false, "use mTLS")
	pflags.StringVar(&iop.SecretRef.Name, "secret.name", "", "name of the mTLS secret")
	pflags.StringVar(&iop.SecretRef.Namespace, "secret.namespace", "", "namespace of the mTLS secret")
	pflags.StringVar(&iop.AwsSecretRef.Name, "awssecret.name", "", "name of the AWS secret")
	pflags.StringVar(&iop.AwsSecretRef.Namespace, "awssecret.namespace", "", "namespace of the AWS secret")
	pflags.StringVar(&iop.AwsRegion, "aws-region", "", "AWS region")
	return cmd
}

func install(opts *options.Options) {

	err := qualifyFlags(opts)
	if err != nil {
		fmt.Println(err)
		return
	}

	installClient, err := common.GetInstallClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	var installSpec *v1.Install
	switch opts.Install.MeshType {
	case "consul":
		installSpec = generateConsulInstallSpecFromOpts(opts)
	case "istio":
		installSpec = generateIstioInstallSpecFromOpts(opts)
	case "linkerd2":
		installSpec = generateLinkerd2InstallSpecFromOpts(opts)
	}

	// App mesh is a special case that is installed in the translator syncer, until we refactor install syncer to allow non-helm installs
	if opts.Install.MeshType == "appmesh" {
		installAppMesh(opts)
	} else {
		_, err = (*installClient).Write(installSpec, clients.WriteOpts{})
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	installationSummaryMessage(opts)
	return
}

func installAppMesh(opts *options.Options) error {
	meshClient, err := common.GetMeshClient()
	if err != nil {
		fmt.Println(err)
		return err
	}
	mesh := generateAppMeshInstallSpecFromOpts(opts)
	_, err = (*meshClient).Write(mesh, clients.WriteOpts{})
	return err
}
