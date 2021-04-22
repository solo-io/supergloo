---
title: "Enterprise"
menuTitle: Enterprise
description: Install Gloo Mesh Enterprise
weight: 100
---

{{% notice note %}}
Gloo Mesh Enterprise is the paid version of Gloo Mesh including the Gloo Mesh UI and multi-cluster role-based access control. To complete the installation you will need a license key. You can get a trial license by [requesting a demo](https://www.solo.io/products/gloo-mesh/) from the website.
{{% /notice %}}

In a typical deployment, Gloo Mesh Enterprise uses a single Kubernetes cluster to host the management plane while each service mesh can run on its own independent cluster.
This document describes how to install the Gloo Mesh Enterprise management plane components with both `meshctl` and Helm.

A conceptual overview of the Gloo Mesh Enterprise architecture can be found [here]({{% versioned_link_path fromRoot="/concepts/relay" %}}).

### Installing with `meshctl`

`meshctl` is a CLI tool that helps bootstrap Gloo Mesh Enterprise, register clusters, describe configured resources, and more. Get the latest `meshctl` from the [releases page on solo-io/gloo-mesh](https://github.com/solo-io/gloo-mesh/releases).

You can also quickly install like this:

```shell
curl -sL https://run.solo.io/meshctl/install | sh
```

Installing Gloo Mesh Enterprise with `meshctl` is a simple process. You will use the command `meshctl install enterprise` and supply the license key, as well as any chart values you want to update, and arguments pointing to the cluster where Gloo Mesh Enterprise will be installed. For our example, we are going to install Gloo Mesh Enterprise on the cluster currently in context. First, let's set a variable for the license key.

```shell
GLOO_MESH_LICENSE_KEY=<your_key_here> # You'll need to supply your own key
```

We are not going to change any of the default values in the underlying chart, so the only argument needed is the license key. However `meshctl install enterprise` is backed by the Gloo Mesh Enterprise Helm chart, so you can customize your installation with a Helm values override via the flag `--chart-values-file`. Review the Gloo Mesh Enterprise Helm values documentation [here]({{% versioned_link_path fromRoot="/reference/helm/gloo_mesh_enterprise/" %}}).

```shell
meshctl install enterprise --license $GLOO_MESH_LICENSE_KEY
```

You should see the following output from the command:

```shell
Installing Helm chart
Finished installing chart 'gloo-mesh-enterprise' as release gloo-mesh:gloo-mesh
```

The installer has created the namespace `gloo-mesh` and installed Gloo Mesh Enterprise into the namespace using a Helm chart with default values.

{{% notice note %}}`meshctl` will create a self-signed certificate authority for mTLS if you do not supply your own certificates.{{% /notice %}}

To undo the installation, you can simply run the `uninstall` command:

```shell
meshctl uninstall
```

## Helm

You may prefer to use the Helm chart directly rather than using the `meshctl` CLI tool. This section will take you through the steps necessary to deploy a Gloo Mesh Enterprise installation from the Helm chart.

1. Add the Helm repo

```shell
helm repo add gloo-mesh-enterprise https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise
```

2. (optional) View available versions

```shell
helm search repo gloo-mesh-enterprise
```

3. (optional) View Helm values

```shell
helm show values gloo-mesh-enterprise/gloo-mesh-enterprise
```

Note that the `gloo-mesh-enterprise` Helm chart bundles multiple components, including `enterprise-networking`, `rbac-webhook`, and `gloo-mesh-ui`. Each is versioned in step with the parent `gloo-mesh-enterprise` chart, and each has its own Helm values for advanced customization. Review the Gloo Mesh Enterprise Helm values documentation [here]({{% versioned_link_path fromRoot="/reference/helm/gloo_mesh_enterprise/" %}}).

4. Install

{{% notice note %}}If you are running Gloo Mesh Enterprise's management plane on a cluster you intend to register (i.e. also run a service mesh), set the `enterprise-networking.cluster` value to the cluster name you intend to set for the management cluster at registration time. {{% /notice %}}

```shell
kubectl create ns gloo-mesh

helm install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise --namespace gloo-mesh \
  --set licenseKey=${GLOO_MESH_LICENSE_KEY}
```

{{% notice note %}}The Helm value `selfSigned` is set to `true` by default. This means the Helm chart will create certificates for you if you do not supply them through values.{{% /notice %}}

### Verify install
Once you've installed Gloo Mesh, verify what components were installed:

```shell
kubectl get pods -n gloo-mesh
```

```shell
NAME                                     READY   STATUS    RESTARTS   AGE
dashboard-6d6b944cdb-jcpvl               3/3     Running   0          4m2s
enterprise-networking-84fc9fd6f5-rrbnq   1/1     Running   0          4m2s
rbac-webhook-84865cb7dd-sbwp7            1/1     Running   0          4m2s
```

Running the check command from meshctl will also verify everything was installed correctly:

```shell
meshctl check
```

```shell
Gloo Mesh
-------------------
✅ Gloo Mesh pods are running

Management Configuration
---------------------------
✅ Gloo Mesh networking configuration resources are in a valid state
```

## Next Steps

Now that we have Gloo Mesh Enterprise installed, let's [register a cluster]({{< versioned_link_path fromRoot="/setup/cluster_registration/enterprise_cluster_registration" >}}).
