package configure

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure Kubernetes Clusters registered with Gloo Mesh.",
		Long:  "Create a mapping of clusters to kubeconfig entries in ${HOME}/.gloo-mesh/meshctl-config.yaml.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configure(opts)
		},
	}
	opts.addToFlags(cmd.PersistentFlags())
	return cmd
}

type options struct {
	MeshctlConfigPath string
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.MeshctlConfigPath, "meshctl-config-file", "f", "", "path to the meshctl config file")
}

func configure(opts *options) error {
	var err error
	if opts.MeshctlConfigPath == "" {
		opts.MeshctlConfigPath, err = MeshctlConfigFilePath()
		if err != nil {
			return err
		}
	}
	config, err := ParseMeshctlConfig(opts.MeshctlConfigPath)
	if err != nil {
		return err
	}
	configFile, err := os.OpenFile(opts.MeshctlConfigPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer configFile.Close()

	if len(config.Clusters) == 0 {
		config.ApiVersion = "v1"
		config.Clusters = map[string]KubeCluster{}
	} else if config.ApiVersion != "v1" {
		return fmt.Errorf("unrecognized api version: %v", config.ApiVersion)
	}

	keepGoing := "y"
	for keepGoing == "y" {
		var answer string
		fmt.Println("Are you configuring the management cluster? (y/n)")
		fmt.Scanln(&answer)
		if strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes" {
			mgmtContext := configureCluster()
			config.Clusters[managementPlane] = mgmtContext
		} else if strings.ToLower(answer) == "n" || strings.ToLower(answer) == "no" {
			remoteContext := configureCluster()
			config.Clusters[remoteContext.KubeContext] = remoteContext
		}
		fmt.Println("Would you like to configure another cluster? (y/n)")
		fmt.Scanln(&keepGoing)
	}

	d, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}
	_, err = configFile.Write(d)

	fmt.Printf("Done! Please see your configured meshctl config file at %s\n", opts.MeshctlConfigPath)
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
