package initialize

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/gobuffalo/packr"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/demo/internal/flags"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/enterprise"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
)

const (
	// The default version of k8s under Linux is 1.18 https://github.com/solo-io/gloo-mesh/issues/700
	kindImage      = "kindest/node:v1.17.5"
	managementPort = "32001"
	remotePort     = "32000"
)

func createKindCluster(cluster string, port string, box packr.Box) error {
	fmt.Printf("Creating cluster %s with ingress port %s\n", cluster, port)

	script, err := box.FindString("create_kind_cluster.sh")
	if err != nil {
		return eris.Wrap(err, "Error loading script")
	}

	cmd := exec.Command("bash", "-c", script, cluster, port, kindImage)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return eris.Wrapf(err, "Error creating cluster %s", cluster)
	}

	fmt.Printf("Successfully created cluster %s\n", cluster)
	return nil
}

func getGlooMeshVersion(opts flags.Options) (string, error) {
	// Use user provided version first
	if opts.Version != "" {
		return opts.Version, nil
	}
	// Then if its not enterprise, return the CLI version
	if !opts.Enterprise {
		return version.Version, nil
	}
	// Lastly find the latest version of the enterprise chart
	return helm.GetLatestChartVersion(gloomesh.GlooMeshEnterpriseRepoURI, "gloo-mesh-enterprise", true)
}

func installGlooMesh(ctx context.Context, cluster string, opts flags.Options, box packr.Box) error {
	version, err := getGlooMeshVersion(opts)
	if err != nil {
		return err
	}
	if opts.Enterprise {
		return installGlooMeshEnterprise(ctx, cluster, version, opts.LicenseKey, box)
	}

	return installGlooMeshCommunity(ctx, cluster, version, box)
}

func installGlooMeshCommunity(ctx context.Context, cluster, version string, box packr.Box) error {
	fmt.Printf("Deploying Gloo Mesh to %s from images\n", cluster)
	if err := (helm.Installer{
		ChartUri:    fmt.Sprintf(gloomesh.GlooMeshChartUriTemplate, version),
		KubeContext: "kind-" + cluster,
		Namespace:   defaults.DefaultPodNamespace,
		ReleaseName: gloomesh.GlooMeshReleaseName,
		Verbose:     true,
	}).InstallChart(ctx); err != nil {
		return eris.Wrap(err, "error installing Gloo Mesh")
	}
	if err := glooMeshPostInstall(cluster, box); err != nil {
		return err
	}

	fmt.Printf("Successfully set up Gloo Mesh on cluster %s\n", cluster)
	return nil
}

func installGlooMeshEnterprise(ctx context.Context, cluster, version, licenseKey string, box packr.Box) error {
	fmt.Printf("Deploying Gloo Mesh Enterprise to %s from images\n", cluster)
	if err := (helm.Installer{
		ChartUri:    fmt.Sprintf(gloomesh.GlooMeshEnterpriseChartUriTemplate, version),
		KubeContext: "kind-" + cluster,
		Namespace:   defaults.DefaultPodNamespace,
		ReleaseName: gloomesh.GlooMeshReleaseName,
		Values:      map[string]string{"licenseKey": licenseKey},
		Verbose:     true,
	}).InstallChart(ctx); err != nil {
		return eris.Wrap(err, "error installing Gloo Mesh Enterprise")
	}
	if err := glooMeshEnterprisePostInstall(cluster, box); err != nil {
		return err
	}

	fmt.Printf("Successfully set up Gloo Mesh on cluster %s\n", cluster)
	return nil
}

func registerCluster(ctx context.Context, mgmtCluster, remoteCluster string, opts flags.Options, box packr.Box) error {
	version, err := getGlooMeshVersion(opts)
	if err != nil {
		return err
	}

	regOpts := registration.Options{
		MgmtContext:     "kind-" + mgmtCluster,
		RemoteContext:   "kind-" + remoteCluster,
		ClusterName:     remoteCluster,
		MgmtNamespace:   defaults.DefaultPodNamespace,
		RemoteNamespace: defaults.DefaultPodNamespace,
		Version:         version,
		Verbose:         true,
	}
	if opts.Enterprise {
		return registerEnterpriseCluster(ctx, regOpts, box)
	}

	return registerCommunityCluster(ctx, regOpts, box)
}

func registerCommunityCluster(ctx context.Context, regOpts registration.Options, box packr.Box) error {
	fmt.Printf("Registering cluster %s with cert-agent image\n", regOpts.ClusterName)
	apiServerAddress, err := getApiAddress(regOpts.ClusterName, box)
	if err != nil {
		return err
	}
	regOpts.ApiServerAddress = apiServerAddress
	registrant, err := registration.NewRegistrant(regOpts)
	if err != nil {
		return eris.Wrapf(err, "initializing registrant for cluster %s", regOpts.ClusterName)
	}
	if err := registrant.RegisterCluster(ctx); err != nil {
		return eris.Wrapf(err, "registering cluster %s", regOpts.ClusterName)
	}

	fmt.Printf("Successfully registered cluster %s\n", regOpts.ClusterName)
	return nil
}

func registerEnterpriseCluster(ctx context.Context, regOpts registration.Options, box packr.Box) error {
	fmt.Printf("Registering cluster %s with enterprise-agent image\n", regOpts.ClusterName)
	relayServerAddress, err := getRelayServerAddress(regOpts.ClusterName, box)
	if err != nil {
		return err
	}
	entOpts := enterprise.RegistrationOptions{
		Options:            regOpts,
		RelayServerAddress: relayServerAddress,
	}
	if err := enterprise.RegisterCluster(ctx, entOpts); err != nil {
		return err
	}
	fmt.Printf("Successfully registered cluster %s\n", regOpts.ClusterName)
	return nil
}

func getApiAddress(cluster string, box packr.Box) (string, error) {
	var buf bytes.Buffer
	if err := runScript(box, &buf, "get_api_address.sh", cluster); err != nil {
		return "", eris.Wrap(err, "Error getting API server address")
	}

	return buf.String(), nil
}

func getRelayServerAddress(cluster string, box packr.Box) (string, error) {
	var buf bytes.Buffer
	if err := runScript(box, &buf, "get_relay_server_address.sh", cluster); err != nil {
		return "", eris.Wrap(err, "Error getting relay server address")
	}

	return buf.String(), nil
}

func switchContext(cluster string, box packr.Box) error {
	if err := runScript(box, os.Stdout, "switch_context.sh", cluster); err != nil {
		return eris.Wrapf(err, "Could not switch context to %s", fmt.Sprintf("kind-%s", cluster))
	}

	return nil
}

func glooMeshPostInstall(cluster string, box packr.Box) error {
	if err := runScript(box, os.Stdout, "post_install_gloomesh.sh", cluster); err != nil {
		return eris.Wrap(err, "Error running post-install script")
	}

	return nil
}

func glooMeshEnterprisePostInstall(cluster string, box packr.Box) error {
	if err := runScript(box, os.Stdout, "post_install_gloomesh_enterprise.sh", cluster); err != nil {
		return eris.Wrap(err, "Error running post-install script")
	}

	return nil
}

func runScript(box packr.Box, out io.Writer, script string, args ...string) error {
	script, err := box.FindString(script)
	if err != nil {
		return eris.Wrap(err, "Error loading script")
	}

	cmd := exec.Command("bash", append([]string{"-c", script}, args...)...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
