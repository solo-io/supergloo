---
title: "Enterprise"
menuTitle: Enterprise
description: Install Gloo Mesh Enterprise
weight: 100
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

Gloo Mesh Enterprise uses a Kubernetes cluster to host the management plane (Gloo Mesh) while each service mesh can run on its own independent cluster. If you don't have access to multiple clusters, see the [Getting Started Guide]({{% versioned_link_path fromRoot="/getting_started/" %}}) to get started with Kubernetes in Docker, or refer to our [Using Kind]({{% versioned_link_path fromRoot="/setup/kind_setup" %}}) setup guide to provision two clusters.

{{% notice note %}}
Gloo Mesh Enterprise is the paid version of Gloo Mesh including the Gloo Mesh UI and multi-cluster role-based access control. To complete the installation you will need a license key. You can get a trial license by [requesting a demo](https://www.solo.io/products/gloo-mesh/) from the website.
{{% /notice %}}

This document describes how to install Gloo Mesh Enterprise.

A conceptual overview of the Gloo Mesh Enterprise architecture can be found [here]({{% versioned_link_path fromRoot="/concepts/relay" %}}). Make sure you have followed the [prerequisites guide]({{% versioned_link_path fromRoot="/setup/prerequisites/enterprise_prerequisites/" %}}). We also recommend following our guide on [configuring Role-based API control]({{% versioned_link_path fromRoot="/guides/configure_role_based_api//" %}}).

## Assumptions for setup

We will assume in this and following guides that we have access to two clusters and the following two contexts available in our `kubeconfig` file. 

Your actual context names will likely be different.

* `mgmt-cluster-context`
    - kubeconfig context pointing to a cluster where we will install and operate Gloo Mesh Enterprise
* `remote-cluster-context`
    - kubeconfig context pointing to a cluster where we will install and manage a service mesh using Gloo Mesh Enterprise

To verify you're running the following commands in the correct context, run:

```shell
MGMT_CONTEXT=kind-mgmt-cluster # Change value as needed
REMOTE_CONTEXT=kind-remote-cluster # Change value as needed

kubectl config use-context $MGMT_CONTEXT
```

## Install Gloo Mesh Enterprise

Below we will show examples of installing Gloo Mesh Enterprise with both `meshctl` and Helm.

### Installing with `meshctl`

`meshctl` is a CLI tool that helps bootstrap Gloo Mesh Enterprise, register clusters, describe configured resources, and more. Get the latest `meshctl` from the [releases page on solo-io/gloo-mesh](https://github.com/solo-io/gloo-mesh/releases).

You can also quickly install like this:

```shell
curl -sL https://run.solo.io/meshctl/install | sh
```

Installing Gloo Mesh Enterprise with `meshctl` is a simple process. You will use the command `meshctl install enterprise` and supply the license key, as well as any chart values you want to update, and arguments pointing to the cluster where Gloo Mesh Enterprise will be installed. For our example, we are going to install Gloo Mesh Enterprise on the cluster `mgmt-cluster`. First, let's set a variable for the license key.

```shell
GLOO_MESH_LICENSE_KEY=<your_key_here> # You'll need to supply your own key
```

We are not going to change any of the default values in the underlying chart, so the only argument needed is the license key.

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

Note that the `gloo-mesh-enterprise` Helm chart bundles multiple components, including `enterprise-networking`, `rbac-webhook`, and `gloo-mesh-ui`. Each is versioned in step with the parent `gloo-mesh-enterprise` chart, and each has its own Helm values for advanced customization.

4. Install

There are several [helm values]({{< versioned_link_path fromRoot="/reference/helm/gloo_mesh_enterprise" >}}) that you may have
to modify, depending on your setup. Here are several examples:

Case 1: Your Gloo Mesh control plane will be running on its own cluster.

Case 2: You are running Gloo Mesh Enterprise's management plane on a cluster
that you intend to register (i.e. also run a service mesh). In this case, set
the enterprise-networking.cluster value to your management cluster name. Then, register
the management cluster vie meshctl or helm with the same management cluster name.

Case 3: You intend to provide your own certificates. The Helm value
`selfSigned` is set to `true` by default. This means the Helm chart
will create certificates for you if you do not supply them through values.
Set `selfSigned` to `false` and follow the steps [here]({{< versioned_link_path fromRoot="/setup/prerequisites/enterprise_prerequisites/#manual-certificate-creation-optional" >}})
to manually create certs.


{{< tabs >}}
{{< tab name="Case 1" codelang="shell">}}
kubectl create ns gloo-mesh

helm install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise --namespace gloo-mesh \
  --set licenseKey=${GLOO_MESH_LICENSE_KEY}
{{< /tab >}}
{{< tab name="Case 2" codelang="shell">}}
CLUSTER_NAME=mgmt-cluster # Update this value as needed
kubectl create ns gloo-mesh

helm install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise --namespace gloo-mesh \
  --set licenseKey=${GLOO_MESH_LICENSE_KEY},enterprise-networking.cluster=$CLUSTER_NAME
{{< /tab >}}
{{< tab name="Case 3" codelang="shell">}}
kubectl create ns gloo-mesh

helm install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise --namespace gloo-mesh \
  --set licenseKey=${GLOO_MESH_LICENSE_KEY},enterprise-networking.selfSigned=false
{{< /tab >}}
{{< /tabs >}}


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
