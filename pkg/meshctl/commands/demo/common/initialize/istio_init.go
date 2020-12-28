package initialize

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/rotisserie/eris"

	"github.com/gobuffalo/packr"
	"github.com/spf13/cobra"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/demo/internal/flags"
)

func IstioCommand(ctx context.Context, mgmtCluster, remoteCluster string) *cobra.Command {
	opts := &flags.Options{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap a multicluster Istio demo with Gloo Mesh",
		Long: `
Bootstrap a multicluster Istio demo with Gloo Mesh.

Running the Gloo Mesh demo setup locally requires 4 tools to be installed and 
accessible via your PATH: kubectl >= v1.18.8, kind >= v0.8.1, istioctl, and docker.
We recommend allocating at least 8GB of RAM for Docker.

This command will bootstrap 2 clusters, one of which will run the Gloo Mesh
management-plane as well as Istio, and the other will just run Istio.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initIstioCmd(ctx, mgmtCluster, remoteCluster, opts)
		},
	}
	opts.AddToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	return cmd
}

func initIstioCmd(ctx context.Context, mgmtCluster, remoteCluster string, opts *flags.Options) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	box := packr.NewBox("./scripts")

	// make sure istio version is supported
	if err := checkIstioVersion(box); err != nil {
		return err
	}

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

	// install GlooMesh to management cluster
	if err := installGlooMesh(ctx, mgmtCluster, box); err != nil {
		return err
	}

	// install GlooMesh Enterprise to management cluster, if enabled
	if opts.Enterprise {
		if err := installGlooMeshEnterprise(ctx, mgmtCluster, opts.EnterpriseVersion, opts.LicenseKey, box); err != nil {
			return err
		}
	}

	// register management cluster
	if err := registerCluster(ctx, mgmtCluster, mgmtCluster, opts.Enterprise, box); err != nil {
		return err
	}

	// register remote cluster
	if err := registerCluster(ctx, mgmtCluster, remoteCluster, opts.Enterprise, box); err != nil {
		return err
	}

	// set context to management cluster
	return switchContext(mgmtCluster, box)
}

func installIstio(cluster string, port string, box packr.Box) error {
	fmt.Printf("Installing Istio to cluster %s\n", cluster)
	if err := runScript(box, os.Stdout, "install_istio.sh", cluster, port); err != nil {
		return eris.Wrapf(err, "Error installing Istio on cluster %s", cluster)
	}

	fmt.Printf("Successfully installed Istio on cluster %s\n", cluster)
	return nil
}

func checkIstioVersion(box packr.Box) error {
	var buf bytes.Buffer
	if err := runScript(box, &buf, "check_istio_version.sh"); err != nil {
		return eris.Wrap(err, buf.String())
	}

	return nil
}
