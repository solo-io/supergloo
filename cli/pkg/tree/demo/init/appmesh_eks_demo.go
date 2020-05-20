package demo_init

import (
	"fmt"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common/exec"
)

func AppmeshEksDemo(runner exec.Runner, awsRegion string) error {
	return runner.Run("bash", "-eux", "-c", fmt.Sprintf(appmeshEksDemoScript, awsRegion))
}

/*
	Prerequisites:
		1. meshctl
		2. eksctl (https://github.com/weaveworks/eksctl)
		3. Helm (https://helm.sh/docs/intro/install/)
		4. AWS API credentials must be configured, either through the `~/.aws/credentials` [file](https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html)
			or [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html)
 */
const (
	appmeshEksDemoScript = `
awsAccountID=$(echo $(aws sts get-caller-identity --query 'Account'))
region=%s
clusterName=smh-demo-cluster-$(xxd -l8 -ps /dev/urandom)
meshName=smh-demo-mesh-$(xxd -l8 -ps /dev/urandom)

if [ -z ${awsAccountID+x} ]; 
then echo "Unable to fetch AWS account ID, check that your AWS credentials are configured properly." && exit 0 ; 
else echo "Using AWS Account ID $awsAccountID" ; 
fi

# Provision EKS cluster.
echo "******* Note that provisioning a new EKS cluster can take up to 20 minutes *******"
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
aws appmesh create-mesh --mesh-name=$meshName --region=$region | cat

# Label the default namespace for appmesh injection.
kubectl label namespace default appmesh.k8s.aws/sidecarInjectorWebhook=enabled

# Manually set the CA_BUNDLE env variable to fix the issue documented here, https://github.com/aws/aws-app-mesh-inject#troubleshooting
kubectl -n appmesh-system set env deployment/appmesh-inject -c appmesh-inject CA_BUNDLE=$(kubectl config view --raw -o json --minify | jq -r '.clusters[0].cluster."certificate-authority-data"' | tr -d '"')

# Install Service Mesh Hub.
meshctl install

# Generate AWS secret
aws_access_key_id=$(echo -n $(aws configure get aws_access_key_id) | base64)
aws_secret_access_key=$(echo -n $(aws configure get aws_secret_access_key) | base64)

kubectl -n service-mesh-hub apply -f - <<EOF
apiVersion: v1
kind: Secret
type: solo.io/register/aws-credentials
metadata:
  name: smh-demo-aws
  namespace: service-mesh-hub
data:
  aws_access_key_id: $aws_access_key_id
  aws_secret_access_key: $aws_secret_access_key
EOF

# Start discovery.
meshARN=$(echo $(aws appmesh describe-mesh --mesh-name=$meshName --query 'mesh.metadata.arn'))
eksClusterARN=$(echo $(aws eks describe-cluster --name=$clusterName --query 'cluster.arn'))

kubectl -n service-mesh-hub replace -f - <<EOF
apiVersion: core.zephyr.solo.io/v1alpha1
kind: Settings
metadata:
  namespace: service-mesh-hub
  name: settings
spec:
  aws:
    accounts:
      - accountId: $awsAccountID
        meshDiscovery:
          resourceSelectors:
          - arn: $meshARN
        eksDiscovery:
          resourceSelectors:
          - arn: $eksClusterARN
EOF

# Deploy bookinfo sample application to demonstrate discovery.
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml
`
)
