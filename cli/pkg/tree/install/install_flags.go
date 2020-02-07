package install

import (
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddInstallFlags(set *pflag.FlagSet, opts *options.Options) {
	set.BoolVarP(&opts.SmhInstall.DryRun, "dry-run", "d", false, "Send the raw installation yaml to stdout instead of applying it to kubernetes")
	set.StringVarP(&opts.SmhInstall.HelmChartOverride, "file", "f", "", "Install Service Mesh Hub from this Helm chart archive file rather than from a release")
	set.StringSliceVarP(&opts.SmhInstall.HelmChartValueFileNames, "values", "", []string{}, "List of files with value overrides for the Service Mesh Hub Helm chart, (e.g. --values file1,file2 or --values file1 --values file2)")
	set.StringVar(&opts.SmhInstall.HelmReleaseName, "release-name", cliconstants.ReleaseName, "Helm release name")
	set.StringVar(&opts.SmhInstall.Version, "version", "", "Version to install (e.g. v1.2.0, defaults to latest)")
	set.BoolVar(&opts.SmhInstall.CreateNamespace, "create-namespace", true, "Create the namespace to install Service Mesh Hub into")
}
