
---
title: "Gloo Mesh"
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
|discovery|struct|{"image":{"repository":"gloo-mesh","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"resources":{"requests":{"cpu":"125m","memory":"256Mi"}},"serviceType":"ClusterIP","ports":{"metrics":9091},"env":[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}}],"customPodAnnotations":{"sidecar.istio.io/inject":"\"false\""}}|Configuration for the discovery deployment.|
|discovery.image|struct|{"repository":"gloo-mesh","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"}|Specify the deployment image.|
|discovery.image.tag|string| |Tag for the container.|
|discovery.image.repository|string|gloo-mesh|Image name (repository).|
|discovery.image.registry|string|gcr.io/gloo-mesh|Image registry.|
|discovery.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|discovery.image.pullSecret|string| |Image pull policy. |
|discovery.Resources|struct|{"requests":{"cpu":"125m","memory":"256Mi"}}|Specify deployment resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|discovery.serviceType|string|ClusterIP|Specify the service type. Can be either "ClusterIP", "NodePort", "LoadBalancer", or "ExternalName".|
|discovery.ports|map[string, uint32]| |Specify service ports as a map from port name to port number.|
|discovery.ports.<MAP_KEY>|uint32| |Specify service ports as a map from port name to port number.|
|discovery.ports.metrics|uint32|9091|Specify service ports as a map from port name to port number.|
|discovery.Env[]|slice|[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}}]|Specify environment variables for the deployment. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|discovery.customPodLabels|map[string, string]| |Custom labels for the pod|
|discovery.customPodLabels.<MAP_KEY>|string| |Custom labels for the pod|
|discovery.customPodAnnotations|map[string, string]| |Custom annotations for the pod|
|discovery.customPodAnnotations.<MAP_KEY>|string| |Custom annotations for the pod|
|discovery.customPodAnnotations.sidecar.istio.io/inject|string|"false"|Custom annotations for the pod|
|discovery.customDeploymentLabels|map[string, string]| |Custom labels for the deployment|
|discovery.customDeploymentLabels.<MAP_KEY>|string| |Custom labels for the deployment|
|discovery.customDeploymentAnnotations|map[string, string]| |Custom annotations for the deployment|
|discovery.customDeploymentAnnotations.<MAP_KEY>|string| |Custom annotations for the deployment|
|discovery.customServiceLabels|map[string, string]| |Custom labels for the service|
|discovery.customServiceLabels.<MAP_KEY>|string| |Custom labels for the service|
|discovery.customServiceAnnotations|map[string, string]| |Custom annotations for the service|
|discovery.customServiceAnnotations.<MAP_KEY>|string| |Custom annotations for the service|
|networking|struct|{"image":{"repository":"gloo-mesh","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"resources":{"requests":{"cpu":"125m","memory":"256Mi"}},"serviceType":"","ports":{},"env":[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}}],"customPodAnnotations":{"sidecar.istio.io/inject":"\"false\""}}|Configuration for the networking deployment.|
|networking.image|struct|{"repository":"gloo-mesh","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"}|Specify the deployment image.|
|networking.image.tag|string| |Tag for the container.|
|networking.image.repository|string|gloo-mesh|Image name (repository).|
|networking.image.registry|string|gcr.io/gloo-mesh|Image registry.|
|networking.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|networking.image.pullSecret|string| |Image pull policy. |
|networking.Resources|struct|{"requests":{"cpu":"125m","memory":"256Mi"}}|Specify deployment resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|networking.serviceType|string| |Specify the service type. Can be either "ClusterIP", "NodePort", "LoadBalancer", or "ExternalName".|
|networking.ports|map[string, uint32]| |Specify service ports as a map from port name to port number.|
|networking.ports.<MAP_KEY>|uint32| |Specify service ports as a map from port name to port number.|
|networking.Env[]|slice|[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}}]|Specify environment variables for the deployment. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|networking.customPodLabels|map[string, string]| |Custom labels for the pod|
|networking.customPodLabels.<MAP_KEY>|string| |Custom labels for the pod|
|networking.customPodAnnotations|map[string, string]| |Custom annotations for the pod|
|networking.customPodAnnotations.<MAP_KEY>|string| |Custom annotations for the pod|
|networking.customPodAnnotations.sidecar.istio.io/inject|string|"false"|Custom annotations for the pod|
|networking.customDeploymentLabels|map[string, string]| |Custom labels for the deployment|
|networking.customDeploymentLabels.<MAP_KEY>|string| |Custom labels for the deployment|
|networking.customDeploymentAnnotations|map[string, string]| |Custom annotations for the deployment|
|networking.customDeploymentAnnotations.<MAP_KEY>|string| |Custom annotations for the deployment|
|networking.customServiceLabels|map[string, string]| |Custom labels for the service|
|networking.customServiceLabels.<MAP_KEY>|string| |Custom labels for the service|
|networking.customServiceAnnotations|map[string, string]| |Custom annotations for the service|
|networking.customServiceAnnotations.<MAP_KEY>|string| |Custom annotations for the service|
