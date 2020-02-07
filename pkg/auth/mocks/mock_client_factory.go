package mock_auth

import (
	"github.com/solo-io/mesh-projects/pkg/auth"
)

// create a client factory that just returns mock implementations of the clients
func MockClients(saClient auth.ServiceAccountClient, rbacClient auth.RbacClient, secretClient auth.SecretClient) *auth.Clients {
	return &auth.Clients{ServiceAccountClient: saClient, RbacClient: rbacClient, SecretClient: secretClient}
}
