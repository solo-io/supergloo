package common

import (
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	upgrade_assets "github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade/assets"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	"github.com/solo-io/mesh-projects/pkg/auth"
	k8sapiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// a grab bag of various clients that command implementations may use
type KubeClients struct {
	ClusterAuthorization auth.ClusterAuthorization
	SecretWriter         SecretWriter
}

type KubeClientsFactory func(masterConfig *rest.Config, writeNamespace string) (*KubeClients, error)

type Clients struct {
	ServerVersionClient   server.ServerVersionClient
	ReleaseAssetHelper    upgrade_assets.AssetHelper
	KubeLoader            common_config.KubeLoader
	MasterClusterVerifier common_config.MasterKubeConfigVerifier
}

type ClientsFactory func(opts *options.Options) (*Clients, error)

// write a k8s secret
// used to pare down from the complexity of all the methods on the k8s client-go SecretInterface
//go:generate mockgen -destination ../mocks/mock_secret_writer.go -package cli_mocks github.com/solo-io/mesh-projects/cli/pkg/common SecretWriter
type SecretWriter interface {
	Write(secret *k8sapiv1.Secret) error
}

func ClientsProvider(
	serverVersionClient server.ServerVersionClient,
	upgrader upgrade_assets.AssetHelper,
	loader common_config.KubeLoader,
	verifier common_config.MasterKubeConfigVerifier) *Clients {
	return &Clients{
		ServerVersionClient:   serverVersionClient,
		ReleaseAssetHelper:    upgrader,
		KubeLoader:            loader,
		MasterClusterVerifier: verifier,
	}
}

// facilitates wire codegen
func KubeClientsProvider(authorization auth.ClusterAuthorization, writer SecretWriter) *KubeClients {
	return &KubeClients{
		ClusterAuthorization: authorization,
		SecretWriter:         writer,
	}
}
