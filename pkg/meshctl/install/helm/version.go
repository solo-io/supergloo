package helm

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-version"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func GetLatestChartVersion(repoURI, chartName string, stable bool) (string, error) {
	return getLatestChartVersion(repoURI, chartName, func(version version.Version) bool {
		return !stable || version.Prerelease() == ""
	})
}

func GetLatestChartMinorVersion(repoURI, chartName string, stable bool, major, minor int) (string, error) {
	return getLatestChartVersion(repoURI, chartName, func(version version.Version) bool {
		return version.Segments()[0] == major && version.Segments()[1] == minor &&
			(!stable || version.Prerelease() == "")
	})
}

func getLatestChartVersion(
	repoURI, chartName string,
	isVersionCopmatible func(version version.Version) bool,
) (string, error) {
	versions, err := getChartVersions(repoURI, chartName)
	if err != nil {
		return "", nil
	}
	latestVersion := version.Must(version.NewVersion("0"))
	foundVersion := false
	for i, version := range versions {
		if !isVersionCopmatible(version) {
			continue
		}

		if version.GreaterThan(latestVersion) {
			latestVersion = &versions[i]
			foundVersion = true
		}
	}
	if !foundVersion {
		logrus.Debug("available versions: %v", versions)
		return "", eris.New("chart version not found")
	}

	return latestVersion.Original(), nil
}

func getChartVersions(repoURI, chartName string) ([]version.Version, error) {
	res, err := http.Get(fmt.Sprintf("%s/%s/index.yaml", repoURI, chartName))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, eris.Wrap(err, "unable to read response body")
	}
	if res.StatusCode != http.StatusOK {
		logrus.Debug(string(b))
		return nil, eris.Errorf("invalid response from the Helm repository: %d %s", res.StatusCode, res.Status)
	}
	index := struct {
		Entries map[string][]struct {
			Version string `yaml:"version"`
		} `yaml:"entries"`
	}{}
	if err := yaml.Unmarshal(b, &index); err != nil {
		return nil, err
	}
	chartReleases, ok := index.Entries[chartName]
	if !ok {
		logrus.Debug(string(b))
		return nil, eris.Errorf("chart not found in index: %s", chartName)
	}
	versions := make([]version.Version, 0, len(chartReleases))
	for _, release := range chartReleases {
		version, err := version.NewVersion(release.Version)
		if err != nil {
			logrus.Warnf("invalid release version: %s", release.Version)
			continue
		}
		versions = append(versions, *version)
	}

	return versions, nil
}
