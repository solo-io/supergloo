package initialize

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/rotisserie/eris"

	"github.com/gobuffalo/packr"
	"github.com/spf13/cobra"
)

func IstioCommand(ctx context.Context, mgmtCluster, remoteCluster string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap a multicluster Istio demo with Service Mesh Hub",
		Long: `
Bootstrap a multicluster Istio demo with Service Mesh Hub.

Running the Service Mesh Hub demo setup locally requires 4 tools to be installed and 
accessible via your PATH: kubectl >= v1.18.8, kind >= v0.8.1, istioctl, and docker.
We recommend allocating at least 8GB of RAM for Docker.

This command will bootstrap 2 clusters, one of which will run the Service Mesh Hub
management-plane as well as Istio, and the other will just run Istio.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initIstioCmd(ctx, mgmtCluster, remoteCluster)
		},
	}
	cmd.SilenceUsage = true
	return cmd
}

func initIstioCmd(ctx context.Context, mgmtCluster, remoteCluster string) error {
	box := packr.NewBox("./scripts")

	// management cluster
	if err := createKindCluster(mgmtCluster, managementPort, box); err != nil {
		return err
	}

	if err := installIstio(mgmtCluster, managementPort, box); err != nil {
		return err
	}

	// remote cluster
	if err := createKindCluster(remoteCluster, remotePort, box); err != nil {
		return err
	}
	if err := installIstio(remoteCluster, remotePort, box); err != nil {
		return err
	}

	// install SMH to management cluster
	if err := installServiceMeshHub(ctx, mgmtCluster, box); err != nil {
		return err
	}

	// register remote cluster
	if err := registerCluster(ctx, mgmtCluster, remoteCluster, box); err != nil {
		return err
	}

	// set context to management cluster
	return switchContext(mgmtCluster, box)
}

func installIstio(cluster string, port string, box packr.Box) error {
	fmt.Printf("Installing Istio to cluster %s\n", cluster)

	script, err := box.FindString("install_istio.sh")
	if err != nil {
		return eris.Wrap(err, "Error loading script")
	}
	cmd := exec.Command("bash", "-c", script, cluster, port)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return eris.Wrapf(err, "Error installing Istio on cluster %s", cluster)
	}

	fmt.Printf("Successfully installed Istio on cluster %s\n", cluster)
	return nil
}
