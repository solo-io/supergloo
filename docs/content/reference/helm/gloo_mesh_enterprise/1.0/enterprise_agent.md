
---
title: "Enterprise Agent"
description: Reference for Helm values.
weight: 2
---

|Option|Type|Default Value|Description|
|------|----|-----------|-------------|
|defaultMetricsPort|uint32|9091|The port on which to serve internal Prometheus metrics for the Gloo Mesh application. Set to 0 to disable.|
|relay|struct|{"cluster":"","serverAddress":"","authority":"enterprise-networking.gloo-mesh","insecure":false,"clientCertSecret":{"name":"relay-client-tls-secret"},"rootTlsSecret":{"name":"relay-root-tls-secret"},"tokenSecret":{"name":"relay-identity-token-secret","namespace":"","key":"token"},"maxGrpcMessageSize":"4294967295"}|options for connecting relay|
|relay.cluster|string| |cluster identifier for the relay agent|
|relay.serverAddress|string| |address of the relay server|
|relay.authority|string|enterprise-networking.gloo-mesh|set the authority/host header to this value when dialing the Relay gRPC Server|
|relay.insecure|bool|false|communicate with relay server over plain HTTP|
|relay.clientCertSecret|struct|{"name":"relay-client-tls-secret"}|Reference to a Secret containing the Client TLS Certificates used to identify the Relay Agent to the Server. If the secret does not exist, a Token and Root cert secret are required.|
|relay.clientCertSecret.name|string|relay-client-tls-secret||
|relay.clientCertSecret.namespace|string| ||
|relay.rootTlsSecret|struct|{"name":"relay-root-tls-secret"}|Reference to a Secret containing a Root TLS Certificates used to verify the Relay Server Certificate. The secret can also optionally specify a 'tls.key' which will be used to generate the Agent Client Certificate.|
|relay.rootTlsSecret.name|string|relay-root-tls-secret||
|relay.rootTlsSecret.namespace|string| ||
|relay.tokenSecret|struct|{"name":"relay-identity-token-secret","namespace":"","key":"token"}|Reference to a Secret containing a shared Token for authenticating to the Relay Server|
|relay.tokenSecret.name|string|relay-identity-token-secret|Name of the Kubernetes Secret|
|relay.tokenSecret.namespace|string| |Namespace of the Kubernetes Secret|
|relay.tokenSecret.key|string|token|Key value of the data within the Kubernetes Secret|
|relay.maxGrpcMessageSize|string|4294967295|Specify to set a custom maximum message size for grpc messages sent to the Relay server|
|settingsRef|struct|{"name":"settings","namespace":"gloo-mesh"}|ref to the settings object that will be received from the networking server.|
|settingsRef.name|string|settings||
|settingsRef.namespace|string|gloo-mesh||
|verbose|bool|false||
|enterpriseAgent|struct|{"image":{"repository":"enterprise-agent","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"resources":{"requests":{"cpu":"50m","memory":"128Mi"}},"serviceType":"ClusterIP","ports":{"grpc":9977,"http":9988},"env":[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}}],"customPodAnnotations":{"sidecar.istio.io/inject":"\"false\""}}|Configuration for the enterpriseAgent deployment.|
|enterpriseAgent.image|struct|{"repository":"enterprise-agent","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"}|Specify the deployment image.|
|enterpriseAgent.image.tag|string| |Tag for the container.|
|enterpriseAgent.image.repository|string|enterprise-agent|Image name (repository).|
|enterpriseAgent.image.registry|string|gcr.io/gloo-mesh|Image registry.|
|enterpriseAgent.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|enterpriseAgent.image.pullSecret|string| |Image pull policy. |
|enterpriseAgent.Resources|struct|{"requests":{"cpu":"50m","memory":"128Mi"}}|Specify deployment resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|enterpriseAgent.serviceType|string|ClusterIP|Specify the service type. Can be either "ClusterIP", "NodePort", "LoadBalancer", or "ExternalName".|
|enterpriseAgent.ports|map[string, uint32]| |Specify service ports as a map from port name to port number.|
|enterpriseAgent.ports.<MAP_KEY>|uint32| |Specify service ports as a map from port name to port number.|
|enterpriseAgent.ports.grpc|uint32|9977|Specify service ports as a map from port name to port number.|
|enterpriseAgent.ports.http|uint32|9988|Specify service ports as a map from port name to port number.|
|enterpriseAgent.Env[]|slice|[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}}]|Specify environment variables for the deployment. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|enterpriseAgent.customPodLabels|map[string, string]| |Custom labels for the pod|
|enterpriseAgent.customPodLabels.<MAP_KEY>|string| |Custom labels for the pod|
|enterpriseAgent.customPodAnnotations|map[string, string]| |Custom annotations for the pod|
|enterpriseAgent.customPodAnnotations.<MAP_KEY>|string| |Custom annotations for the pod|
|enterpriseAgent.customPodAnnotations.sidecar.istio.io/inject|string|"false"|Custom annotations for the pod|
|enterpriseAgent.customDeploymentLabels|map[string, string]| |Custom labels for the deployment|
|enterpriseAgent.customDeploymentLabels.<MAP_KEY>|string| |Custom labels for the deployment|
|enterpriseAgent.customDeploymentAnnotations|map[string, string]| |Custom annotations for the deployment|
|enterpriseAgent.customDeploymentAnnotations.<MAP_KEY>|string| |Custom annotations for the deployment|
|enterpriseAgent.customServiceLabels|map[string, string]| |Custom labels for the service|
|enterpriseAgent.customServiceLabels.<MAP_KEY>|string| |Custom labels for the service|
|enterpriseAgent.customServiceAnnotations|map[string, string]| |Custom annotations for the service|
|enterpriseAgent.customServiceAnnotations.<MAP_KEY>|string| |Custom annotations for the service|
