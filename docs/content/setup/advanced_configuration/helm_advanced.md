---
title: Helm Chart Customization
weight: 20
description: Overriding arbitrary deployment and service spec fields
---

{{% notice note %}} This feature will only be available for Gloo Mesh oss version >= v1.1.0-beta12 and for Gloo Mesh Enterprise version >= v1.1.0-beta13.{{% /notice %}}

## Motivation

Gloo Mesh’s helm chart is very customizable, but does not contain every possible kubernetes value you may want to tweak. In this document we will demonstrate a method of tweaking the helm release by passing in a helm value file.

This allows you to tailor the installation manifests to your specific needs quickly and easily.


## Overrides

The merging is done via a [helm library chart function](https://github.com/helm/charts/blob/master/incubator/common/templates/_util.tpl).

Each sub component (e.g. discovery, networking, dashboard, cert-agent) will have a `deploymentOverrides` and `serviceOverrides` field. The
yaml specified in this field will be merged with the default deployment and service spec fields.

## Examples

The following values.yaml file, passed into the Gloo Mesh community helm chart, will add a custom label to the discovery pod:

```yaml
discovery:
  deploymentOverrides:
    spec:
      template:
        metadata:
          annotations:
            test: new-annotation
```

To see the new annotation being applied, run:
```
helm template gloo-mesh https://storage.googleapis.com/gloo-mesh/gloo-mesh/gloo-mesh-$GLOO_MESH_VERSION.tgz --namespace gloo-mesh --values values.yaml
```

The following values.yaml file, passed into the Gloo Mesh Enterprise helm chart, will replace a volume mount for the discovery pod:

```yaml
discovery:
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
the service account used by the enterprise networking pod:

```yaml
enterprise-networking:
  enterpriseNetworking:
    serviceOverrides:
      spec:
        serviceAccountName: other-service-account
```

To see the new service account being used, run:
```
helm template gloo-mesh https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise/gloo-mesh-enterprise-$GLOO_MESH_VERSION.tgz --namespace gloo-mesh --values values.yaml --set licenseKey=$GLOO_MESH_LICENSE_KEY
```
