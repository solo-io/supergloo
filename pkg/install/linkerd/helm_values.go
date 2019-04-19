package linkerd

import (
	"github.com/linkerd/linkerd2/controller/gen/config"
	"github.com/solo-io/go-utils/errors"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

const chartPath_stable221 = "https://storage.googleapis.com/supergloo-charts/linkerd-stable-2.3.0.tgz"

func (o *installOpts) chartURI() (string, error) {
	switch o.installVersion {
	case Version_stable230:
		return chartPath_stable221, nil
	}
	return "", errors.Errorf("version %v is not a supported linkerd version. supported: %v", o.installVersion, supportedVersions)
}

func (o *installOpts) values(kube kubernetes.Interface) (*injector, string, error) {

	opts := newInstallOptionsWithDefaults(o.installNamespace)
	opts.proxyAutoInject = o.enableAutoInject
	if o.enableMtls {
		// cannot currently disable tls in
	}

	var values *installValues
	var cfg *config.All
	if linkerdAlreadyInstalled(o.installNamespace, kube) {
		var err error
		opts := newUpgradeOptions(opts)
		values, cfg, err = opts.validateAndBuild(o.installNamespace, kube)
		if err != nil {
			return nil, "", err
		}
	} else {
		var err error
		values, cfg, err = opts.validateAndBuild()
		if err != nil {
			return nil, "", err
		}
	}

	rawYaml, err := yaml.Marshal(values)
	if err != nil {
		return nil, "", err
	}
	injector := &injector{
		configs: cfg,
		proxyOutboundCapacity: map[string]uint{
			values.PrometheusImage: prometheusProxyOutboundCapacity,
		},
	}
	return injector, string(rawYaml), nil
}
