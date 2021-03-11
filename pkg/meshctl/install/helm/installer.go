package helm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rotisserie/eris"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

const (
	helmNamespaceEnvVar      = "HELM_NAMESPACE"
	tempChartFilePermissions = 0644
)

type Installer struct {
	KubeConfig  string
	KubeContext string
	ChartUri    string
	Namespace   string
	ReleaseName string
	ValuesFile  string
	Values      map[string]string
	Verbose     bool
	DryRun      bool
}

func (i Installer) InstallChart(ctx context.Context) error {
	kubeConfig := i.KubeConfig
	kubeContext := i.KubeContext
	chartUri := i.ChartUri
	namespace := i.Namespace
	releaseName := i.ReleaseName
	valuesFile := i.ValuesFile
	verbose := i.Verbose
	dryRun := i.DryRun

	kubeClient, err := utils.BuildClient(kubeConfig, kubeContext)
	if err != nil {
		return err
	}

	if !dryRun {
		if err = ensureNamespace(ctx, kubeClient, namespace); err != nil {
			return eris.Wrapf(err, "creating namespace")
		}
	}

	actionConfig, settings, err := newActionConfig(kubeConfig, kubeContext, namespace)
	if err != nil {
		return eris.Wrapf(err, "creating helm config")
	}
	settings.Debug = verbose
	settings.KubeConfig = kubeConfig
	settings.KubeContext = kubeContext

	chartObj, err := downloadChart(chartUri)
	if err != nil {
		return eris.Wrapf(err, "loading chart file")
	}

	// Merge values provided via the '--values' flag
	valueOpts := &values.Options{}
	if valuesFile != "" {
		valueOpts.ValueFiles = []string{valuesFile}
	}
	for key, value := range i.Values {
		valueOpts.Values = append(valueOpts.Values, key+"="+value)
	}
	parsedValues, err := valueOpts.MergeValues(getter.All(settings))
	if err != nil {
		return eris.Wrapf(err, "parsing values")
	}

	// must apply CRDs before installing since the Helm chart will apply CRD objects
	if !dryRun {
		if err = upsertCrds(ctx, kubeClient, chartObj); err != nil {
			return eris.Wrapf(err, "updating CRDs")
		}
	}

	isUpgrade := false
	var release *release.Release

	h, err := actionConfig.Releases.History(releaseName)
	if err == nil && len(h) > 0 {
		client := action.NewUpgrade(actionConfig)
		client.Namespace = namespace
		client.DryRun = dryRun
		isUpgrade = true

		release, err = client.Run(releaseName, chartObj, parsedValues)
		if err != nil {
			return eris.Wrapf(err, "upgrading helm chart")
		}
	} else {
		// release does not exist, perform install

		client := action.NewInstall(actionConfig)
		client.ReleaseName = releaseName
		client.Namespace = namespace
		client.DryRun = dryRun
		if dryRun {
			client.ClientOnly = true
		}

		release, err = client.Run(chartObj, parsedValues)
		if err != nil {
			return eris.Wrapf(err, "installing helm chart")
		}
	}

	updateReleaseManifestWithCrds(chartObj, release)
	// output to stdout
	output(release, dryRun, isUpgrade)

	return nil
}

func output(release *release.Release, dryRun bool, isUpgrade bool) {
	if dryRun {
		// dry run should only output a pipe-able manifest
		fmt.Printf("%v", release.Manifest)
	} else {
		verb := "installing"
		if isUpgrade {
			verb = "upgrading"
		}
		logrus.Infof(
			"Finished %s chart '%s' as release %s:%s",
			verb, release.Chart.Name(), release.Namespace, release.Name,
		)
		logrus.Debugf("%v", release.Manifest)
	}
}

