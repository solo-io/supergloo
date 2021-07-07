
---
title: "Enterprise Networking"
description: Reference for Helm values.
weight: 2
---

|Option|Type|Default Value|Description|
|------|----|-----------|-------------|
|glooMeshOperatorArgs|struct|{"settingsRef":{"name":"settings","namespace":"gloo-mesh"}}|Command line argument to Gloo Mesh deployments.|
|glooMeshOperatorArgs.settingsRef|struct|{"name":"settings","namespace":"gloo-mesh"}|Name/namespace of the Settings object.|
|glooMeshOperatorArgs.settingsRef.name|string|settings|Name of the Settings object.|
|glooMeshOperatorArgs.settingsRef.namespace|string|gloo-mesh|Namespace of the Settings object.|
|settings|struct|{"mtls":{"istio":{"tlsMode":"ISTIO_MUTUAL"}},"networkingExtensionServers":[],"discovery":{"istio":{"ingressGatewayDetectors":{}}},"relay":{"enabled":false,"server":{"address":"","insecure":false,"reconnectOnNetworkFailures":false}}}|Values for the Settings object. See the [Settings API doc](../../../../api/github.com.solo-io.gloo-mesh.api.settings.v1.settings) for details.|
|settings.mtls|struct|{"istio":{"tls_mode":2}}||
|settings.mtls.istio|struct|{"tls_mode":2}||
|settings.mtls.istio.tls_mode|int32|2||
|settings.networking_extension_servers[]|[]ptr|null||
|settings.networking_extension_servers[]|struct| ||
|settings.networking_extension_servers[].address|string| ||
|settings.networking_extension_servers[].insecure|bool| ||
|settings.networking_extension_servers[].reconnect_on_network_failures|bool| ||
|settings.discovery|struct|{"istio":{}}||
|settings.discovery.istio|struct|{}||
|settings.discovery.istio.ingress_gateway_detectors|map[string, struct]| ||
|settings.discovery.istio.ingress_gateway_detectors.<MAP_KEY>|struct| ||
|settings.discovery.istio.ingress_gateway_detectors.<MAP_KEY>.gateway_workload_labels|map[string, string]| ||
|settings.discovery.istio.ingress_gateway_detectors.<MAP_KEY>.gateway_workload_labels.<MAP_KEY>|string| ||
|settings.discovery.istio.ingress_gateway_detectors.<MAP_KEY>.gateway_tls_port_name|string| ||
|settings.relay|struct|{"server":{}}||
|settings.relay.enabled|bool|false||
|settings.relay.server|struct|{}||
|settings.relay.server.address|string| ||
|settings.relay.server.insecure|bool|false||
|settings.relay.server.reconnect_on_network_failures|bool|false||
|disallowIntersectingConfig|bool|false|If true, Gloo Mesh will detect and report errors when outputting service mesh configuration that overlaps with existing config not managed by Gloo Mesh.|
|watchOutputTypes|bool|true|If true, Gloo Mesh will watch service mesh config types output by Gloo Mesh, and resync upon changes.|
|defaultMetricsPort|uint32|9091|The port on which to serve internal Prometheus metrics for the Gloo Mesh application. Set to 0 to disable.|
|verbose|bool|false|If true, enables verbose/debug logging.|
|global|struct|{"insecure":false}|global chart values which can be set from parent charts|
|global.insecure|bool|false|Set to true to enable insecure communication between Gloo Mesh components|
|cluster|string| |the cluster in which the management plane will deployed, if it is also a managed cluster|
|relayTlsSecret|struct|{"name":"relay-server-tls-secret"}|Reference to a Secret containing TLS Certificates used to secure the Networking gRPC Server with TLS.|
|relayTlsSecret.name|string|relay-server-tls-secret||
|relayTlsSecret.namespace|string| ||
|signingTlsSecret|struct|{"name":"relay-tls-signing-secret"}|Reference to a Secret containing TLS Certificates used to sign CSRs created by Relay Agents.|
|signingTlsSecret.name|string|relay-tls-signing-secret||
|signingTlsSecret.namespace|string| ||
|tokenSecret|struct|{"name":"relay-identity-token-secret","namespace":"","key":"token"}|Reference to a Secret containing a shared Token for authenticating Relay Agents.|
|tokenSecret.name|string|relay-identity-token-secret|Name of the Kubernetes Secret|
|tokenSecret.namespace|string| |Namespace of the Kubernetes Secret|
|tokenSecret.key|string|token|Key value of the data within the Kubernetes Secret|
|forwardingTokenSecret|struct|{"name":"relay-forwarding-identity-token-secret","namespace":"","key":"token"}|Reference to a Secret containing a shared Token for authenticating with the forwarding relay server.|
|forwardingTokenSecret.name|string|relay-forwarding-identity-token-secret|Name of the Kubernetes Secret|
|forwardingTokenSecret.namespace|string| |Namespace of the Kubernetes Secret|
|forwardingTokenSecret.key|string|token|Key value of the data within the Kubernetes Secret|
|maxGrpcMessageSize|string|4294967295|Specify to set a custom maximum message size for grpc messages sent and received by the Relay server|
|metricsBackend|struct|{"prometheus":{"enabled":true,"url":"http://prometheus-server"}}|Specify a metrics backend for persisting and querying aggregated metrics|
|metricsBackend.prometheus|struct|{"enabled":true,"url":"http://prometheus-server"}|Specify settings for using Prometheus as the metrics storage backend.|
|metricsBackend.prometheus.enabled|bool|true|If true, use Prometheus as the metrics storage backend.|
|metricsBackend.prometheus.url|string|http://prometheus-server|Specify the URL of the Prometheus server.|
|Prometheus|map| |Helm values for configuring Prometheus. See the [Prometheus Helm chart](https://github.com/prometheus-community/helm-charts/blob/main/charts/prometheus/values.yaml) for the complete set of values.|
|selfSigned|bool|true|Provision self signed certificates and bootstrap token for the relay server.|
|admin|struct|{"port":{"name":"admin","port":11100}}||
|admin.port|struct|{"name":"admin","port":11100}||
|admin.port.name|string|admin|The name of this port within the service.|
|admin.port.port|int32|11100|The default port that will be exposed by this service.|
|disableRelayCa|bool|false||
|relay|struct|{"additionalSans":null,"serverCommonName":"enterprise-networking","rootCommonName":"enterprise-networking-ca"}||
|relay.additionalSans[]|[]string|null|additional SANs to add to  relay-server cert|
|relay.additionalSans[]|string| |additional SANs to add to  relay-server cert|
|relay.serverCommonName|string|enterprise-networking|CN (CommonName) to use for the relay-server cert. Default: enterprise-networking|
|relay.rootCommonName|string|enterprise-networking-ca|CN (CommonName) to use for the relay-rooot cert. Default: enterprise-networking-ca|
|enterpriseNetworking|struct|{"image":{"repository":"enterprise-networking","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}},{"name":"LICENSE_KEY","valueFrom":{"secretKeyRef":{"name":"gloo-mesh-enterprise-license","key":"key"}}}],"resources":{"requests":{"cpu":"125m","memory":"256Mi"}},"sidecars":{},"serviceType":"LoadBalancer","ports":{"grpc":9900,"http":8080}}|Configuration for the enterpriseNetworking deployment.|
|enterpriseNetworking|struct|{"image":{"repository":"enterprise-networking","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}},{"name":"LICENSE_KEY","valueFrom":{"secretKeyRef":{"name":"gloo-mesh-enterprise-license","key":"key"}}}],"resources":{"requests":{"cpu":"125m","memory":"256Mi"}}}||
|enterpriseNetworking.image|struct|{"repository":"enterprise-networking","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"}|Specify the container image|
|enterpriseNetworking.image.tag|string| |Tag for the container.|
|enterpriseNetworking.image.repository|string|enterprise-networking|Image name (repository).|
|enterpriseNetworking.image.registry|string|gcr.io/gloo-mesh|Image registry.|
|enterpriseNetworking.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|enterpriseNetworking.image.pullSecret|string| |Image pull secret.|
|enterpriseNetworking.Env[]|slice|[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}},{"name":"LICENSE_KEY","valueFrom":{"secretKeyRef":{"name":"gloo-mesh-enterprise-license","key":"key"}}}]|Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|enterpriseNetworking.resources|struct|{"requests":{"cpu":"125m","memory":"256Mi"}}|Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|enterpriseNetworking.resources.limits|map[string, struct]| ||
|enterpriseNetworking.resources.limits.<MAP_KEY>|struct| ||
|enterpriseNetworking.resources.limits.<MAP_KEY>|string| ||
|enterpriseNetworking.resources.requests|map[string, struct]| ||
|enterpriseNetworking.resources.requests.<MAP_KEY>|struct| ||
|enterpriseNetworking.resources.requests.<MAP_KEY>|string| ||
|enterpriseNetworking.resources.requests.cpu|struct|"125m"||
|enterpriseNetworking.resources.requests.cpu|string|DecimalSI||
|enterpriseNetworking.resources.requests.memory|struct|"256Mi"||
|enterpriseNetworking.resources.requests.memory|string|BinarySI||
|enterpriseNetworking.sidecars|map[string, struct]| |Configuration for the deployed containers.|
|enterpriseNetworking.sidecars.<MAP_KEY>|struct| |Configuration for the deployed containers.|
|enterpriseNetworking.sidecars.<MAP_KEY>.image|struct| |Specify the container image|
|enterpriseNetworking.sidecars.<MAP_KEY>.image.tag|string| |Tag for the container.|
|enterpriseNetworking.sidecars.<MAP_KEY>.image.repository|string| |Image name (repository).|
|enterpriseNetworking.sidecars.<MAP_KEY>.image.registry|string| |Image registry.|
|enterpriseNetworking.sidecars.<MAP_KEY>.image.pullPolicy|string| |Image pull policy.|
|enterpriseNetworking.sidecars.<MAP_KEY>.image.pullSecret|string| |Image pull secret.|
|enterpriseNetworking.sidecars.<MAP_KEY>.Env[]|slice| |Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|enterpriseNetworking.sidecars.<MAP_KEY>.resources|struct| |Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|enterpriseNetworking.sidecars.<MAP_KEY>.resources.limits|map[string, struct]| ||
|enterpriseNetworking.sidecars.<MAP_KEY>.resources.limits.<MAP_KEY>|struct| ||
|enterpriseNetworking.sidecars.<MAP_KEY>.resources.limits.<MAP_KEY>|string| ||
|enterpriseNetworking.sidecars.<MAP_KEY>.resources.requests|map[string, struct]| ||
|enterpriseNetworking.sidecars.<MAP_KEY>.resources.requests.<MAP_KEY>|struct| ||
|enterpriseNetworking.sidecars.<MAP_KEY>.resources.requests.<MAP_KEY>|string| ||
|enterpriseNetworking.serviceType|string|LoadBalancer|Specify the service type. Can be either "ClusterIP", "NodePort", "LoadBalancer", or "ExternalName".|
|enterpriseNetworking.ports|map[string, uint32]| |Specify service ports as a map from port name to port number.|
|enterpriseNetworking.ports.<MAP_KEY>|uint32| |Specify service ports as a map from port name to port number.|
|enterpriseNetworking.ports.grpc|uint32|9900|Specify service ports as a map from port name to port number.|
|enterpriseNetworking.ports.http|uint32|8080|Specify service ports as a map from port name to port number.|
|enterpriseNetworking.DeploymentOverrides|invalid| |Provide arbitrary overrides for the component's [deployment template](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/deployment-v1/)|
|enterpriseNetworking.ServiceOverrides|invalid| |Provide arbitrary overrides for the component's [service template](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/).|
