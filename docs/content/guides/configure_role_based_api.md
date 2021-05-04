---
title: Configuring the Role-based API
menuTitle: Configure Role-based API
weight: 90
---

{{< notice note >}}
This feature is available in Gloo Mesh Enterprise only. If you are using the open source version of Gloo Mesh, this guide will not work.
{{< /notice >}}

Gloo Mesh's role-based API allows organizations to restrict access to policy configuration (i.e. creation, updating, and deletion of policy configuration objects)
based on the roles of individual users, represented by a `Role` CRD. Similar to the Kubernetes RBAC model, Gloo Mesh users are bound to one or more roles. A user may create, update, or delete a networking policy if they are bound to at least one role that permits access for that policy.

When you install Gloo Mesh Enterprise with default settings, the role-based API is disabled by default. If you enable it by setting the helm value `rbac-webhook.enabled=true`, this creates an implicit **deny** on all networking policy actions, and requires that Roles are created and bound to subjects to grant them the permissions they need to perform their job duties.

In this guide, we will walk through some deployment options when it comes to configuring the role-based API. We are assuming that you already have installed Gloo Mesh Enterprise or have a Kubernetes cluster on which to install it.

## Checking the role-based API

The role-based API makes use of a `rbac-webhook` deployment and a ValidatingWebhookConfiguration. You can view the components by running the following on a cluster that has Gloo Edge Enterprise installed.

```shell
kubectl get deployment rbac-webhook -n gloo-mesh -oyaml
kubectl get ValidatingWebhookConfiguration rbac-webhook -oyaml
```

When a CRUD operation is attempted against the API group `networking.mesh.gloo.solo.io`, the web hook is triggered and passed to the `rbac-webhook` service for validation. The service makes a determination of whether the action should be allowed.

## Enable role-based API

You can choose to install the role-based API components when installing Gloo Mesh Enterprise. You can do this by installing with either `meshctl` or using Helm. Select the tab below to pick the installation method you prefer.

{{< tabs >}}
{{< tab name="meshctl install" codelang="shell" >}}
meshctl install enterprise --include-rbac --license LICENSE_KEY_STRING
{{< /tab >}}
{{< tab name="helm install" codelang="shell">}}
kubectl create ns gloo-mesh
helm install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise -n gloo-mesh \
  --set rbac-webhook.enabled=true \
  --set license.key=LICENSE_KEY_STRING
{{< /tab >}}
{{< /tabs >}}

You could also update your existing installation to add the role-based API components or remove them. Another option is to install the role-based API components, but set them to permissive mode.

## Enable permissive mode

By default, the role-based API will deny any requests that are not explicitly allowed through a role and role binding. Rather than removing the role-based API, you can instead choose to run it in permissive mode, which will implicitly allow all requests. Even though the requests will be allowed, the `rbac-webhook` container will still evaluate all requests and log whether they would be allowed or not. This is perfect for testing your roles before enforcing them!

Permissive mode is enabled by setting the environment variable `RBAC_PERMISSIVE_MODE` to true for `rbac-webhook` pod. You can do this through a Helm chart installation or update. The updated value.yaml file you will need is below:

```yaml
licenseKey: LICENSE_KEY_STRING
rbac-webhook:
  rbacWebhook:
    env:
      - name: POD_NAMESPACE
        valueFrom:
          fieldRef:
            fieldPath: metadata.namespace
      - name: SERVICE_NAME
        value: rbac-webhook
      - name: SECRET_NAME
        value: rbac-webhook
      - name: VALIDATING_WEBHOOK_CONFIGURATION_NAME
        value: rbac-webhook
      - name: CERT_DIR
        value: /etc/certs/admission
      - name: WEBHOOK_PATH
        value: /admission
      - name: RBAC_PERMISSIVE_MODE
        value: "true"
      - name: LOG_LEVEL
        value: debug
      - name: LICENSE_KEY
        valueFrom:
          secretKeyRef:
            key: key
            name: gloo-mesh-enterprise-license
```

We are setting `RBAC_PERMISSIVE_MODE` to `true` and the `LOG_LEVEL` to `debug` to ensure that we can capture the evaluation of policies even though requests will not be blocked. You can simply save the above yaml as vaules.yaml and then run the following command:

```shell
helm upgrade gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise -n gloo-mesh \
  --set-file values.yaml
```

When you are ready to update the configuration to an enforced mode, you can simply update the `RBAC_PERMISSIVE_MODE` to `false` and change the `LOG_LEVEL` back to `info`.

## Summary and Next Steps

With the role-based API enabled, you can start creating roles and bindings with our [Using the Role-based API]({{% versioned_link_path fromRoot="/guides/using_role_based_api" %}}).
