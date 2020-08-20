package init

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/rotisserie/eris"

	"github.com/gobuffalo/packr"
	"github.com/solo-io/service-mesh-hub/codegen/helm"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/common/version"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/smh"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/registration"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context, masterCluster string, remoteCluster string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap a multicluster Istio demo with Service Mesh Hub",
		Long: `
Running the Service Mesh Hub demo setup locally requires 4 tools to be installed and 
accessible via your PATH: kubectl, kind, docker, and istioctl < v1.7.0.

This command will bootstrap 2 clusters, one of which will run the Service Mesh Hub
management-plane as well as Istio, and the other will just run Istio.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initCmd(ctx, masterCluster, remoteCluster)
		},
	}
	cmd.SilenceUsage = true
	return cmd
}

const (
	// The default version of k8s under Linux is 1.18 https://github.com/solo-io/service-mesh-hub/issues/700
	kindImage  = "kindest/node:v1.17.5"
	masterPort = "32001"
	remotePort = "32000"
)

func initCmd(ctx context.Context, masterCluster string, remoteCluster string) error {
	box := packr.NewBox("./scripts")
	projectRoot, err := os.Getwd()
	if err != nil {
		return eris.Wrap(err, "Unable to get working directory")
	}
	fmt.Printf("Using project root %s\n", projectRoot)

	// master cluster
	err = createKindCluster(masterCluster, masterPort, box)
	if err != nil {
		return err
	}
	err = installIstio(masterCluster, masterPort, box)
	if err != nil {
		return err
	}

	// remote cluster
	err = createKindCluster(remoteCluster, remotePort, box)
	if err != nil {
		return err
	}
	err = installIstio(remoteCluster, remotePort, box)
	if err != nil {
		return err
	}

	// install SMH to master cluster
	err = installServiceMeshHub(ctx, masterCluster, box)
	if err != nil {
		return err
	}

	// register remote cluster
	err = registerCluster(ctx, masterCluster, remoteCluster, box)
	if err != nil {
		return err
	}

	// set context to master cluster
	err = switchContext(masterCluster, box)

	return err
}

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

func installServiceMeshHub(ctx context.Context, cluster string, box packr.Box) error {
	fmt.Printf("Deploying Service Mesh Hub to %s from images\n", cluster)

	apiServerAddress, err := getApiAddress(cluster, box)
	if err != nil {
		return err
	}

	kubeContext := fmt.Sprintf("kind-%s", cluster)
	kubeConfigPath := ""
	namespace := defaults.DefaultPodNamespace
	verbose := true
	smhChartUri := fmt.Sprintf(smh.ServiceMeshHubChartUriTemplate, version.Version)
	certAgentChartUri := fmt.Sprintf(smh.CertAgentChartUriTemplate, version.Version)

	err = smh.Installer{
		HelmChartPath:  smhChartUri,
		HelmValuesPath: "",
		KubeConfig:     kubeConfigPath,
		KubeContext:    kubeContext,
		Namespace:      namespace,
		ReleaseName:    helm.Chart.Data.Name,
		Verbose:        verbose,
		DryRun:         false,
	}.InstallServiceMeshHub(
		ctx,
	)
	if err != nil {
		return eris.Wrap(err, "Error installing Service Mesh Hub")
	}

	registrantOpts := &registration.RegistrantOptions{
		RegistrationOptions: register.RegistrationOptions{
			ClusterName:       cluster,
			KubeCfgPath:       kubeConfigPath,
			KubeContext:       kubeContext,
			RemoteKubeContext: kubeContext,
			Namespace:         namespace,
			RemoteNamespace:   namespace,
			APIServerAddress:  apiServerAddress,
			ClusterDomain:     "",
		},
		CertAgentInstallOptions: registration.CertAgentInstallOptions{
			ChartPath:   certAgentChartUri,
			ChartValues: "",
		},
		Verbose: verbose,
	}

	err = registration.NewRegistrant(registrantOpts).RegisterCluster(ctx)
	if err != nil {
		return eris.Wrapf(err, "Error registering cluster %s", cluster)
	}

	script, err := box.FindString("post_install_smh.sh")
	if err != nil {
		return eris.Wrap(err, "Error loading script")
	}

	cmd := exec.Command("bash", "-c", script, cluster)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return eris.Wrap(err, "Error running post-install script")
	}

	fmt.Printf("Successfully set up Service Mesh Hub on cluster %s\n", cluster)
	return nil
}

func registerCluster(ctx context.Context, masterCluster string, cluster string, box packr.Box) error {
	fmt.Printf("Registering cluster %s with cert-agent image\n", cluster)

	apiServerAddress, err := getApiAddress(cluster, box)
	if err != nil {
		return err
	}

	kubeContext := fmt.Sprintf("kind-%s", masterCluster)
	remoteKubeContext := fmt.Sprintf("kind-%s", cluster)
	namespace := defaults.DefaultPodNamespace
	certAgentChartUri := fmt.Sprintf(smh.CertAgentChartUriTemplate, version.Version)

	registrantOpts := &registration.RegistrantOptions{
		RegistrationOptions: register.RegistrationOptions{
			ClusterName:       cluster,
			KubeCfgPath:       "",
			KubeContext:       kubeContext,
			RemoteKubeContext: remoteKubeContext,
			Namespace:         namespace,
			RemoteNamespace:   namespace,
			APIServerAddress:  apiServerAddress,
			ClusterDomain:     "",
		},
		CertAgentInstallOptions: registration.CertAgentInstallOptions{
			ChartPath:   certAgentChartUri,
			ChartValues: "",
		},
		Verbose: true,
	}

	err = registration.NewRegistrant(registrantOpts).RegisterCluster(ctx)
	if err != nil {
		return eris.Wrapf(err, "Error registering cluster %s", cluster)
	}

	fmt.Printf("Successfully registered cluster %s\n", cluster)
	return nil
}

func getApiAddress(cluster string, box packr.Box) (string, error) {
	script, err := box.FindString("get_api_address.sh")
	if err != nil {
		return "", eris.Wrap(err, "Error loading script")
	}
	cmd := exec.Command("bash", "-c", script, cluster)
	bytes, err := cmd.Output()
	if err != nil {
		return "", eris.Wrap(err, "Error getting API server address")
	}

	return string(bytes), nil
}

func switchContext(cluster string, box packr.Box) error {
	script, err := box.FindString("switch_context.sh")
	if err != nil {
		return err
	}
	cmd := exec.Command("bash", "-c", script, cluster)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return eris.Wrapf(err, "Could not switch context to %s", fmt.Sprintf("kind-%s", cluster))
	}
	return nil
}
