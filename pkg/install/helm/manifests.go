package helm

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/tiller"
)

type Manifests []manifest.Manifest

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
