---
title: Helm Chart Customization
weight: 20
description: Overriding arbitrary deployment and service spec fields
---

## Motivation

Gloo Mesh’s helm chart is very customizable, but does not contain every possible kubernetes value you may want to tweak. In this document we will demonstrate a method of tweaking the helm release by passing in a helm value file.

This allows you to tailor the installation manifests to your specific needs quickly and easily.


## Overrides

The merging is done via a [helm library chart function](https://github.com/helm/charts/blob/master/incubator/common/templates/_util.tpl).

Each sub component (e.g. discovery, networking, dashboard, cert-agent) will have a `deploymentOverrides` and `serviceOverrides` field. The
yaml specified in this field will be merged with the default deployment and service spec fields.

## Examples

The following values.yaml file, passed into the Gloo Mesh Enterprise helm chart, will add a custom label to the dashboard pod:
```yaml
dashboard:
  deploymentOverrides:
    spec:
      replicas: 5
      template:
        metadata:
          annotations:
            test: new-annotation
```

The following values.yaml file, passed into the Gloo Mesh Enterprise helm chart, will replace a volume mount for the dashboard pod:

```yaml
dashboard:
  deploymentOverrides:
    spec:
      template:
        spec:
          volumeMounts:
            - name: envoy-config
              configMap:
                name: my-custom-envoy-config
```


The following values.yaml file, passed into the Gloo Mesh Enterprise helm chart, will replace
the service account used by the discovery pod:

```yaml
discovery:
  serviceOverrides:
    spec:
      serviceAccountName: other-discovery-service-account
```