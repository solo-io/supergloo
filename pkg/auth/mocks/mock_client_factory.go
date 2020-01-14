package mock_auth

import (
	"github.com/solo-io/mesh-projects/pkg/auth"
	"k8s.io/client-go/rest"
)

// create a client factory that just returns mock implementations of the clients
func MockClients(saClient auth.ServiceAccountClient, rbacClient auth.RbacClient, secretClient auth.SecretClient) auth.ClientFactory {
	return func(cfg *rest.Config, writeNamespace string) (*auth.Clients, error) {
		return &auth.Clients{ServiceAccountClient: saClient, RbacClient: rbacClient, SecretClient: secretClient}, nil
	}
}
