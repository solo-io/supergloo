package istio

// Tested working for istio 1.0.3, 1.0.5 and 1.0.6
// be sure to test if adding new versions of istio!
const helmValues = `
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

security:
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
# security configuration
#
security:
  replicaCount: 1
  image: citadel
  selfSigned: {{ .Mtls.SelfSignedCert }} # indicate if self-signed CA is used.

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
`
