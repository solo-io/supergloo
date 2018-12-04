package install

import (
	"fmt"

	"github.com/solo-io/supergloo/pkg/constants"

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
					Path: constants.ConsulInstallPath,
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
		ChartLocator: &v1.HelmChartLocator{
			Kind: &v1.HelmChartLocator_ChartPath{
				ChartPath: &v1.HelmChartPath{
					Path: constants.IstioInstallPath,
				},
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
		ChartLocator: &v1.HelmChartLocator{
			Kind: &v1.HelmChartLocator_ChartPath{
				ChartPath: &v1.HelmChartPath{
					Path: constants.LinkerdInstallPath,
				},
			},
		},
	}
	installSpec.Encryption = getEncryptionFromOpts(opts)
	return installSpec
}

func generateAppMeshInstallSpecFromOpts(opts *options.Options) *v1.Mesh {
	installSpec := &v1.Mesh{
		Metadata: getMetadataFromOpts(opts),
		MeshType: &v1.Mesh_AppMesh{
			AppMesh: &v1.AppMesh{
				AwsRegion:      opts.Install.AwsRegion,
				AwsCredentials: &opts.Install.AwsSecretRef,
			},
		},
	}
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
