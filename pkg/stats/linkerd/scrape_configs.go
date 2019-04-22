package linkerd

import (
	"bytes"
	"text/template"

	"github.com/prometheus/prometheus/config"
	"github.com/solo-io/go-utils/errors"
	"gopkg.in/yaml.v2"
)

func PrometheusScrapeConfigs(linkerdNamespace string) ([]*config.ScrapeConfig, error) {
	buf := &bytes.Buffer{}
	if err := linkerdScrapeConfigsYamlTemplate.Execute(buf, struct {
		Namespace string
	}{
		Namespace: linkerdNamespace,
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to execute linkerdScrapeConfigsYaml template")
	}
	var scrapeConfigs []*config.ScrapeConfig
	if err := yaml.Unmarshal(buf.Bytes(), &scrapeConfigs); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal linkerdScrapeConfigsYaml")
	}
	return scrapeConfigs, nil
}

// imported from default linkerd install
var linkerdScrapeConfigsYamlTemplate = template.Must(template.New("linkerd-scrape-configs").Parse(`
- job_name: 'linkerd'
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names: ['{{.Namespace}}']

  relabel_configs:
  - source_labels:
    - __meta_kubernetes_pod_container_name
    action: keep
    regex: ^prometheus$

  honor_labels: true
  metrics_path: '/federate'

  params:
    'match[]':
      - '{job="linkerd-proxy"}'
      - '{job="linkerd-controller"}'
`))
