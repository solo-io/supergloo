package initialize

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/codegen/helm"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/common/registration"
	"github.com/solo-io/service-mesh-hub/pkg/common/version"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/smh"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"github.com/solo-io/skv2/pkg/multicluster/register"
)

const (
	// The default version of k8s under Linux is 1.18 https://github.com/solo-io/service-mesh-hub/issues/700
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

func installServiceMeshHub(ctx context.Context, cluster string, box packr.Box) error {
	fmt.Printf("Deploying Service Mesh Hub to %s from images\n", cluster)

	apiServerAddress, err := getApiAddress(cluster, box)
	if err != nil {
		return err
	}

	kubeContext := fmt.Sprintf("kind-%s", cluster)
	kubeConfig, err := kubeconfig.NewKubeLoader(5*time.Second).GetClientConfigForContext("", kubeContext)
	if err != nil {
		return err
	}

	namespace := defaults.DefaultPodNamespace
	verbose := true
	smhChartUri := fmt.Sprintf(smh.ServiceMeshHubChartUriTemplate, version.Version)
	certAgentChartUri := fmt.Sprintf(smh.CertAgentChartUriTemplate, version.Version)

	err = smh.Installer{
		HelmChartPath:  smhChartUri,
		HelmValuesPath: "",
		KubeConfig:     kubeConfig,
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
			ClusterName:      cluster,
			KubeCfg:          kubeConfig,
			RemoteKubeCfg:    kubeConfig,
			RemoteCtx:        kubeContext,
			Namespace:        namespace,
			RemoteNamespace:  namespace,
			APIServerAddress: apiServerAddress,
			ClusterDomain:    "",
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

func registerCluster(ctx context.Context, mgmtCluster string, cluster string, box packr.Box) error {
	fmt.Printf("Registering cluster %s with cert-agent image\n", cluster)

	apiServerAddress, err := getApiAddress(cluster, box)
	if err != nil {
		return err
	}

	kubeLoader := kubeconfig.NewKubeLoader(5 * time.Second)
	kubeConfig, err := kubeLoader.GetClientConfigForContext("", fmt.Sprintf("kind-%s", mgmtCluster))
	if err != nil {
		return err
	}

	remoteKubeContext := fmt.Sprintf("kind-%s", cluster)
	remoteKubeConfig, err := kubeLoader.GetClientConfigForContext("", remoteKubeContext)
	if err != nil {
		return err
	}

	namespace := defaults.DefaultPodNamespace
	certAgentChartUri := fmt.Sprintf(smh.CertAgentChartUriTemplate, version.Version)

	registrantOpts := &registration.RegistrantOptions{
		RegistrationOptions: register.RegistrationOptions{
			ClusterName:      cluster,
			KubeCfg:          kubeConfig,
			RemoteKubeCfg:    remoteKubeConfig,
			RemoteCtx:        remoteKubeContext,
			Namespace:        namespace,
			RemoteNamespace:  namespace,
			APIServerAddress: apiServerAddress,
			ClusterDomain:    "",
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
