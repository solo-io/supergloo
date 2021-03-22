
---
title: "Enterprise Networking"
description: Reference for Helm values. 
weight: 2
---

|Option|Type|Default Value|Description|
|------|----|-----------|-------------|
|glooMeshOperatorArgs|struct| ||
|glooMeshOperatorArgs.settingsRef|struct| ||
|glooMeshOperatorArgs.settingsRef.name|string|settings||
|glooMeshOperatorArgs.settingsRef.namespace|string|gloo-mesh||
|settings|struct| ||
|settings.mtls|struct| ||
|settings.mtls.istio|struct| ||
|settings.mtls.istio.tls_mode|int32|2||
|settings.networking_extension_servers[]|struct| ||
|settings.networking_extension_servers[].address|string| ||
|settings.networking_extension_servers[].insecure|bool| ||
|settings.networking_extension_servers[].reconnect_on_network_failures|bool| ||
|settings.discovery|struct| ||
|settings.discovery.istio|struct| ||
|settings.discovery.istio.ingress_gateway_detectors.NAME|struct| ||
|settings.discovery.istio.ingress_gateway_detectors.NAME.gateway_workload_labels.NAME|string| ||
|settings.discovery.istio.ingress_gateway_detectors.NAME.gateway_tls_port_name|string| ||
|settings.relay|struct| ||
|settings.relay.enabled|bool|false||
|settings.relay.server|struct| ||
|settings.relay.server.address|string|||
|settings.relay.server.insecure|bool|false||
|settings.relay.server.reconnect_on_network_failures|bool|false||
|disallowIntersectingConfig|bool|false||
|watchOutputTypes|bool|true||
|cluster|string||the cluster in which the management plane will deployed, if it is also a managed cluster|
|relayTlsSecret|struct| |Reference to a Secret containing TLS Certificates used to secure the Networking gRPC Server with TLS.|
|relayTlsSecret.name|string|relay-server-tls-secret||
|relayTlsSecret.namespace|string|||
|signingTlsSecret|struct| |Reference to a Secret containing TLS Certificates used to sign CSRs created by Relay Agents.|
|signingTlsSecret.name|string|relay-tls-signing-secret||
|signingTlsSecret.namespace|string|||
|tokenSecret|struct| |Reference to a Secret containing a shared Token for authenticating Relay Agents.|
|tokenSecret.name|string|relay-identity-token-secret|Name of the Kubernetes Secret|
|tokenSecret.namespace|string||Namespace of the Kubernetes Secret|
|tokenSecret.key|string|token|Key value of the data within the Kubernetes Secret|
|maxGrpcMessageSize|string|4294967295|Specify to set a custom maximum message size for grpc messages sent and received by the Relay server|
|metricsBackend|struct| |Specify a metrics backend for persisting and querying aggregated metrics|
|metricsBackend.prometheus|struct| |Specify settings for using Prometheus as the metrics storage backend.|
|metricsBackend.prometheus.enabled|bool|false|If true, use Prometheus as the metrics storage backend.|
|metricsBackend.prometheus.url|string|http://prometheus-server|Specify the URL of the Prometheus server.|
|Prometheus|struct| |Helm values for configuring Prometheus. See the [Prometheus Helm chart](https://github.com/prometheus-community/helm-charts/blob/main/charts/prometheus/values.yaml) for the complete set of values.|
|selfSigned|bool|true|Provision self signed certificates and bootstrap token for the relay server.|
