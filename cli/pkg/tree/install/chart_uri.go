package install

import (
	"fmt"
	"path"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/pkg/version"
)

var (
	UnreleasedWithoutOverrideErr = eris.Errorf("you must provide a Service Mesh Hub Helm chart URI via the 'file' option " +
		"when running an unreleased version of meshctl")
	ChartAndReleaseFlagErr = func(chartOverride, versionOverride string) error {
		return eris.Errorf("you may not specify both a chart with -f and a release version with --version. Received: -f=%s and --version=%s",
			chartOverride, versionOverride)
	}
	UnsupportedHelmFileExtErr = func(helmChartArchiveUri string) error {
		return eris.Errorf("unsupported file extension for Helm chart URI: [%s]. Extension must either be .tgz or .tar.gz",
			helmChartArchiveUri)
	}
)

func GetChartUri(chartOverride, versionOverride string) (string, error) {
	if chartOverride != "" && versionOverride != "" {
		return "", ChartAndReleaseFlagErr(chartOverride, versionOverride)
	}
	if !version.IsReleaseVersion() && chartOverride == "" {
		return "", UnreleasedWithoutOverrideErr
	}

	helmChartVersion := version.Version
	if versionOverride != "" {
		helmChartVersion = versionOverride
	}

	helmChartArchiveUri := fmt.Sprintf(cliconstants.HelmRepoTemplate, strings.TrimPrefix(helmChartVersion, "v"))
	if chartOverride != "" {
		helmChartArchiveUri = chartOverride
	}

	if path.Ext(helmChartArchiveUri) != ".tgz" && !strings.HasSuffix(helmChartArchiveUri, ".tar.gz") {
		return "", UnsupportedHelmFileExtErr(helmChartArchiveUri)
	}
	return helmChartArchiveUri, nil
}
