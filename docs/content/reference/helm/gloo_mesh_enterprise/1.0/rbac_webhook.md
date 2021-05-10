
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
|rbacWebhook|struct|{"image":{"repository":"rbac-webhook","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"},"resources":{"requests":{"cpu":"125m","memory":"256Mi"}},"serviceType":"ClusterIP","ports":{"webhook":8443},"env":[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}},{"name":"SERVICE_NAME","value":"rbac-webhook"},{"name":"SECRET_NAME","value":"rbac-webhook"},{"name":"VALIDATING_WEBHOOK_CONFIGURATION_NAME","value":"rbac-webhook"},{"name":"CERT_DIR","value":"/etc/certs/admission"},{"name":"WEBHOOK_PATH","value":"/admission"},{"name":"RBAC_PERMISSIVE_MODE","value":"false"},{"name":"LOG_LEVEL","value":"info"},{"name":"LICENSE_KEY","valueFrom":{"secretKeyRef":{"name":"gloo-mesh-enterprise-license","key":"key"}}}]}|Configuration for the rbacWebhook deployment.|
|rbacWebhook.image|struct|{"repository":"rbac-webhook","registry":"gcr.io/gloo-mesh","pullPolicy":"IfNotPresent"}|Specify the deployment image.|
|rbacWebhook.image.tag|string| |Tag for the container.|
|rbacWebhook.image.repository|string|rbac-webhook|Image name (repository).|
|rbacWebhook.image.registry|string|gcr.io/gloo-mesh|Image registry.|
|rbacWebhook.image.pullPolicy|string|IfNotPresent|Image pull policy.|
|rbacWebhook.image.pullSecret|string| |Image pull policy. |
|rbacWebhook.Resources|struct|{"requests":{"cpu":"125m","memory":"256Mi"}}|Specify deployment resource requirements. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#resourcerequirements-v1-core) for specification details.|
|rbacWebhook.serviceType|string|ClusterIP|Specify the service type. Can be either "ClusterIP", "NodePort", "LoadBalancer", or "ExternalName".|
|rbacWebhook.ports|map[string, uint32]| |Specify service ports as a map from port name to port number.|
|rbacWebhook.ports.<MAP_KEY>|uint32| |Specify service ports as a map from port name to port number.|
|rbacWebhook.ports.webhook|uint32|8443|Specify service ports as a map from port name to port number.|
|rbacWebhook.Env[]|slice|[{"name":"POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace"}}},{"name":"SERVICE_NAME","value":"rbac-webhook"},{"name":"SECRET_NAME","value":"rbac-webhook"},{"name":"VALIDATING_WEBHOOK_CONFIGURATION_NAME","value":"rbac-webhook"},{"name":"CERT_DIR","value":"/etc/certs/admission"},{"name":"WEBHOOK_PATH","value":"/admission"},{"name":"RBAC_PERMISSIVE_MODE","value":"false"},{"name":"LOG_LEVEL","value":"info"},{"name":"LICENSE_KEY","valueFrom":{"secretKeyRef":{"name":"gloo-mesh-enterprise-license","key":"key"}}}]|Specify environment variables for the deployment. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#envvarsource-v1-core) for specification details.|
