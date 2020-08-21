---
title: "EKS Guide"
menuTitle: EKS Guide
description: Guide for getting started using Service Mesh Hub with EKS.
weight: 120
---

# Service Mesh Hub with Appmesh

[EKS](https://aws.amazon.com/eks/) provides hosted Kubernetes clusters as a service.

Service Mesh Hub provides EKS cluster discovery and can manage the cluster and any service meshes on that cluster.

In this guide, we will configure SMH to automatically discover and register EKS clusters.

## Prerequisites

There are three pre-requisites to following these guides:

1. Install `kubectl`
    - https://kubernetes.io/docs/tasks/tools/install-kubectl/
2. Install `meshctl`
    - `curl -sL https://run.solo.io/meshctl/install | sh && export PATH=$PATH:$HOME/.service-mesh-hub/bin`
    - see the [guide]({{% versioned_link_path fromRoot="/getting_started" %}})
3. Install `helm`, [instructions here](https://helm.sh/docs/intro/install/)

## Provision an Appmesh-enabled EKS cluster

{{% notice note %}}
This step can be skipped if you have an existing EKS cluster you can use for testing (note that as this is still a developing feature,
 some SMH related resources may not be cleaned up).
{{% /notice %}}

Also note that enabling Appmesh is optional, we include it here for users who are looking to use Appmesh with EKS.

1. Install `eksctl`, the [official CLI for EKS](https://eksctl.io/)

2. Create the cluster with the following command, making the appropriate substitutions:

```shell script
eksctl create cluster --name=<cluster-name> \
--region=<aws-region> \
--nodes <num-nodes> \
--node-volume-size=<node-volume-size> \
--appmesh-access
```

3. Create an [Open ID Connect issuer (OIDC)](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html) for the cluster with the following command
(this is necessary for authorizing resources on the cluster to the AWS API):

```shell script
eksctl utils associate-iam-oidc-provider \
    --region=<region> \
    --cluster <cluster-namne> \
    --approve
```

4. Create a k8s service account and associate it with an AWS policy role for the appmesh-controller (installed in the next step) with the following command:

```shell script
eksctl create iamserviceaccount \
    --cluster <cluster-name> \
    --namespace appmesh-system \
    --name appmesh-controller \
    --attach-policy-arn  arn:aws:iam::aws:policy/AWSCloudMapFullAccess,arn:aws:iam::aws:policy/AWSAppMeshFullAccess \
    --override-existing-serviceaccounts \
    --approve
```

5. Install the [appmesh-controller](https://github.com/aws/aws-app-mesh-controller-for-k8s) (this allows configuration of appmesh resources through k8s CRDs):


Add the EKS helm repo:
```shell script
helm repo add eks https://aws.github.io/eks-charts
```

Install the controller:
```shell script
helm install appmesh-controller eks/appmesh-controller \
    --namespace appmesh-system \
    --set region=<region> \
    --set serviceAccount.create=false \
    --set serviceAccount.name=appmesh-controller
```

6. Create an Appmesh instance (this can be skipped if you have an Appmesh instance already):

```
aws appmesh create-mesh --mesh-name=<mesh-name> --region=<region>
```

7. Install the appmesh-injector (this component automatically injects labelled namespaces with the appmesh envoy sidecar):

```shell script
helm upgrade -i appmesh-inject eks/appmesh-inject \
    --namespace appmesh-system \
    --set mesh.name=<mesh-name>
```

8. Run the following command and copy its resulting value:

```shell script
$(kubectl config view --raw -o json --minify | jq -r '.clusters[0].cluster."certificate-authority-data"' | tr -d '"')
```

Edit the appmesh-inject deployment by adding/replacing the environment variable `CA_BUNDLE` with above output
for the container entry with image name `appmesh-inject`. It should look similar to:

```yaml
spec:
      containers:
      - command:
        - ./appmeshinject
        - -sidecar-image=840364872350.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.12.3.0-prod
        - -sidecar-cpu-requests=10m
        - -sidecar-memory-requests=32Mi
        - -init-image=111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:v2
        - -enable-stats-tags=false
        - -enable-statsd=false
        env:
        - name: APPMESH_NAME
          value: appmesh-test
        - name: APPMESH_LOG_LEVEL
          value: info
        - name: CA_BUNDLE
          value: LS0tLS1CRUdJTiB.....S1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
        image: 602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon/aws-app-mesh-inject:v0.5.0
...
```

9. Label a namespace for sidecar injection:

````shell script
kubectl label namespace <namespace> appmesh.k8s.aws/sidecarInjectorWebhook=enabled
````

10. (Optional) Install the bookinfo app:

```k -n <namespace> apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml```

## Grant Permissions to AWS API

In order for Service Mesh Hub to interact with AWS resources, credentials for the AWS account
must be granted by creating a k8s secret containing the API credentials in the k8s cluster
on which Service Mesh Hub is deployed. The secret must take the following form:

```yaml
apiVersion: v1
kind: Secret
type: solo.io/register/aws-credentials
metadata:
  name: <secret-name>
  namespace: <namespace>
stringData:
  aws_access_key_id: <aws_access_key_id>
  aws_secret_access_key: <aws_secret_access_key>
```

Once this secret is created, the logs for `mesh-networking` should show a log entry like the following:

```json
{"level":"info","ts":1595251860.7976727,"logger":"mesh-networking",
"msg":"Adding new compute target with name: <secret-name>","version":"0.6.1"}
```

## Configure EKS discovery

Upon installation of Service Mesh Hub v0.4.12+, you should see a `settings.core.smh.solo.io` CRD instance with the name 
`settings` in the SMH write namespace (by default this is `service-mesh-hub`), populated with

```yaml
spec:
  aws:
    disabled: true
```

By default discovery for AWS resources is disabled. To enable discovery for your EKS cluster, replace the Settings spec with the following,
making the relevant substitutions (note for simplicity we disable Appmesh discovery, see the guides section for a tutorial on Appmesh discovery):

```yaml
apiVersion: core.smh.solo.io/v1alpha1
kind: Settings
metadata:
  namespace: service-mesh-hub
  name: settings
spec:
  aws:
    accounts:
      - accountId: "<aws-account-id>"
        eksDiscovery:
          resourceSelectors:
          - arn: <eks-cluster-arn>
        meshDiscovery:
          disabled: true
```

This configuration instructs SMH to discover the EKS cluster with the specified `<eks-cluster-arn>`.

Once the settings object is updated, you should see that SMH has discovered the cluster by running `kubectl -n <smh-write-namespace> get kubernetescluster`.
The name of the KubernetesCluster takes the form `eks-<eks-cluster-name>-<region>`.

Note that workload (MeshWorkload CRD) and service (MeshService CRD) discovery will not occur unless the workloads correspond
to an existing discovered Mesh.

The [DiscoverySelector]({{% versioned_link_path fromRoot="/reference/api/settings/#core.smh.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher" %}}) API
allows for matching by region and tags. For instance, to discover all EKS clusters in `us-east-2`, apply the following Settings config:

```yaml
apiVersion: core.smh.solo.io/v1alpha1
kind: Settings
metadata:
  namespace: service-mesh-hub
  name: settings
spec:
  aws:
    disabled: false
    accounts:
      - accountId: "<aws-account-id>"
        eksDiscovery:
          resourceSelectors:
          - matcher:
              regions:
              - us-east-2
        meshDiscovery:
          disabled: true
```
