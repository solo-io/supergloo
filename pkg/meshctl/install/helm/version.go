package helm

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/rotisserie/eris"
	"gopkg.in/yaml.v2"
)

func GetLatestChartVersion(repoURI, chartName string) (string, error) {
	res, err := http.Get(filepath.Join(repoURI, chartName, "index.yaml"))
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

	return entries[0].Version, nil
}
