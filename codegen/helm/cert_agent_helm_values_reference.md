
---
title: "Cert Agent"
description: Reference for Helm values.
weight: 2
---

|Option|Type|Default Value|Description|
|------|----|-----------|-------------|
|glooMeshOperatorArgs|struct|{"settingsRef":{"name":"","namespace":""}}|Command line argument to Gloo Mesh deployments.|
|glooMeshOperatorArgs.settingsRef|struct|{"name":"","namespace":""}|Name/namespace of the Settings object.|
|glooMeshOperatorArgs.settingsRef.name|string| |Name of the Settings object.|
|glooMeshOperatorArgs.settingsRef.namespace|string| |Namespace of the Settings object.|
|settings|struct|{"mtls":null,"networkingExtensionServers":[],"discovery":null,"relay":null}|Values for the Settings object. See the [Settings API doc](../../../../api/github.com.solo-io.gloo-mesh.api.settings.v1.settings) for details.|
|settings.mtls|struct| ||
|settings.mtls.istio|struct| ||
|settings.mtls.istio.tls_mode|int32| ||
|settings.networking_extension_servers[]|[]ptr|null||
|settings.networking_extension_servers[]|struct| ||
|settings.networking_extension_servers[].address|string| ||
|settings.networking_extension_servers[].insecure|bool| ||
|settings.networking_extension_servers[].reconnect_on_network_failures|bool| ||
|settings.discovery|struct| ||
|settings.discovery.istio|struct| ||
|settings.discovery.istio.ingress_gateway_detectors|map[string, struct]| ||
|settings.discovery.istio.ingress_gateway_detectors.<MAP_KEY>|struct| ||
|settings.discovery.istio.ingress_gateway_detectors.<MAP_KEY>.gateway_workload_labels|map[string, string]| ||
|settings.discovery.istio.ingress_gateway_detectors.<MAP_KEY>.gateway_workload_labels.<MAP_KEY>|string| ||
|settings.discovery.istio.ingress_gateway_detectors.<MAP_KEY>.gateway_tls_port_name|string| ||
|settings.relay|struct| ||
|settings.relay.enabled|bool| ||
|settings.relay.server|struct| ||
|settings.relay.server.address|string| ||
|settings.relay.server.insecure|bool| ||
|settings.relay.server.reconnect_on_network_failures|bool| ||
|disallowIntersectingConfig|bool|false|If true, Gloo Mesh will detect and report errors when outputting service mesh configuration that overlaps with existing config not managed by Gloo Mesh.|
|watchOutputTypes|bool|false|If true, Gloo Mesh will watch service mesh config types output by Gloo Mesh, and resync upon changes.|
|defaultMetricsPort|uint32|0|The port on which to serve internal Prometheus metrics for the Gloo Mesh application. Set to 0 to disable.|
|verbose|bool|false|If true, enables verbose/debug logging.|
|certAgent|struct|{"image":{"repository":"cert-agent","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}}],"resources":{"requests":{"cpu":"50m","memory":"128Mi"}},"sidecars":{},"serviceType":"ClusterIP","ports":{"metrics":9091}}|Configuration for the certAgent deployment.|
|certAgent|struct|{"image":{"repository":"cert-agent","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}}],"resources":{"requests":{"cpu":"50m","memory":"128Mi"}}}||
|certAgent.image|struct|{"repository":"cert-agent","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"}|Specify the container image|
|certAgent.image.tag|string| |Tag for the container.|
|certAgent.image.repository|string|cert-agent|Image name (repository).|
|certAgent.image.registry|string|gcr.io/gloo-mesh|Image registry.|
|certAgent.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|certAgent.image.pullSecret|string| |Image pull secret.|
|certAgent.Env[]|slice|[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}}]|Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|certAgent.resources|struct|{"requests":{"cpu":"50m","memory":"128Mi"}}|Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|certAgent.resources.limits|map[string, struct]| ||
|certAgent.resources.limits.<MAP_KEY>|struct| ||
|certAgent.resources.limits.<MAP_KEY>|string| ||
|certAgent.resources.requests|map[string, struct]| ||
|certAgent.resources.requests.<MAP_KEY>|struct| ||
|certAgent.resources.requests.<MAP_KEY>|string| ||
|certAgent.resources.requests.cpu|struct|"50m"||
|certAgent.resources.requests.cpu|string|DecimalSI||
|certAgent.resources.requests.memory|struct|"128Mi"||
|certAgent.resources.requests.memory|string|BinarySI||
|certAgent.sidecars|map[string, struct]| |Configuration for the deployed containers.|
|certAgent.sidecars.<MAP_KEY>|struct| |Configuration for the deployed containers.|
|certAgent.sidecars.<MAP_KEY>.image|struct| |Specify the container image|
|certAgent.sidecars.<MAP_KEY>.image.tag|string| |Tag for the container.|
|certAgent.sidecars.<MAP_KEY>.image.repository|string| |Image name (repository).|
|certAgent.sidecars.<MAP_KEY>.image.registry|string| |Image registry.|
|certAgent.sidecars.<MAP_KEY>.image.pullPolicy|string| |Image pull policy.|
|certAgent.sidecars.<MAP_KEY>.image.pullSecret|string| |Image pull secret.|
|certAgent.sidecars.<MAP_KEY>.Env[]|slice| |Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|certAgent.sidecars.<MAP_KEY>.resources|struct| |Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|certAgent.sidecars.<MAP_KEY>.resources.limits|map[string, struct]| ||
|certAgent.sidecars.<MAP_KEY>.resources.limits.<MAP_KEY>|struct| ||
|certAgent.sidecars.<MAP_KEY>.resources.limits.<MAP_KEY>|string| ||
|certAgent.sidecars.<MAP_KEY>.resources.requests|map[string, struct]| ||
|certAgent.sidecars.<MAP_KEY>.resources.requests.<MAP_KEY>|struct| ||
|certAgent.sidecars.<MAP_KEY>.resources.requests.<MAP_KEY>|string| ||
|certAgent.serviceType|string|ClusterIP|Specify the service type. Can be either "ClusterIP", "NodePort", "LoadBalancer", or "ExternalName".|
|certAgent.ports|map[string, uint32]| |Specify service ports as a map from port name to port number.|
|certAgent.ports.<MAP_KEY>|uint32| |Specify service ports as a map from port name to port number.|
|certAgent.ports.metrics|uint32|9091|Specify service ports as a map from port name to port number.|
|certAgent.DeploymentOverrides|invalid| |Provide arbitrary overrides for the component's [deployment template](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/deployment-v1/)|
|certAgent.ServiceOverrides|invalid| |Provide arbitrary overrides for the component's [service template](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/).|
