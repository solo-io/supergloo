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
	opts := &options{}

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure Kubernetes Clusters registered with Gloo Mesh.",
		Long: `Create a mapping of clusters to kubeconfig entries in ${HOME}/.gloo-mesh/meshctl-config.yaml.

There are two modes for this - interactive and non-interactive. Each data plane cluster should be configured with
a cluster name. Note that if a cluster is both a management and data plane cluster, it will need to be configured twice.`,
		Example: " meshctl cluster configure --disable-prompt --kubeconfig ${HOME}/.kube/config --kubecontext cluster1 ## Registers a management plane cluster\n" +
			" meshctl cluster configure --disable-prompt --cluster-name cluster2 --kubeconfig ${HOME}/.kube/config --kubecontext cluster2 ## Registers a data plane cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.meshctlConfigPath == "" {
				var err error
				opts.meshctlConfigPath, err = utils.DefaultMeshctlConfigFilePath()
				if err != nil {
					return err
				}
			}
			if opts.disablePrompt {
				if opts.kubeconfig == "" || opts.kubecontext == "" {
					return eris.Errorf("must pass in --kubeconfig and --kubecontext flags when configuring in non-interactive mode")
				}
				return configure(opts)
			}
			return configureInteractive(opts.meshctlConfigPath)
		},
	}
	opts.addToFlags(cmd.PersistentFlags())
	return cmd
}

type options struct {
	meshctlConfigPath string

	disablePrompt bool
	clusterName   string
	kubeconfig    string
	kubecontext   string
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.meshctlConfigPath, "meshctl-config-file", "c", "",
		"path to the meshctl config file. defaults to `$HOME/.gloo-mesh/meshctl-config.yaml`")
	flags.BoolVar(&o.disablePrompt, "disable-prompt", false,
		"Disable the interactive prompt. Use this to configure the meshctl config file with flags instead.")
	flags.StringVar(&o.clusterName, "cluster-name", "",
		"data plane cluster name (leave empty if this is the management cluster)")
	flags.StringVar(&o.kubeconfig, "kubeconfig", "", "Path to the kubeconfig from which the cluster will be accessed")
	flags.StringVar(&o.kubecontext, "kubecontext", "", "Name of the kubeconfig context to use for the cluster")
}

func configure(opts *options) error {
	config, err := utils.ParseMeshctlConfig(opts.meshctlConfigPath)
	if err != nil {
		return err
	}
	if err = validateKubeConfigExists(opts.kubeconfig); err != nil {
		return err
	}
	validContexts, err := getKubeContextOptions(opts.kubeconfig)
	if err != nil {
		return err
	}
	valid := false
	for _, context := range validContexts {
		if opts.kubecontext == context {
			valid = true
			break
		}
	}
	if !valid {
		return eris.Errorf("context %v does not exist in kubeconfig file %s", opts.kubecontext, opts.kubeconfig)
	}
	cluster := utils.MeshctlCluster{
		KubeConfig:  opts.kubeconfig,
		KubeContext: opts.kubecontext,
	}
	if opts.clusterName != "" {
		if err := config.AddDataPlaneCluster(opts.clusterName, cluster); err != nil {
			return err
		}
	} else {
		config.AddMgmtCluster(cluster)
	}
	return writeConfigToFile(config, opts.meshctlConfigPath)
}

var defaultConfigFile string

func configureInteractive(meshctlConfigPath string) error {
	config, err := utils.ParseMeshctlConfig(meshctlConfigPath)
	if err != nil {
		return err
	}

	keepGoing := "Yes"
	for keepGoing == "Yes" {
		answer, err := selectValueInteractive("Are you configuring a management cluster or a data plane cluster?",
			[]string{"Management Plane", "Data Plane"})
		if err != nil {
			return err
		}
		switch answer {
		case "Management Plane":
			mgmtContext, err := configureCluster()
			if err != nil {
				return err
			}
			config.AddMgmtCluster(mgmtContext)
		case "Data Plane":
			clusterName, dataPlaneContext, err := configureDataPlaneCluster()
			if err != nil {
				return err
			}
			if err := config.AddDataPlaneCluster(clusterName, dataPlaneContext); err != nil {
				return err
			}
		}
		keepGoing, err = selectValueInteractive("Would you like to configure another cluster?",
			[]string{"Yes", "No"})
		if err != nil {
			return err
		}
	}

	return writeConfigToFile(config, meshctlConfigPath)
}

func writeConfigToFile(config utils.MeshctlConfig, meshctlConfigPath string) error {
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
	kubeConfigFilePrompt := promptui.Prompt{
		Label:    "What is the path to your kubernetes config file?",
		Validate: validateKubeConfigExists,
		Default:  defaultConfigFile,
	}
	kubeConfigFile, err := kubeConfigFilePrompt.Run()
	if err != nil {
		return meshctlCluster, err
	}
	defaultConfigFile = kubeConfigFile

	clusters, err := getKubeContextOptions(kubeConfigFile)
	if err != nil {
		return meshctlCluster, err
	}
	if len(clusters) == 0 {
		return meshctlCluster, eris.Errorf("no clusters found in kubernetes config file %s", kubeConfigFile)
	}
	kubeContext, err := selectValueInteractive("What is the name of your kube context?", clusters)
	if err != nil {
		return meshctlCluster, err
	}

	return utils.MeshctlCluster{
		KubeConfig:  kubeConfigFile,
		KubeContext: kubeContext,
	}, nil
}

func selectValueInteractive(message string, options interface{}) (string, error) {
	prompt := promptui.Select{
		Label: message,
		Items: options,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

func validateKubeConfigExists(filePath string) error {
	if _, fileErr := os.Stat(filePath); fileErr != nil {
		return eris.Errorf("no kube config file found at %s", filePath)
	}
	return nil
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
