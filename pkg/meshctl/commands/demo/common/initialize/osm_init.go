package initialize

import (
	"context"
	"fmt"
	"os"

	"github.com/gobuffalo/packr"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/demo/internal/flags"
	"github.com/spf13/cobra"
)

func OsmCommand(ctx context.Context, mgmtCluster string) *cobra.Command {
	opts := flags.Options{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap an OSM demo with Gloo Mesh",
		Long: `
Bootstrap an  OSM demo with Gloo Mesh.

Running the Gloo Mesh demo setup locally requires 4 tools to be installed and 
accessible via your PATH: kubectl >= v1.18.8, kind >= v0.8.1, OSM >= v0.3.0, and docker.
We recommend allocating at least 4GB of RAM for Docker.

This command will initialize a local kubernetes cluster using KinD. It will then install
all default OSM resources, which include the control-plane, prometheus, grafana, and zipkin. 
It will also install Gloo Mesh, which includes the discovery, networking, and cert-agent
deployments.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initOSMCmd(ctx, mgmtCluster, opts)
		},
	}
	opts.AddToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	return cmd
}

func initOSMCmd(ctx context.Context, mgmtCluster string, opts flags.Options) error {
	box := packr.NewBox("./scripts")

	// management cluster
	if err := createKindCluster(mgmtCluster, managementPort, box); err != nil {
		return err
	}

	if err := installOSM(mgmtCluster, box); err != nil {
		return err
	}

	// install GlooMesh to management cluster
	if err := installGlooMesh(ctx, mgmtCluster, opts, box); err != nil {
		return err
	}

	// register management cluster
	if err := registerCluster(ctx, mgmtCluster, mgmtCluster, opts, box); err != nil {
		return err
	}

	// set context to management cluster
	return switchContext(mgmtCluster, box)
}

func installOSM(cluster string, box packr.Box) error {
	fmt.Printf("Installing OSM to cluster %s\n", cluster)
	if err := runScript(box, os.Stdout, "install_osm.sh", cluster); err != nil {
		return eris.Wrapf(err, "Error installing Istio on cluster %s", cluster)
	}

	fmt.Printf("Successfully installed OSM on cluster %s\n", cluster)
	return nil
}
