package install

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"gopkg.in/AlecAivazis/survey.v1"
)

func generateConsulInstallSpecFromOpts(opts *options.Options) *v1.Install {
	installSpec := &v1.Install{
		Metadata: getMetadataFromOpts(opts),
		MeshType: &v1.Install_Consul{
			Consul: &v1.Consul{
				InstallationNamespace: opts.Install.Namespace,
				ServerAddress:         opts.Install.ConsulServerAddress,
			},
		},
		ChartLocator: &v1.HelmChartLocator{
			Kind: &v1.HelmChartLocator_ChartPath{
				ChartPath: &v1.HelmChartPath{
					Path: "https://github.com/hashicorp/consul-helm/archive/5daf413626046d31dcb1030db889a7c96e078a1c.tar.gz",
				},
			},
		},
	}
	installSpec.Encryption = getEncryptionFromOpts(opts)

	return installSpec
}

func generateIstioInstallSpecFromOpts(opts *options.Options) *v1.Install {
	installSpec := &v1.Install{
		Metadata: getMetadataFromOpts(opts),
		MeshType: &v1.Install_Istio{
			Istio: &v1.Istio{
				InstallationNamespace: opts.Install.Namespace,
				WatchNamespaces:       opts.Install.WatchNamespaces,
			},
		},
	}
	installSpec.Encryption = getEncryptionFromOpts(opts)
	return installSpec
}

func generateLinkerd2InstallSpecFromOpts(opts *options.Options) *v1.Install {
	installSpec := &v1.Install{
		Metadata: getMetadataFromOpts(opts),
		MeshType: &v1.Install_Linkerd2{
			Linkerd2: &v1.Linkerd2{
				InstallationNamespace: opts.Install.Namespace,
				WatchNamespaces:       opts.Install.WatchNamespaces,
			},
		},
	}
	installSpec.Encryption = getEncryptionFromOpts(opts)
	return installSpec
}

func chooseWatchNamespaces(opts *options.Options) ([]string, error) {

	prompt := &survey.MultiSelect{
		Message: "Which namespaces should this mesh watch:",
		Options: opts.Cache.Namespaces,
	}

	chosenNamespaces := []string{}
	// survey.AskOne(prompt, &chosenNamespaces, nil)
	if err := survey.AskOne(prompt, &chosenNamespaces, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return []string{}, err
	}

	return chosenNamespaces, nil
}
