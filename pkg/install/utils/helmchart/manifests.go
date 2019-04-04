package helmchart

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/timeconv"

	"k8s.io/apimachinery/pkg/runtime"
	yaml2json "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/solo-io/supergloo/pkg/install/utils/kuberesource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

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

func (m Manifests) ResourceList() (kuberesource.UnstructuredResources, error) {
	snippets := strings.Split(m.CombinedString(), "---")

	var resources kuberesource.UnstructuredResources
	for _, objectYaml := range snippets {
		if IsEmptyManifest(objectYaml) {
			continue
		}
		jsn, err := yaml2json.ToJSON([]byte(objectYaml))
		if err != nil {
			return nil, err
		}
		uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsn)
		if err != nil {
			return nil, err
		}
		resources = append(resources, uncastObj.(*unstructured.Unstructured))
	}

	return resources, nil
}

var commentRegex = regexp.MustCompile("#.*")

func IsEmptyManifest(manifest string) bool {
	removeComments := commentRegex.ReplaceAllString(manifest, "")
	removeNewlines := strings.Replace(removeComments, "\n", "", -1)
	removeDashes := strings.Replace(removeNewlines, "---", "", -1)
	removeSpaces := strings.Replace(removeDashes, " ", "", -1)
	return removeSpaces == ""
}

var defaultKubeVersion = fmt.Sprintf("%s.%s", chartutil.DefaultKubeVersion.Major, chartutil.DefaultKubeVersion.Minor)

func RenderManifests(ctx context.Context, chartUri, values, releaseName, namespace, kubeVersion string) (Manifests, error) {
	var file io.Reader
	if strings.HasPrefix(chartUri, "http://") || strings.HasPrefix(chartUri, "https://") {
		resp, err := http.Get(chartUri)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.Errorf("http GET returned status %d", resp.StatusCode)
		}

		file = resp.Body
	} else {
		path, err := filepath.Abs(chartUri)
		if err != nil {
			return nil, errors.Wrapf(err, "getting absolute path for %v", chartUri)
		}

		f, err := os.Open(path)
		if err != nil {
			return nil, errors.Wrapf(err, "opening file %v", path)
		}
		file = f
	}

	if kubeVersion == "" {
		kubeVersion = defaultKubeVersion
	}
	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      releaseName,
			IsInstall: true,
			Time:      timeconv.Now(),
			Namespace: namespace,
		},
		KubeVersion: kubeVersion,
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	c, err := chartutil.LoadArchive(file)
	if err != nil {
		return nil, errors.Wrapf(err, "loading chart")
	}

	config := &chart.Config{Raw: values, Values: map[string]*chart.Value{}}
	renderedTemplates, err := renderutil.Render(c, config, renderOpts)
	if err != nil {
		return nil, err
	}

	for file, man := range renderedTemplates {
		if IsEmptyManifest(man) {
			contextutils.LoggerFrom(ctx).Warnf("is an empty manifest, removing %v", file)
			delete(renderedTemplates, file)
		}
	}
	manifests := manifest.SplitManifests(renderedTemplates)
	return tiller.SortByKind(manifests), nil
}
