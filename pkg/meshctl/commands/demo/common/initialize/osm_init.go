package initialize

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/gobuffalo/packr"
	"github.com/rotisserie/eris"
	"github.com/spf13/cobra"
)

func OsmCommand(ctx context.Context, mgmtCluster string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap an OSM demo with Service Mesh Hub",
		Long: `
Bootstrap an  OSM demo with Service Mesh Hub.

Running the Service Mesh Hub demo setup locally requires 4 tools to be installed and 
accessible via your PATH: kubectl >= v1.18.8, kind >= v0.8.1, OSM >= v0.3.0, and docker.
We recommend allocating at least 4GB of RAM for Docker.

This command will initialize a local kubernetes cluster using KinD. It will then install
all default OSM resources, which include the control-plane, prometheus, grafana, and zipkin. 
It will also install Service Mesh Hub, which includes the discovery, networking, and cert-agent
deployments.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initOSMCmd(ctx, mgmtCluster)
		},
	}
	cmd.SilenceUsage = true
	return cmd
}

func initOSMCmd(ctx context.Context, mgmtCluster string) error {
	box := packr.NewBox("./scripts")

	// management cluster
	if err := createKindCluster(mgmtCluster, managementPort, box); err != nil {
		return err
	}

	if err := installOSM(mgmtCluster, box); err != nil {
		return err
	}

	// install SMH to management cluster
	if err := installServiceMeshHub(ctx, mgmtCluster, box); err != nil {
		return err
	}
	// set context to management cluster
	return switchContext(mgmtCluster, box)
}

func installOSM(cluster string, box packr.Box) error {
	fmt.Printf("Installing OSM to cluster %s\n", cluster)

	script, err := box.FindString("install_osm.sh")
	if err != nil {
		return eris.Wrap(err, "Error loading script")
	}
	cmd := exec.Command("bash", "-c", script, cluster)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return eris.Wrapf(err, "Error installing Istio on cluster %s", cluster)
	}

	fmt.Printf("Successfully installed OSM on cluster %s\n", cluster)
	return nil
}
