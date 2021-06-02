package configure

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/manifoldco/promptui"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context) *cobra.Command {
	var meshctlConfigPath string
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure Kubernetes Clusters registered with Gloo Mesh.",
		Long:  "Create a mapping of clusters to kubeconfig entries in ${HOME}/.gloo-mesh/meshctl-config.yaml.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configure(meshctlConfigPath)
		},
	}
	pflag.StringVarP(&meshctlConfigPath, "meshctl-config-file", "f", "", "path to the meshctl config file. defaults to `$HOME/.gloo-mesh/meshctl-config.yaml`")
	return cmd
}

func configure(meshctlConfigPath string) error {
	config, err := utils.ParseMeshctlConfig(meshctlConfigPath)
	if err != nil {
		return err
	}

	keepGoing := "Yes"
	for keepGoing == "Yes" {
		prompt := promptui.Select{
			Label: "Are you configuring a management cluster or a data plane cluster?",
			Items: []string{"Management Plane", "Data Plane"},
		}
		answer, _, err := prompt.Run()
		if err != nil {
			return err
		}
		switch answer {
		case 0:
			mgmtContext, err := configureCluster()
			if err != nil {
				return err
			}
			config.AddMgmtCluster(mgmtContext)
		case 1:
			clusterName, dataPlaneContext, err := configureDataPlaneCluster()
			if err != nil {
				return err
			}
			if err := config.AddDataPlaneCluster(clusterName, dataPlaneContext); err != nil {
				return err
			}
		}
		keepGoingPrompt := promptui.Select{
			Label: "Would you like to configure another cluster?",
			Items: []string{"Yes", "No"},
		}
		_, keepGoing, err = keepGoingPrompt.Run()
		if err != nil {
			return err
		}
	}

	bytes, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(meshctlConfigPath, bytes, 0644); err != nil {
		return err
	}

	fmt.Printf("Done! Please see your configured meshctl config file at %s\n", meshctlConfigPath)
	return err
}

func configureDataPlaneCluster() (string, utils.MeshctlCluster, error) {
	var cluster utils.MeshctlCluster
	clusterNamePrompt := promptui.Prompt{
		Label: "What is your kubernetes cluster name?",
	}
	clusterName, err := clusterNamePrompt.Run()
	if err != nil {
		return "", cluster, err
	}
	cluster, err = configureCluster()
	return clusterName, cluster, err
}

func configureCluster() (utils.MeshctlCluster, error) {
	meshctlCluster := utils.MeshctlCluster{}
	validateKubeConfigExists := func(filePath string) error {
		if _, fileErr := os.Stat(filePath); fileErr != nil {
			return eris.Errorf("no kube config file found at %s", filePath)
		}
		return nil
	}
	kubeConfigFilePrompt := promptui.Prompt{
		Label:    "What is the path to your kubernetes config file?",
		Validate: validateKubeConfigExists,
	}
	kubeConfigFile, err := kubeConfigFilePrompt.Run()
	if err != nil {
		return meshctlCluster, err
	}

	clusters, err := getKubeContextOptions(kubeConfigFile)
	if err != nil {
		return meshctlCluster, err
	}
	if len(clusters) == 0 {
		return meshctlCluster, eris.Errorf("no clusters found in kubernetes config file %s", kubeConfigFile)
	}
	kubeContextPrompt := promptui.Select{
		Label: "What is the name of your kube context?",
		Items: clusters,
	}
	_, kubeContext, err := kubeContextPrompt.Run()
	if err != nil {
		return meshctlCluster, err
	}

	return utils.MeshctlCluster{
		KubeConfig:  kubeConfigFile,
		KubeContext: kubeContext,
	}, nil
}

func getKubeContextOptions(kubeConfigFile string) ([]string, error) {
	config := utils.KubeConfig{}
	if _, fileErr := os.Stat(kubeConfigFile); fileErr == nil {
		contentString, err := ioutil.ReadFile(kubeConfigFile)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(contentString, &config); err != nil {
			return nil, err
		}
	}
	var clusters []string
	for _, cluster := range config.Clusters {
		clusters = append(clusters, cluster.Name)
	}
	return clusters, nil
}
