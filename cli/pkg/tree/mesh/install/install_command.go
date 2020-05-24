package mesh_install

import (
	"io"
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/files"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	install_istio "github.com/solo-io/service-mesh-hub/cli/pkg/tree/mesh/install/istio"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/mesh/install/istio/operator"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/container-runtime/docker"
	"github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig"
	"github.com/spf13/cobra"
)

type MeshInstallCommand *cobra.Command

var (
	MeshInstallProviderSet = wire.NewSet(
		MeshInstallRootCmd,
	)
	validMeshTypes = map[string]zephyr_core_types.MeshType{
		"istio1.5": zephyr_core_types.MeshType_ISTIO1_5,
		"istio1.6": zephyr_core_types.MeshType_ISTIO1_6,
	}
	UnsupportedMeshTypeError = func(meshType string, validMeshTypeArgs []string) error {
		return eris.Errorf(
			"Mesh Type: (%s) is not one of the supported Mesh types [%s]",
			meshType,
			strings.Join(validMeshTypeArgs, "|"),
		)
	}
	NotImplementedMeshTypeError = func(meshType string) error {
		return eris.Errorf("Mesh type %s has not had its installation implemented. This is unexpected.", meshType)
	}
)

func MeshInstallRootCmd(
	clientsFactory common.ClientsFactory,
	opts *options.Options,
	out io.Writer,
	in io.Reader,
	kubeLoader kubeconfig.KubeLoader,
	imageNameParser docker.ImageNameParser,
	fileReader files.FileReader,
) MeshInstallCommand {
	var validMeshTypeArgs []string
	for validArg, _ := range validMeshTypes {
		validMeshTypeArgs = append(validMeshTypeArgs, validArg)
	}

	installCommand := cliconstants.MeshInstallCommand(validMeshTypeArgs)
	cmd := &cobra.Command{
		Use:     installCommand.Use,
		Short:   installCommand.Short,
		Aliases: installCommand.Aliases,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			meshType, ok := validMeshTypes[args[0]]

			if !ok {
				return UnsupportedMeshTypeError(args[0], validMeshTypeArgs)
			}

			switch meshType {
			case zephyr_core_types.MeshType_ISTIO1_5:
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
				return istioInstaller.Install(operator.Istio1_5)
			default:
				return NotImplementedMeshTypeError(args[0])
			}
		},
	}

	options.AddMeshInstallFlags(cmd, opts)

	return cmd
}
