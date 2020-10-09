package snapshot

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/solo-io/go-utils/cliutils"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Input and Output snapshots for the discovery and networking pod",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx)
		},
	}

	return cmd
}

func debugSnapshot(ctx context.Context) error {
	freePort, err := cliutils.GetFreePort()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	localPort := strconv.Itoa(freePort)

	for _, podName := range []string{"discovery", "networking"} {
		for _, snapshotType := range []string{"input", "output"} {
			snapshot := getSnapshot(ctx, localPort, podName, snapshotType)
			fileName := fmt.Sprintf("%s-%s-snapshot.json", podName, snapshotType)
			err := ioutil.WriteFile(fileName, []byte(snapshot), 0644)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
		}
	}
	return nil
}

func getSnapshot(ctx context.Context, localPort, podName, snapshotType string) string {
	snapshot, portFwdCmd, err := cliutils.PortForwardGet(ctx,
		"service-mesh-hub", "deploy/"+podName,
		localPort, "9091", true, "/snapshots/"+snapshotType)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	if portFwdCmd.Process != nil {
		defer portFwdCmd.Process.Release()
		defer portFwdCmd.Process.Kill()
	}
	return snapshot
}
