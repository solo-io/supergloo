package mesh_install

import (
	"io"
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	install_istio "github.com/solo-io/mesh-projects/cli/pkg/tree/mesh/install/istio"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	"github.com/spf13/cobra"
)

type MeshInstallCommand *cobra.Command

var (
	MeshInstallProviderSet = wire.NewSet(
		MeshInstallRootCmd,
	)
	validMeshTypes = []string{
		strings.ToLower(types.MeshType_ISTIO.String()),
	}
	UnsupportedMeshTypeError = func(meshType string) error {
		return eris.Errorf(
			"Mesh Type: (%s) is not one of the supported Mesh types [%s]",
			meshType,
			strings.Join(validMeshTypes, "|"),
		)
	}
)

func MeshInstallRootCmd(
	clientsFactory common.ClientsFactory,
	opts *options.Options,
	out io.Writer,
	in io.Reader,
	kubeLoader common_config.KubeLoader,
	imageNameParser docker.ImageNameParser,
	fileReader common.FileReader,
) MeshInstallCommand {
	installCommand := cliconstants.MeshInstallCommand(validMeshTypes)
	cmd := &cobra.Command{
		Use:     installCommand.Use,
		Short:   installCommand.Short,
		Aliases: installCommand.Aliases,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			meshType := args[0]

			switch strings.ToUpper(meshType) {
			case types.MeshType_ISTIO.String():
				istioInstaller, err := install_istio.NewIstioInstaller(
					out,
					in,
					clientsFactory,
					opts,
					opts.Root.KubeConfig,
					opts.Root.KubeContext,
					kubeLoader,
					imageNameParser,
					fileReader,
				)
				if err != nil {
					return err
				}
				return istioInstaller.Install()
			default:
				return UnsupportedMeshTypeError(meshType)
			}
		},
	}

	options.AddMeshInstallFlags(cmd, opts)

	return cmd
}
