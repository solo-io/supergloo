package wire

import (
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/files"
)

// factory for clients that require kube clients
func DefaultKubeClientsFactoryProvider(fileReader files.FileReader) common.KubeClientsFactory {
	return DefaultKubeClientsFactory(fileReader)
}

// factory for clients that don't require kube clients
func DefaultClientsFactoryProvider() common.ClientsFactory {
	return DefaultClientsFactory
}
