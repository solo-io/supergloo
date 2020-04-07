package version

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/version/server"
	"github.com/solo-io/service-mesh-hub/pkg/version"
)

const (
	undefinedServer = "version undefined, could not find any version of service mesh hub running"
)

func ReportVersion(out io.Writer, clientsFactory common.ClientsFactory, opts *options.Options) error {
	clientVersionInfo := map[string]string{
		"version": version.Version,
	}
	clientBytes, err := json.Marshal(clientVersionInfo)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Client: %s\n", string(clientBytes))

	serverVersionString := undefinedServer
	clients, err := clientsFactory(opts)
	if err != nil {
		return err
	}
	serverVersionInfo, err := clients.ServerVersionClient.GetServerVersion()
	if err != nil && !eris.Is(err, server.ConfigClientError(err)) {
		return err
	}
	if serverVersionInfo != nil {
		serverVersionBytes, err := json.Marshal(serverVersionInfo)
		if err != nil {
			return err
		}
		serverVersionString = string(serverVersionBytes)
	}
	fmt.Fprintf(out, "Server: %s\n", serverVersionString)
	return nil
}
