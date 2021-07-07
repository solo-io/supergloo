
---
title: "RBAC Webhook"
description: Reference for Helm values.
weight: 2
---

|Option|Type|Default Value|Description|
|------|----|-----------|-------------|
||map[string, interface]| ||
|<MAP_KEY>|interface| ||
|adminSubjects|interface| ||
|createAdminRole|interface| ||
|licenseKey|interface| ||
|rbacWebhook|struct|{"image":{"repository":"rbac-webhook","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}},{"name":"SERVICE_NAME","value":"rbac-webhook"},{"name":"SECRET_NAME","value":"rbac-webhook"},{"name":"VALIDATING_WEBHOOK_CONFIGURATION_NAME","value":"rbac-webhook"},{"name":"CERT_DIR","value":"/etc/certs/admission"},{"name":"WEBHOOK_PATH","value":"/admission"},{"name":"RBAC_PERMISSIVE_MODE","value":"false"},{"name":"LOG_LEVEL","value":"info"},{"name":"LICENSE_KEY","valueFrom":{"secretKeyRef":{"name":"gloo-mesh-enterprise-license","key":"key"}}}],"resources":{"requests":{"cpu":"125m","memory":"256Mi"}},"sidecars":{},"serviceType":"ClusterIP","ports":{"webhook":8443}}|Configuration for the rbacWebhook deployment.|
|rbacWebhook|struct|{"image":{"repository":"rbac-webhook","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"env":[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}},{"name":"SERVICE_NAME","value":"rbac-webhook"},{"name":"SECRET_NAME","value":"rbac-webhook"},{"name":"VALIDATING_WEBHOOK_CONFIGURATION_NAME","value":"rbac-webhook"},{"name":"CERT_DIR","value":"/etc/certs/admission"},{"name":"WEBHOOK_PATH","value":"/admission"},{"name":"RBAC_PERMISSIVE_MODE","value":"false"},{"name":"LOG_LEVEL","value":"info"},{"name":"LICENSE_KEY","valueFrom":{"secretKeyRef":{"name":"gloo-mesh-enterprise-license","key":"key"}}}],"resources":{"requests":{"cpu":"125m","memory":"256Mi"}}}||
|rbacWebhook.image|struct|{"repository":"rbac-webhook","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"}|Specify the container image|
|rbacWebhook.image.tag|string| |Tag for the container.|
|rbacWebhook.image.repository|string|rbac-webhook|Image name (repository).|
|rbacWebhook.image.registry|string|gcr.io/gloo-mesh|Image registry.|
|rbacWebhook.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|rbacWebhook.image.pullSecret|string| |Image pull secret.|
|rbacWebhook.Env[]|slice|[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}},{"name":"SERVICE_NAME","value":"rbac-webhook"},{"name":"SECRET_NAME","value":"rbac-webhook"},{"name":"VALIDATING_WEBHOOK_CONFIGURATION_NAME","value":"rbac-webhook"},{"name":"CERT_DIR","value":"/etc/certs/admission"},{"name":"WEBHOOK_PATH","value":"/admission"},{"name":"RBAC_PERMISSIVE_MODE","value":"false"},{"name":"LOG_LEVEL","value":"info"},{"name":"LICENSE_KEY","valueFrom":{"secretKeyRef":{"name":"gloo-mesh-enterprise-license","key":"key"}}}]|Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|rbacWebhook.resources|struct|{"requests":{"cpu":"125m","memory":"256Mi"}}|Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|rbacWebhook.resources.limits|map[string, struct]| ||
|rbacWebhook.resources.limits.<MAP_KEY>|struct| ||
|rbacWebhook.resources.limits.<MAP_KEY>|string| ||
|rbacWebhook.resources.requests|map[string, struct]| ||
|rbacWebhook.resources.requests.<MAP_KEY>|struct| ||
|rbacWebhook.resources.requests.<MAP_KEY>|string| ||
|rbacWebhook.resources.requests.cpu|struct|"125m"||
|rbacWebhook.resources.requests.cpu|string|DecimalSI||
|rbacWebhook.resources.requests.memory|struct|"256Mi"||
|rbacWebhook.resources.requests.memory|string|BinarySI||
|rbacWebhook.sidecars|map[string, struct]| |Configuration for the deployed containers.|
|rbacWebhook.sidecars.<MAP_KEY>|struct| |Configuration for the deployed containers.|
|rbacWebhook.sidecars.<MAP_KEY>.image|struct| |Specify the container image|
|rbacWebhook.sidecars.<MAP_KEY>.image.tag|string| |Tag for the container.|
|rbacWebhook.sidecars.<MAP_KEY>.image.repository|string| |Image name (repository).|
|rbacWebhook.sidecars.<MAP_KEY>.image.registry|string| |Image registry.|
|rbacWebhook.sidecars.<MAP_KEY>.image.pullPolicy|string| |Image pull policy.|
|rbacWebhook.sidecars.<MAP_KEY>.image.pullSecret|string| |Image pull secret.|
|rbacWebhook.sidecars.<MAP_KEY>.Env[]|slice| |Specify environment variables for the container. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
|rbacWebhook.sidecars.<MAP_KEY>.resources|struct| |Specify container resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|rbacWebhook.sidecars.<MAP_KEY>.resources.limits|map[string, struct]| ||
|rbacWebhook.sidecars.<MAP_KEY>.resources.limits.<MAP_KEY>|struct| ||
|rbacWebhook.sidecars.<MAP_KEY>.resources.limits.<MAP_KEY>|string| ||
|rbacWebhook.sidecars.<MAP_KEY>.resources.requests|map[string, struct]| ||
|rbacWebhook.sidecars.<MAP_KEY>.resources.requests.<MAP_KEY>|struct| ||
|rbacWebhook.sidecars.<MAP_KEY>.resources.requests.<MAP_KEY>|string| ||
|rbacWebhook.serviceType|string|ClusterIP|Specify the service type. Can be either "ClusterIP", "NodePort", "LoadBalancer", or "ExternalName".|
|rbacWebhook.ports|map[string, uint32]| |Specify service ports as a map from port name to port number.|
|rbacWebhook.ports.<MAP_KEY>|uint32| |Specify service ports as a map from port name to port number.|
|rbacWebhook.ports.webhook|uint32|8443|Specify service ports as a map from port name to port number.|
|rbacWebhook.DeploymentOverrides|invalid| |Provide arbitrary overrides for the component's [deployment template](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/deployment-v1/)|
|rbacWebhook.ServiceOverrides|invalid| |Provide arbitrary overrides for the component's [service template](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/).|
