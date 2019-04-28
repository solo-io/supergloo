package istio

import (
	"bytes"
	"text/template"

	"github.com/solo-io/go-utils/installutils/helmchart"
)

var supportedIstioVersions = map[string]versionedInstall{
	IstioVersion103: {
		chartPath:      IstioVersion103Chart,
		valuesTemplate: helmValuesTemplate,
	},
	IstioVersion105: {
		chartPath:      IstioVersion105Chart,
		valuesTemplate: helmValuesTemplate,
	},
	IstioVersion106: {
		chartPath:      IstioVersion106Chart,
		valuesTemplate: helmValuesTemplate,
	},
	IstioVersion113: {
		extraManifests: istio113ExtraManifests,
		chartPath:      IstioVersion113Chart,
		valuesTemplate: helmValuesTemplate,
	},
}

type versionedInstall struct {
	extraManifests helmchart.Manifests
	chartPath      string
	valuesTemplate *template.Template
}

type helmChartParams struct {
	valuesTemplate *template.Template
	// these fields are used to render the values template
	AutoInject    autoInjectInstallOptions
	Mtls          mtlsInstallOptions
	Observability observabilityInstallOptions
	Gateway       gatewayInstallOptions
}

func (p helmChartParams) helmValues() (string, error) {
	buf := &bytes.Buffer{}
	if err := p.valuesTemplate.Execute(buf, p); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type autoInjectInstallOptions struct {
	Enabled bool
}

type mtlsInstallOptions struct {
	Enabled        bool
	SelfSignedCert bool
}

type observabilityInstallOptions struct {
	EnableGrafana      bool
	EnablePrometheus   bool
	EnableJaeger       bool
	EnableServiceGraph bool
}

type gatewayInstallOptions struct {
	EnableIngress bool
	EnableEgress  bool
}

// Tested working for istio 1.0.3, 1.0.5 and 1.0.6
// be sure to test if adding new versions of istio!
var helmValuesTemplate = template.Must(template.New("istio-1.0.x-helmvalues").Parse(`
global:
  proxy:
    {{- if .AutoInject.Enabled }}
    autoInject: enabled
    {{- else }}
    autoInject: disabled
    {{- end }}

  # controlPlaneMtls enabled.
  controlPlaneSecurityEnabled: {{ .Mtls.Enabled }}

  # Default mtls policy. If true, mtls between services will be enabled by default.
  mtls:
    # Default setting for service-to-service mtls. Can be set explicitly using
    # destination rules or service annotations.
    enabled: {{ .Mtls.Enabled }}

#
# security configuration
#
security:
  replicaCount: 1
  image: citadel
  selfSigned: {{ .Mtls.SelfSignedCert }} # indicate if self-signed CA is used.
  enabled: true

#
# ingress configuration
#
ingress:
  enabled: {{ .Gateway.EnableIngress }}

#
# Gateways Configuration
# By default (if enabled) a pair of Ingress and Egress Gateways will be created for the mesh.
# You can add more gateways in addition to the defaults but make sure those are uniquely named
# and that NodePorts are not conflicting.
# Disable specifc gateway by setting the "enabled" to false.
#
gateways:
  enabled: {{ or .Gateway.EnableIngress .Gateway.EnableEgress }}

  istio-ingressgateway:
    enabled: {{ .Gateway.EnableIngress }}

  istio-egressgateway:
    enabled: {{ .Gateway.EnableEgress }}

#
# sidecar-injector webhook configuration
#
sidecarInjectorWebhook:
  enabled: {{ .AutoInject.Enabled }}

#
# addons configuration
#
telemetry-gateway:
  grafanaEnabled: false
  prometheusEnabled: false

grafana:
  enabled: {{ .Observability.EnableGrafana }}

prometheus:
  enabled: {{ .Observability.EnablePrometheus }}

servicegraph:
  enabled: {{ and .Observability.EnableServiceGraph .Observability.EnablePrometheus }}

tracing:
  enabled: {{ .Observability.EnableJaeger }}
`))
