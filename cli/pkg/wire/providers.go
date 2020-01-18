package wire

import "github.com/solo-io/mesh-projects/cli/pkg/common"

func DefaultClientsFactoryProvider() common.ClientsFactory {
	return DefaultClientsFactory
}
