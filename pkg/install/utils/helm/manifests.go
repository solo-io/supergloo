package helm

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"gopkg.in/yaml.v2"

	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/tiller"
)

type Manifests []manifest.Manifest

func (m Manifests) Find(name string) *manifest.Manifest {
	for _, man := range m {
		if man.Name == name {
			return &man
		}
	}
	return nil
}

func (m Manifests) CombinedString() string {
	buf := &bytes.Buffer{}

	for _, m := range tiller.SortByKind(m) {
		data := m.Content
		b := filepath.Base(m.Name)
		if b == "NOTES.txt" {
			continue
		}
		if strings.HasPrefix(b, "_") {
			continue
		}
		fmt.Fprintf(buf, "---\n# Source: %s\n", m.Name)
		fmt.Fprintln(buf, data)
	}

	return buf.String()
}

// returns the set of manifests as a gzipped+base64'ed string
func (m Manifests) Gzipped() (string, error) {
	if len(m) == 0 {
		return "", nil
	}
	raw, err := yaml.Marshal(m)
	if err != nil {
		return "", errors.Wrapf(err, "marshalling manifests to yaml")
	}
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	if _, err := gz.Write(raw); err != nil {
		return "", errors.Wrapf(err, "compressing yaml")
	}
	if err := gz.Flush(); err != nil {
		return "", errors.Wrapf(err, "flushing gzip writer")
	}
	if err := gz.Close(); err != nil {
		return "", errors.Wrapf(err, "closing gzip writer")
	}
	asBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return asBase64, nil
}

// returns creates a set of manifests from a gzipped, base64 encoded string
func NewManifestsFromGzippedString(gzippedString string) (Manifests, error) {
	if gzippedString == "" {
		return nil, nil
	}
	base64Decoded, err := base64.StdEncoding.DecodeString(gzippedString)
	if err != nil {
		return nil, errors.Wrapf(err, "decoding base64 string")
	}
	gz, err := gzip.NewReader(bytes.NewBuffer(base64Decoded))
	if err != nil {
		return nil, errors.Wrapf(err, "creating gzip reader")
	}

	raw, err := ioutil.ReadAll(gz)
	if err != nil {
		return nil, errors.Wrapf(err, "decoding gzipped string")
	}

	var m Manifests
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return nil, errors.Wrapf(err, "unmarshalling manifests as yaml")
	}
	return m, nil
}

func (m Manifests) SplitByCrds() (Manifests, Manifests) {
	var crdManifests, nonCrdManifests Manifests
	for _, man := range m {
		if isCrdManifest(man) {
			crdManifests = append(crdManifests, man)
		} else {
			nonCrdManifests = append(nonCrdManifests, man)
		}
	}
	return crdManifests, nonCrdManifests
}

const customResourceDefinitionKind = "CustomResourceDefinition"

func isCrdManifest(man manifest.Manifest) bool {
	return man.Head.Kind == customResourceDefinitionKind
}
