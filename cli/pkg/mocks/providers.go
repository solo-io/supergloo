package cli_mocks

import (
	"github.com/solo-io/go-utils/installutils/helminstall"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	upgrade_assets "github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade/assets"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	"github.com/solo-io/mesh-projects/pkg/auth"
	"k8s.io/client-go/rest"
)

func MockKubeClientsFactoryProvider(authorization auth.ClusterAuthorization, writer common.SecretWriter, helmInstaller helminstall.Installer) common.KubeClientsFactory {
	return func(masterConfig *rest.Config, writeNamespace string) (clients *common.KubeClients, err error) {
		return &common.KubeClients{
			ClusterAuthorization: authorization,
			SecretWriter:         writer,
			HelmInstaller:        helmInstaller,
		}, nil
	}
}

func MockClientsFactoryProvider(
	client server.ServerVersionClient,
	kubeLoader common_config.KubeLoader,
	verifier common_config.MasterKubeConfigVerifier,
	upgrader upgrade_assets.AssetHelper) common.ClientsFactory {
	return func(opts *options.Options) (clients *common.Clients, err error) {
		return &common.Clients{
			ServerVersionClient:   client,
			KubeLoader:            kubeLoader,
			MasterClusterVerifier: verifier,
			ReleaseAssetHelper:    upgrader,
		}, nil
	}
}