// Helm does not update CRDs upon upgrade, https://github.com/helm/helm/issues/7735
// so we need to update the CRDs ourselves, during both install and upgrade
func upsertCrds(ctx context.Context, kubeClient client.Client, chartObj *chart.Chart) error {
	// unmarshal CRD definitions from Helm manifests
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	crdManifests := chartObj.CRDObjects()
	for _, crdManifest := range crdManifests {
		var crds []*unstructured.Unstructured
		crdsRaw := strings.Split(string(crdManifest.File.Data), "\n---")

		for _, crdRaw := range crdsRaw {
			crd := &unstructured.Unstructured{}
			if _, _, err := decoder.Decode([]byte(crdRaw), nil, crd); err != nil {
				return err
			}
			crds = append(crds, crd)
		}

		// upsert each CRD
		for _, crd := range crds {
			crd := crd
			err := kubeClient.Create(ctx, crd)
			if errors.IsAlreadyExists(err) {
				// update requires manually setting the resource version
				existingCrd := &v1beta1.CustomResourceDefinition{}
				if err := kubeClient.Get(ctx, client.ObjectKey{Name: crd.GetName()}, existingCrd); err != nil {
					return err
				}
				crd.SetResourceVersion(existingCrd.GetResourceVersion())
				if err = kubeClient.Update(ctx, crd); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
	}
	return nil
}

func updateReleaseManifestWithCrds(chartObj *chart.Chart, release *release.Release) {
	manifest := bytes.NewBuffer([]byte{})
	for _, crd := range chartObj.CRDObjects() {
		fmt.Fprintf(manifest, "---\n# Source: %s\n%s\n", crd.Name, string(crd.File.Data[:]))
	}
	fmt.Fprintf(manifest, release.Manifest)
	release.Manifest = manifest.String()
}

func ensureNamespace(ctx context.Context, kubeClient client.Client, namespace string) error {
	namespaces := v1.NewNamespaceClient(kubeClient)
	return namespaces.UpsertNamespace(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Spec: corev1.NamespaceSpec{Finalizers: []corev1.FinalizerName{"kubernetes"}},
	})
}

// Returns an action configuration that can be used to create Helm actions and the Helm env settings.
// We currently get the Helm storage driver from the standard HELM_DRIVER env (defaults to 'secret').
func newActionConfig(kubeConfig, kubeContext, namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	actionConfig := new(action.Configuration)

	settings := newCLISettings(kubeConfig, kubeContext, namespace)

	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), logrus.Debugf); err != nil {
		return nil, nil, err
	}
	return actionConfig, settings, nil
}

// Build a Helm EnvSettings struct
// basically, abstracted cli.New() into our own function call because of the weirdness described in the big comment below
func newCLISettings(kubeConfig, kubeContext, namespace string) *cli.EnvSettings {
	// The installation namespace is expressed as a "config override" in the Helm internals
	// It's normally set by the --namespace flag when invoking the Helm binary, which ends up
	// setting a non-exported field in the Helm settings struct (https://github.com/helm/helm/blob/v3.0.1/pkg/cli/environment.go#L77)
	// However, we are not invoking the Helm binary, so that field doesn't get set. It is left as "", which means
	// that any resources that are non-namespaced (at the time of writing, some of Prometheus's resources do not
	// have a namespace attached to them but they probably should) wind up in the default namespace from YOUR
	// kube config. To get around this, we temporarily set an env var before the Helm settings are initialized
	// so that the proper namespace override is piped through. (https://github.com/helm/helm/blob/v3.0.1/pkg/cli/environment.go#L64)
	if os.Getenv(helmNamespaceEnvVar) == "" {
		os.Setenv(helmNamespaceEnvVar, namespace)
		defer os.Setenv(helmNamespaceEnvVar, "")
	}

	settings := cli.New()
	settings.KubeContext = kubeContext
	settings.KubeConfig = kubeConfig

	return settings
}

func downloadChart(chartArchiveUri string) (*chart.Chart, error) {
	charFilePath := ""
	if fi, err := os.Stat(chartArchiveUri); err == nil && fi.IsDir() {
		charFilePath = chartArchiveUri
	} else {

		// 1. Get a reader to the chart file (remote URL or local file path)
		chartFileReader, err := getResource(chartArchiveUri)
		if err != nil {
			return nil, err
		}
		defer func() { _ = chartFileReader.Close() }()

		// 2. Write chart to a temporary file
		chartBytes, err := ioutil.ReadAll(chartFileReader)
		if err != nil {
			return nil, err
		}

		chartFile, err := ioutil.TempFile("", "temp-helm-chart")
		if err != nil {
			return nil, err
		}
		charFilePath = chartFile.Name()
		defer func() { _ = os.RemoveAll(charFilePath) }()

		if err := ioutil.WriteFile(charFilePath, chartBytes, tempChartFilePermissions); err != nil {
			return nil, err
		}
	}

	// 3. Load the chart file
	chartObj, err := loader.Load(charFilePath)
	if err != nil {
		return nil, err
	}

	return chartObj, nil
}

// Get the resource identified by the given URI.
// The URI can either be an http(s) address or a relative/absolute file path.
func getResource(uri string) (io.ReadCloser, error) {
	var file io.ReadCloser
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		resp, err := http.Get(uri)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, eris.Errorf("http GET returned status %d for resource %s", resp.StatusCode, uri)
		}

		file = resp.Body
	} else {
		path, err := filepath.Abs(uri)
		if err != nil {
			return nil, eris.Wrapf(err, "getting absolute path for %v", uri)
		}

		f, err := os.Open(path)
		if err != nil {
			return nil, eris.Wrapf(err, "opening file %v", path)
		}
		file = f
	}

	// Write the body to file
	return file, nil
}
