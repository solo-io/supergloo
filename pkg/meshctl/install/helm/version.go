package helm

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-version"
	"github.com/rotisserie/eris"
	"gopkg.in/yaml.v2"
)

func GetLatestChartVersion(repoURI, chartName string, stable bool) (string, error) {
	res, err := http.Get(fmt.Sprintf("%s/%s/index.yaml", repoURI, chartName))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, res.Body)
		return "", fmt.Errorf("invalid response from the Helm repository: %d %s", res.StatusCode, res.Status)
	}
	index := struct {
		Entries map[string][]struct {
			Version string `yaml:"version"`
		} `yaml:"entries"`
	}{}
	if err := yaml.NewDecoder(res.Body).Decode(&index); err != nil {
		return "", err
	}
	entries, ok := index.Entries[chartName]
	if !ok {
		return "", fmt.Errorf("no entry found for chart: %s", chartName)
	}
	if len(entries) == 0 {
		return "", eris.New("no versions found")
	}

	// entries are sorted by version so the first will have the latest

	if stable {
		// To install the latest stable version, we will iterate through the entry versions
		// until we find one with no prerelease descriptors (e.g. beta, rc).
		for _, entry := range entries {
			if entryVersion, err := version.NewVersion(entry.Version); err != nil {
				// skip this invalid version
			} else {
				if entryVersion.Prerelease() == "" {
					return entry.Version, nil
				}
			}
		}
	}

	return entries[0].Version, nil
}
