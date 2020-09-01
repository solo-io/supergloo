package helm

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/helm/internal"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
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
	KubeConfig  *register.KubeCfg
	ChartUri    string
	Namespace   string
	ReleaseName string
	ValuesFile  string
	Verbose     bool
	DryRun      bool
}

func (i Installer) InstallChart(ctx context.Context) error {
	kubeConfig := i.KubeConfig
	chartUri := i.ChartUri
	namespace := i.Namespace
	releaseName := i.ReleaseName
	valuesFile := i.ValuesFile
	verbose := i.Verbose
	dryRun := i.DryRun

	err := ensureNamespace(ctx, kubeConfig, namespace)
	if err != nil {
		return eris.Wrapf(err, "creating namespace")
	}

	actionConfig, settings, err := newActionConfig(kubeConfig, namespace)
	if err != nil {
		return eris.Wrapf(err, "creating helm config")
	}
	settings.Debug = verbose

	chartObj, err := downloadChart(chartUri)
	if err != nil {
		return eris.Wrapf(err, "loading chart file")
	}

	// Merge values provided via the '--values' flag
	valueOpts := &values.Options{}
	if valuesFile != "" {
		valueOpts.ValueFiles = []string{valuesFile}
	}
	parsedValues, err := valueOpts.MergeValues(getter.All(settings))
	if err != nil {
		return eris.Wrapf(err, "parsing values")
	}

	h, err := actionConfig.Releases.History(releaseName)
	if err == nil && len(h) > 0 {
		client := action.NewUpgrade(actionConfig)
		client.Namespace = namespace
		client.DryRun = dryRun

		release, err := client.Run(releaseName, chartObj, parsedValues)
		if err != nil {
			return eris.Wrapf(err, "installing helm chart")
		}

		logrus.Infof("finished upgrading chart as release: %+v", release)

	} else {
		// release does not exist, perform install

		client := action.NewInstall(actionConfig)
		client.ReleaseName = releaseName
		client.Namespace = namespace
		client.DryRun = dryRun

		release, err := client.Run(chartObj, parsedValues)
		if err != nil {
			return eris.Wrapf(err, "installing helm chart")
		}

		logrus.Infof("finished installing chart as release: %+v", release)
	}

	return nil
}

func ensureNamespace(ctx context.Context, kubeConfig *register.KubeCfg, namespace string) error {
	cfg, err := kubeConfig.ConstructRestConfig()
	if err != nil {
		return err
	}
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return err
	}
	namespaces := v1.NewNamespaceClient(c)
	return namespaces.UpsertNamespace(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Spec: corev1.NamespaceSpec{Finalizers: []corev1.FinalizerName{"kubernetes"}},
	})
}

// Returns an action configuration that can be used to create Helm actions and the Helm env settings.
// We currently get the Helm storage driver from the standard HELM_DRIVER env (defaults to 'secret').
func newActionConfig(
	kubeConfig *register.KubeCfg,
	namespace string,
) (*action.Configuration, *cli.EnvSettings, error) {
	if kubeConfig.GetKubeCfgDisk() != nil {
		kubeCfgDisk := kubeConfig.GetKubeCfgDisk()
		return newActionConfigFromDisk(kubeCfgDisk.KubeConfigPath, kubeCfgDisk.KubeContext, namespace)
	} else if kubeConfig.GetClientConfig() != nil {
		return newActionConfigFromMemory(kubeConfig.GetClientConfig(), namespace)
	}
	return nil, nil, eris.New("KubeCfg must be provided as on-disk Kubeconfig or clientConfig")
}

// Returns an action configuration from a kubeconfig on disk.
func newActionConfigFromDisk(kubeConfigPath, kubeContext, namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	actionConfig := new(action.Configuration)

	settings := newCLISettings(kubeConfigPath, kubeContext, namespace)

	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), logrus.Debugf); err != nil {
		return nil, nil, err
	}
	settings.KubeConfig = kubeConfigPath
	settings.KubeContext = kubeContext
	return actionConfig, settings, nil
}

// Return an action configuration from an in-memory kubeconfig
func newActionConfigFromMemory(config clientcmd.ClientConfig, namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	var noOpDebugLog = func(_ string, _ ...interface{}) {}
	settings := newCLISettings("", "", namespace)
	restClientGetter := internal.NewInMemoryRESTClientGetter(config)
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(restClientGetter, namespace, os.Getenv("HELM_DRIVER"), noOpDebugLog); err != nil {
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
	charFilePath := chartFile.Name()
	defer func() { _ = os.RemoveAll(charFilePath) }()

	if err := ioutil.WriteFile(charFilePath, chartBytes, tempChartFilePermissions); err != nil {
		return nil, err
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
