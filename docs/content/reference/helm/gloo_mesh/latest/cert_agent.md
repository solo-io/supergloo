
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
|certAgent|struct|{"image":{"repository":"cert-agent","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"resources":{"requests":{"cpu":"50m","memory":"128Mi"}},"serviceType":"ClusterIP","ports":{"metrics":9091},"env":[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}}],"customPodAnnotations":{"sidecar.istio.io/inject":"\"false\""}}|Configuration for the certAgent deployment.|
|certAgent.image|struct|{"repository":"cert-agent","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"}|Specify the deployment image.|
|certAgent.image.tag|string| |Tag for the container.|
|certAgent.image.repository|string|cert-agent|Image name (repository).|
|certAgent.image.registry|string|gcr.io/gloo-mesh|Image registry.|
|certAgent.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|certAgent.image.pullSecret|string| |Image pull policy. |
|certAgent.Resources|struct|{"requests":{"cpu":"50m","memory":"128Mi"}}|Specify deployment resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|certAgent.serviceType|string|ClusterIP|Specify the service type. Can be either "ClusterIP", "NodePort", "LoadBalancer", or "ExternalName".|
|certAgent.ports|map[string, uint32]| |Specify service ports as a map from port name to port number.|
|certAgent.ports.<MAP_KEY>|uint32| |Specify service ports as a map from port name to port number.|
|certAgent.ports.metrics|uint32|9091|Specify service ports as a map from port name to port number.|
|certAgent.Env[]|slice|[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}}]|Specify environment variables for the deployment. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|certAgent.customPodLabels|map[string, string]| |Custom labels for the pod|
|certAgent.customPodLabels.<MAP_KEY>|string| |Custom labels for the pod|
|certAgent.customPodAnnotations|map[string, string]| |Custom annotations for the pod|
|certAgent.customPodAnnotations.<MAP_KEY>|string| |Custom annotations for the pod|
|certAgent.customPodAnnotations.sidecar.istio.io/inject|string|"false"|Custom annotations for the pod|
|certAgent.customDeploymentLabels|map[string, string]| |Custom labels for the deployment|
|certAgent.customDeploymentLabels.<MAP_KEY>|string| |Custom labels for the deployment|
|certAgent.customDeploymentAnnotations|map[string, string]| |Custom annotations for the deployment|
|certAgent.customDeploymentAnnotations.<MAP_KEY>|string| |Custom annotations for the deployment|
|certAgent.customServiceLabels|map[string, string]| |Custom labels for the service|
|certAgent.customServiceLabels.<MAP_KEY>|string| |Custom labels for the service|
|certAgent.customServiceAnnotations|map[string, string]| |Custom annotations for the service|
|certAgent.customServiceAnnotations.<MAP_KEY>|string| |Custom annotations for the service|
