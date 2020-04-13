package wire

import (
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
)

// factory for clients that require kube clients
func DefaultKubeClientsFactoryProvider() common.KubeClientsFactory {
	return DefaultKubeClientsFactory
}

// factory for clients that don't require kube clients
func DefaultClientsFactoryProvider() common.ClientsFactory {
	return DefaultClientsFactory
}
