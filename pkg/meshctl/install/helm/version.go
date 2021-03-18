package helm

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/hashicorp/go-version"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func GetLatestChartVersion(repoURI, chartName string, stable bool) (string, error) {
	return getLatestChartVersion(repoURI, chartName, func(version version.Version) bool {
		// Do not allow prereleaes if stable is true.
		return !stable || version.Prerelease() == ""
	})
}

func GetLatestChartMinorVersion(repoURI, chartName string, stable bool, major, minor int) (string, error) {
	return getLatestChartVersion(repoURI, chartName, func(version version.Version) bool {
		// Compatible versions will have the given major and minor version.
		// Do not allow prereleaes if stable is true.
		return version.Segments()[0] == major && version.Segments()[1] == minor &&
			(!stable || version.Prerelease() == "")
	})
}

func getLatestChartVersion(
	repoURI, chartName string,
	isVersionCompatible func(version version.Version) bool,
) (string, error) {
	versions, err := getChartVersions(repoURI, chartName)
	if err != nil {
		return "", nil
	}
	logrus.Debugf("available versions: %v", versions)

	for _, version := range versions {
		if isVersionCompatible(*version) {
			logrus.Debugf("installing chart version %s", version.Original())
			return version.Original(), nil
		}
	}

	return "", eris.New("compatible chart version not found")
}

func getChartVersions(repoURI, chartName string) (version.Collection, error) {
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
	versions := make(version.Collection, 0, len(chartReleases))
	for _, release := range chartReleases {
		version, err := version.NewVersion(release.Version)
		if err != nil {
			logrus.Warnf("invalid release version: %s", release.Version)
			continue
		}
		versions = append(versions, version)
	}

	sort.Sort(sort.Reverse(versions)) // Sort from newest to oldest

	return versions, nil
}
