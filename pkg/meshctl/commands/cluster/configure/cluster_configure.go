package configure

import (
	"context"
	"fmt"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
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

	keepGoing := "y"
	for keepGoing == "y" {
		// TODO replace fmt.Scanln with promptui
		var answer string
		fmt.Println("Are you configuring the management cluster? (y/n)")
		fmt.Scanln(&answer)
		if strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes" {
			mgmtContext := configureCluster()
			config.AddMgmtCluster(mgmtContext)
		} else if strings.ToLower(answer) == "n" || strings.ToLower(answer) == "no" {
			remoteContext := configureCluster()
			/*TODO: get the cluster name as a user input*/
			if err := config.AddDataPlaneCluster(clusterName, remoteContext); err != nil{
				return err
			}
		}
		fmt.Println("Would you like to configure another cluster? (y/n)")
		fmt.Scanln(&keepGoing)
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

func configureCluster() KubeCluster {
	context := KubeCluster{}
	fmt.Println("What is the path to your kubernetes config file?")
	fmt.Scanln(&context.KubeConfig)
	fmt.Println("What is the name of your context?")
	fmt.Scanln(&context.KubeContext)
	return context
}
