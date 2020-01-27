package common

import (
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	"github.com/solo-io/mesh-projects/pkg/auth"
	k8sapiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

// a grab bag of various clients that command implementations may use
type Clients struct {
	ClusterAuthorization auth.ClusterAuthorization
	SecretWriter         SecretWriter
	KubeLoader           KubeLoader
	ServerVersionClient  server.ServerVersionClient
}

type ClientsFactory func(masterConfig *rest.Config, writeNamespace string) (*Clients, error)

// only the pieces from a kube config that we need to operate on
// mainly just used to simplify from the complexity of the actual object
type KubeContext struct {
	CurrentContext string
	Contexts       map[string]*api.Context
	Clusters       map[string]*api.Cluster
}

// given a path to a kube config file, convert it into either creds for hitting the API server of the cluster it points to,
// or return the contexts/clusters it is aware of
//go:generate mockgen -destination ../mocks/mock_kube_loader.go -package cli_mocks github.com/solo-io/mesh-projects/cli/pkg/common KubeLoader
type KubeLoader interface {
	GetRestConfig(path string) (*rest.Config, error)
	ParseContext(path string) (*KubeContext, error)
}

// write a k8s secret
// used to pare down from the complexity of all the methods on the k8s client-go SecretInterface
//go:generate mockgen -destination ../mocks/mock_secret_writer.go -package cli_mocks github.com/solo-io/mesh-projects/cli/pkg/common SecretWriter
type SecretWriter interface {
	Write(secret *k8sapiv1.Secret) error
}

// facilitates wire codegen
func ClientsProvider(authorization auth.ClusterAuthorization, writer SecretWriter, kubeLoader KubeLoader, serverVersionClient server.ServerVersionClient) *Clients {
	return &Clients{
		ClusterAuthorization: authorization,
		SecretWriter:         writer,
		KubeLoader:           kubeLoader,
		ServerVersionClient:  serverVersionClient,
	}
}
