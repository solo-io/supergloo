package helm

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
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/tiller"
	"k8s.io/helm/pkg/timeconv"
)

const customResourceDefinitionKind = "CustomResourceDefinition"

var defaultKubeVersion = fmt.Sprintf("%s.%s", chartutil.DefaultKubeVersion.Major, chartutil.DefaultKubeVersion.Minor)

func RenderManifests(ctx context.Context, chartUri, values, releaseName, namespace, kubeVersion string, releaseIsInstall bool) ([]manifest.Manifest, error) {
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
			IsInstall: releaseIsInstall,
			IsUpgrade: !releaseIsInstall,
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
		if isEmptyManifest(man) {
			contextutils.LoggerFrom(ctx).Warnf("is an empty manifest, removing %v", file)
			delete(renderedTemplates, file)
		}
	}
	manifests := manifest.SplitManifests(renderedTemplates)
	return tiller.SortByKind(manifests), nil
}

func ApplyManifests(ctx context.Context, namespace string, manifests []manifest.Manifest) error {
	kc := kube.New(nil)

	for _, man := range manifests {
		contextutils.LoggerFrom(ctx).Infof("applying manifest %v: %v", man.Name, man.Head)

		var (
			shouldWait bool
			timeout    int64
		)
		// wait for crds. it's that easy!
		if man.Head.Kind == customResourceDefinitionKind {
			shouldWait = true
			timeout = 5
		}
		if err := kc.Create(namespace, bytes.NewBufferString(man.Content), timeout, shouldWait); err != nil {
			if kubeerrs.IsAlreadyExists(err) {
				contextutils.LoggerFrom(ctx).Warnf("already exists, skipping %v", man.Name)
				continue
			}
			return err
		}
	}

	return nil
}

func DeleteManifests(ctx context.Context, namespace string, manifests []manifest.Manifest) error {
	kc := kube.New(nil)

	for _, man := range manifests {
		contextutils.LoggerFrom(ctx).Infof("deleting manifest %v: %v", man.Name, man.Head)

		if err := kc.Delete(namespace, bytes.NewBufferString(man.Content)); err != nil {
			if kubeerrs.IsNotFound(err) {
				contextutils.LoggerFrom(ctx).Warnf("not found, skipping %v", man.Name)
				continue
			}
			return err
		}
	}

	return nil
}

var commentRegex = regexp.MustCompile("#.*")

func isEmptyManifest(manifest string) bool {
	removeComments := commentRegex.ReplaceAllString(manifest, "")
	removeNewlines := strings.Replace(removeComments, "\n", "", -1)
	removeDashes := strings.Replace(removeNewlines, "---", "", -1)
	return removeDashes == ""
}
