package istio1_5

import (
	"context"
	"io"

	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/mesh/install/istio"
	"github.com/solo-io/service-mesh-hub/pkg/container-runtime/docker"
	"github.com/solo-io/service-mesh-hub/pkg/filesystem/files"
	"github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-installation/istio/operator"
	"github.com/spf13/cobra"
)

type Istio1_5Cmd *cobra.Command

func NewIstio1_5InstallCmd(
	ctx context.Context,
	out io.Writer,
	kubeLoader kubeconfig.KubeLoader,
	opts *options.Options,
	kubeClientsFactory common.KubeClientsFactory,
	clientsFactory common.ClientsFactory,
	imageNameParser docker.ImageNameParser,
	fileReader files.FileReader,
) Istio1_5Cmd {
	cmd := &cobra.Command{
		Use:   cliconstants.Istio1_5Command.Use,
		Short: cliconstants.Istio1_5Command.Short,
		RunE: istio.BuildIstioInstallationRunFunc(
			ctx,
			out,
			operator.IstioVersion1_5,
			kubeLoader,
			opts,
			kubeClientsFactory,
			clientsFactory,
			imageNameParser,
			fileReader,
		),
	}

	options.AddIstioInstallFlags(cmd, opts)

	return cmd
}
