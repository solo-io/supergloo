package version

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/solo-io/mesh-projects/cli/pkg/common"

	"github.com/solo-io/mesh-projects/pkg/version"
)

const (
	undefinedServer = "version undefined, could not find any version of service mesh hub running"
)

func ReportVersion(out io.Writer, clientsFactory common.ClientsFactory, globalFlagConfig *common.GlobalFlagConfig) error {
	clientVersionInfo := map[string]string{
		"version": version.Version,
	}
	clientBytes, err := json.Marshal(clientVersionInfo)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Client: %s\n", string(clientBytes))

	serverVersionString := undefinedServer
	clients, err := clientsFactory(globalFlagConfig.MasterKubeConfig, globalFlagConfig.MasterWriteNamespace)
	if err != nil {
		return err
	}
	serverVersionInfo, err := clients.ServerVersionClient.GetServerVersion()
	if err != nil {
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
