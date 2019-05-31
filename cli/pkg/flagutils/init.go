package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/pkg/version"
	"github.com/spf13/pflag"
)

func AddInitFlags(set *pflag.FlagSet, init *options.Init) {
	set.StringVarP(&init.HelmChartOverride, "file", "f", "", "Install SuperGloo from this Helm chart location (file path or URL). Target file must be a tarball")
	set.StringVarP(&init.HelmValues, "values", "v", "", "Provide a custom values.yaml to override default values in the helm chart. Leave empty to use default values.")
	set.StringVarP(&init.InstallNamespace, "namespace", "n", "supergloo-system", "Namespace to install supergloo into")
	if !version.IsReleaseVersion() {
		set.StringVar(&init.ReleaseVersion, "release", "", "install from this release version. Should correspond with the "+
			"name of the release on GitHub")
	}
	set.BoolVarP(&init.DryRun, "dry-run", "d", false, "Dump the raw installation yaml instead of applying it to kubernetes")
}
