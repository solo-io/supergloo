package install

import (
	"context"
	"fmt"
	"os"
	"os/exec"

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
		Use:   "install",
		Short: "Bootstrap a multicluster Istio demo with Service Mesh Hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return install(ctx, masterCluster, remoteCluster)
		},
	}
	return cmd
}

const (
	// The default version of k8s under Linux is 1.18 https://github.com/solo-io/service-mesh-hub/issues/700
	kindImage  = "kindest/node:v1.17.5"
	masterPort = "32001"
	remotePort = "32000"
)

func install(ctx context.Context, masterCluster string, remoteCluster string) error {
	box := packr.NewBox("./scripts")
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
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
	installServiceMeshHub(ctx, masterCluster, box)

	return err
}

func createKindCluster(cluster string, port string, box packr.Box) error {
	fmt.Printf("Creating cluster %s with ingress port %s\n", cluster, port)

	script, err := box.FindString("create_kind_cluster.sh")
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "-c", script, cluster, port, kindImage)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}

func installIstio(cluster string, port string, box packr.Box) error {
	fmt.Printf("Installing Istio to cluster %s\n", cluster)

	script, err := box.FindString("install_istio.sh")
	if err != nil {
		return err
	}
	cmd := exec.Command("bash", "-c", script, cluster, port)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	fmt.Printf("Finished setting up cluster %s\n", cluster)
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
	certAgentCharUri := fmt.Sprintf(smh.CertAgentChartUriTemplate, version.Version)

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
		return err
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
			ChartPath:   certAgentCharUri,
			ChartValues: "",
		},
		Verbose: verbose,
	}
	err = registration.NewRegistrant(registrantOpts).RegisterCluster(ctx)
	if err != nil {
		return err
	}

	script, err := box.FindString("post_install_smh.sh")
	if err != nil {
		return err
	}
	cmd := exec.Command("bash", "-c", script, cluster)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	fmt.Println("Successfully set up Service Mesh Hub")
	return nil
}

//func registerCluster(cluster string) error {
//
//}

func getApiAddress(cluster string, box packr.Box) (string, error) {
	fmt.Printf("getApiAddress %s\n", cluster)

	script, err := box.FindString("get_api_address.sh")
	if err != nil {
		return "", err
	}
	cmd := exec.Command("bash", "-c", script, cluster)
	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
