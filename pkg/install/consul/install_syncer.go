package consul

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/helm/helm/pkg/getter"
	"github.com/helm/helm/pkg/strvals"
	helmlib "k8s.io/helm/pkg/helm"

	"github.com/ghodss/yaml"

	"github.com/solo-io/supergloo/pkg/api/v1"

	"github.com/solo-io/supergloo/pkg/install/helm"
)

type ConsulInstallSyncer struct{}

func (c *ConsulInstallSyncer) Sync(_ context.Context, snap *v1.InstallSnapshot) error {

	// See hack/install/consul/install-on... for steps
	// 1. Create a namespace / project for consul (this step should go away and we use provided namespace)
	// 2. Set up ClusterRoleBinding for consul in that namespace (not currently done)
	// 3. Install Consul via helm chart
	// 4. Fix incorrect configuration -> webhook service looking for wrong adapter config name (this should move into an override)

	for _, install := range snap.Installs.List() {
		if install.Consul != nil {
			// helm install
			helmClient, err := helm.GetHelmClient()
			if err != nil {
				helm.Teardown() // just in case
				return err
			}

			namespace := "consul"

			// TODO: just provide these in code
			overrideFiles := []string{"/Users/rick/code/src/github.com/solo-io/supergloo/hack/install/consul/helm-overrides.yaml"}

			strs := []string{}

			overrides, err := vals(overrideFiles, strs, strs, strs, "", "", "")

			if err != nil {
				helm.Teardown() // just in case
				return err
			}

			response, err := helmClient.InstallRelease(
				install.Consul.Path,
				namespace,
				helmlib.ValueOverrides(overrides))

			fmt.Printf("Response from helm installation: %v", response)

			helm.Teardown()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type valueFiles []string

// vals merges values from files specified via -f/--values and
// directly via --set or --set-string or --set-file, marshaling them to YAML
func vals(valueFiles valueFiles, values []string, stringValues []string, fileValues []string, CertFile, KeyFile, CAFile string) ([]byte, error) {
	base := map[string]interface{}{}

	// User specified a values files via -f/--values
	for _, filePath := range valueFiles {
		currentMap := map[string]interface{}{}

		var bytes []byte
		var err error
		if strings.TrimSpace(filePath) == "-" {
			bytes, err = ioutil.ReadAll(os.Stdin)
		} else {
			bytes, err = readFile(filePath, CertFile, KeyFile, CAFile)
		}

		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	// User specified a value via --set-string
	for _, value := range stringValues {
		if err := strvals.ParseIntoString(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set-string data: %s", err)
		}
	}

	// User specified a value via --set-file
	for _, value := range fileValues {
		reader := func(rs []rune) (interface{}, error) {
			bytes, err := readFile(string(rs), CertFile, KeyFile, CAFile)
			return string(bytes), err
		}
		if err := strvals.ParseIntoFile(value, base, reader); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set-file data: %s", err)
		}
	}

	return yaml.Marshal(base)
}

// Merges source and destination map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

//readFile load a file from the local directory or a remote file with a url.
func readFile(filePath, CertFile, KeyFile, CAFile string) ([]byte, error) {
	u, _ := url.Parse(filePath)
	p := getter.All(helm.Settings)

	// FIXME: maybe someone handle other protocols like ftp.
	getterConstructor, err := p.ByScheme(u.Scheme)

	if err != nil {
		return ioutil.ReadFile(filePath)
	}

	getter, err := getterConstructor(filePath, CertFile, KeyFile, CAFile)
	if err != nil {
		return []byte{}, err
	}
	data, err := getter.Get(filePath)
	return data.Bytes(), err
}
