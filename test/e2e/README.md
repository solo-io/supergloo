# Running Locally

Running locally should "just work"™ (run `RUN_E2E=1 ginkgo`).
If you want to use an existing cluster, you can run the tests like so:

```
RUN_E2E=1 USE_EXISTING=mgmt_cluster_ctx,remote_cluster_ctx ginkgo
```

Where `mgmt_cluster_ctx` and `remote_cluster_ctx` are the kube contexts for the management and remote clusters. This assumes that the clusters are already setup with service mesh hub.

## General Notes

Our general strategy for e2e testing is to deploy the [Istio bookinfo sample application](https://istio.io/docs/examples/bookinfo/)
on the clusters that we're testing on (either local Kind clusters or cloud-provisioned clusters).

We modify the productpage deployment to [include an addtional container](https://github.com/solo-io/service-mesh-hub/blob/a5c99a85026ac69b2b7bca5666eed92a51e6465f/ci/bookinfo.yaml#L329)
for curling the various bookinfo microservices *where the traffic is controlled by the envoy sidecar*. 
The productpage container does not contain the curl utility. The injected envoy sidecar cannot be used either because it assumes 
UID 1337, which is configured such that its egress traffic bypasses the envoy proxy entirely.

## Appmesh EKS

The e2e tests for Appmesh / EKS rely on a persistent EKS cluster and Appmesh instance
shared across all tests.

There's always the possibility that its state gets corrupted. To recreate it, run the following commands:

```shell script
awsAccountID=$(echo $(aws sts get-caller-identity --query 'Account'))
region=%s
clusterName=smh-e2e-test
meshName=smh-e2e-test

eksctl create cluster --name=$clusterName \
--region $region \
--nodes 1 \
--appmesh-access

# Associate an OIDC provider for that cluster.
eksctl utils associate-iam-oidc-provider \
    --region $region \
    --cluster $clusterName \
    --approve

# Create IAM serviceaccount for appmesh-controller workload.
eksctl create iamserviceaccount \
    --cluster $clusterName \
    --namespace appmesh-system \
    --name appmesh-controller \
    --attach-policy-arn  arn:aws:iam::aws:policy/AWSCloudMapFullAccess,arn:aws:iam::aws:policy/AWSAppMeshFullAccess \
    --override-existing-serviceaccounts \
    --approve

# Install appmesh-controller
helm install appmesh-controller eks/appmesh-controller \
    --namespace appmesh-system \
    --set region=$region \
    --set serviceAccount.create=false \
    --set serviceAccount.name=appmesh-controller

# Install appmesh-inject
helm install appmesh-inject eks/appmesh-inject \
    --namespace appmesh-system \
    --set mesh.name=$meshName \
    --set mesh.create=true

# Create Appmesh mesh.
# Note: pipe through cat to prevent the interactive aws prompt form blocking the script.
aws appmesh create-mesh --mesh-name=$meshName --region=$region | cat

# Label the default namespace for appmesh injection.
kubectl label namespace default appmesh.k8s.aws/sidecarInjectorWebhook=enabled

# Manually set the CA_BUNDLE env variable to fix the issue documented here, https://github.com/aws/aws-app-mesh-inject#troubleshooting
kubectl -n appmesh-system set env deployment/appmesh-inject -c appmesh-inject CA_BUNDLE=$(kubectl config view --raw -o json --minify | jq -r '.clusters[0].cluster."certificate-authority-data"' | tr -d '"')
```
