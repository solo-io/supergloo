//+build wireinject

package wire

import (
	"io"

	"github.com/google/wire"
	cli "github.com/solo-io/mesh-projects/cli/pkg"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/solo-io/mesh-projects/pkg/auth"
	"github.com/spf13/cobra"
	k8sclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func DefaultClientsFactory(masterConfig *rest.Config, writeNamespace string) (*common.Clients, error) {
	wire.Build(
		auth.DefaultClientsProvider,
		auth.NewRemoteAuthorityConfigCreator,
		auth.NewRemoteAuthorityManager,
		k8sclientv1.NewForConfig,
		common.DefaultKubeLoaderProvider,
		common.DefaultSecretWriterProvider,
		auth.NewClusterAuthorization,
		common.ClientsProvider,
	)

	return nil, nil
}

func InitializeCLI(out io.Writer) *cobra.Command {
	wire.Build(
		DefaultClientsFactoryProvider,
		common.DefaultFileExistenceCheckerProvider,
		common.DefaultKubeLoaderProvider,
		common.NewMasterKubeConfigVerifier,
		cli.BuildCli,
	)

	return nil
}
