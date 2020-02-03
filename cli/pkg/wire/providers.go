package wire

import "github.com/solo-io/mesh-projects/cli/pkg/common"

func DefaultKubeClientsFactoryProvider() common.KubeClientsFactory {
	return DefaultKubeClientsFactory
}

func DefaultClientsFactoryProvider() common.ClientsFactory {
	return DefaultClientsFactory
}
