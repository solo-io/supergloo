package cli_mocks

import (
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"k8s.io/client-go/rest"
)

func MockClientsFactory(clients *common.Clients) common.ClientsFactory {
	return func(_ *rest.Config, _ string) (*common.Clients, error) {
		return clients, nil
	}
}
