---
title: "Appmesh Guide"
menuTitle: Appmesh Guide
description: Guide for getting started using Service Mesh Hub with Appmesh.
weight: 110
---

# Service Mesh Hub with Appmesh

[App Mesh](https://aws.amazon.com/app-mesh/) provides service mesh support for compute services offered by AWS.

As part of its [multi-mesh vision](https://www.solo.io/blog/delivering-on-the-vision-of-multi-mesh/), 
Service Mesh Hub can configure AppMesh.

In this guide, we will:
* Discover AppMesh instances 

## Prerequisites

There are three pre-requisites to following these guides:

1. Install `kubectl`
    - https://kubernetes.io/docs/tasks/tools/install-kubectl/
2. Install `meshctl`
    - `curl -sL https://run.solo.io/meshctl/install | sh && export PATH=$PATH:$HOME/.service-mesh-hub/bin`
    - see the [guide]({{% versioned_link_path fromRoot="/getting_started" %}})
3. Have an existing AppMesh instance and access to a set of credentials for the associated AWS account (access key ID and secret key pair).

## Configure discovery of Appmesh

First confirm that your Appmesh instance exists by running `aws appmesh describe-mesh --mesh-name <mesh-name>`. Copy the ARN returned in the response, which you'll need in the steps below.

Upon installation of Service Mesh Hub v0.4.12+, you should see a `settings.core.zephyr.solo.io` CRD instance with the name 
`settings` in the SMH write namespace (by default this is `service-mesh-hub`), populated with

```yaml
spec:
  aws:
    disabled: true
```

By default discovery for AWS resources is disabled. To enable discovery for your Appmesh instance, replace the Settings spec with the following,
making the relevant substitutions (note for simplicity we disable EKS discovery, see the guides section for a tutorial on EKS discovery):

```yaml
apiVersion: core.zephyr.solo.io/v1alpha1
kind: Settings
metadata:
  namespace: service-mesh-hub
  name: settings
spec:
  aws:
    disabled: false
    accounts:
      - accountId: "<aws-account-id>"
        meshDiscovery:
          resourceSelectors:
          - arn: <appmesh-instance-arn>
        eksDiscovery:
          disabled: true
```

This configuration instructs SMH to discover the Appmesh instance with the specified `<appmesh-instance-arn>`.

Once the settings object is updated, you should see that SMH has discovered the mesh by running `kubectl -n <smh-write-namespace> get mesh`.
The name of the Mesh takes the form `appmesh-<mesh-name>-<aws-region>-<aws-account-id>`.

The [DiscoverySelector](https://docs.solo.io/service-mesh-hub/latest/reference/api/settings/#core.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher) API
allows for matching by region and tags. For instance, to discover all Appmesh meshes in `us-east-2`, apply the following Settings config:

```yaml
apiVersion: core.zephyr.solo.io/v1alpha1
kind: Settings
metadata:
  namespace: service-mesh-hub
  name: settings
spec:
  aws:
    disabled: false
    accounts:
      - accountId: "<aws-account-id>"
        meshDiscovery:
          resourceSelectors:
          - matcher:
              regions:
              - us-east-2
        eksDiscovery:
          disabled: true
```
