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
	"github.com/solo-io/gloo-mesh/codegen/helm"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	installhelm "github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/skv2/pkg/multicluster/register"
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

func installGlooMesh(ctx context.Context, cluster string, box packr.Box) error {
	fmt.Printf("Deploying Gloo Mesh to %s from images\n", cluster)
	if err := getGlooMeshInstaller(
		cluster, gloomesh.GlooMeshChartUriTemplate,
		version.Version,
		nil,
	).InstallChart(ctx); err != nil {
		return eris.Wrap(err, "Error installing Gloo Mesh")
	}

	if err := glooMeshPostInstall(cluster, box); err != nil {
		return err
	}

	fmt.Printf("Successfully set up Gloo Mesh on cluster %s\n", cluster)
	return nil
}

func installGlooMeshEnterprise(ctx context.Context, cluster, version, licenseKey string, box packr.Box) error {
	fmt.Printf("Deploying Gloo Mesh Enterprise to %s from images\n", cluster)
	if version == "" {
		v, err := installhelm.GetLatestChartVersion(gloomesh.GlooMeshEnterpriseRepoURI, "gloo-mesh-enterprise")
		if err != nil {
			return err
		}
		version = v
	}

	if err := getGlooMeshInstaller(
		cluster, gloomesh.GlooMeshEnterpriseChartUriTemplate,
		version,
		map[string]string{"licenseKey": licenseKey},
	).InstallChart(ctx); err != nil {
		return eris.Wrap(err, "Error installing Gloo Mesh")
	}

	fmt.Printf("Successfully set up Gloo Mesh on cluster %s\n", cluster)
	return nil
}

func getGlooMeshInstaller(cluster, chartTemplate, chartVersion string, values map[string]string) installhelm.Installer {
	return installhelm.Installer{
		ChartUri:    fmt.Sprintf(chartTemplate, chartVersion),
		KubeConfig:  "",
		KubeContext: fmt.Sprintf("kind-%s", cluster),
		Namespace:   defaults.DefaultPodNamespace,
		ReleaseName: helm.Chart.Data.Name,
		Values:      values,
		Verbose:     true,
		DryRun:      false,
	}
}

func registerCluster(ctx context.Context, mgmtCluster, cluster string, installEnterpriseAgent bool, box packr.Box) error {
	fmt.Printf("Registering cluster %s with cert-agent image\n", cluster)
	apiServerAddress, err := getApiAddress(cluster, box)
	if err != nil {
		return err
	}

	mgmtKubeContext := fmt.Sprintf("kind-%s", mgmtCluster)
	remoteKubeContext := fmt.Sprintf("kind-%s", cluster)

	registrantOpts := registration.RegistrantOptions{
		KubeConfigPath: "",
		MgmtContext:    mgmtKubeContext,
		RemoteContext:  remoteKubeContext,
		Registration: register.RegistrationOptions{
			ClusterName:      cluster,
			RemoteCtx:        remoteKubeContext,
			Namespace:        defaults.DefaultPodNamespace,
			RemoteNamespace:  defaults.DefaultPodNamespace,
			APIServerAddress: apiServerAddress,
			ClusterDomain:    "",
		},
		AgentChartPathOverride: fmt.Sprintf(gloomesh.CertAgentChartUriTemplate, version.Version),
		AgentChartValues:       "",
		Verbose:                true,
	}

	registrant, err := registration.NewRegistrant(
		registrantOpts,
		gloomesh.CertAgentReleaseName,
		gloomesh.CertAgentChartUriTemplate,
	)
	if err != nil {
		return eris.Wrapf(err, "initializing registrant for cluster %s", cluster)
	}
	if err := registrant.RegisterCluster(ctx); err != nil {
		return eris.Wrapf(err, "registering cluster %s", cluster)
	}

	fmt.Printf("Successfully registered cluster %s\n", cluster)
	return nil
}

func getApiAddress(cluster string, box packr.Box) (string, error) {
	var buf bytes.Buffer
	if err := runScript(box, &buf, "get_api_address.sh", cluster); err != nil {
		return "", eris.Wrap(err, "Error getting API server address")
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
