package cleanup

import (
	"fmt"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common/exec"
)

func AppmeshEksCleanup(runner exec.Runner, region string, meshName string, eksClusterName string) error {
	return runner.Run("bash", fmt.Sprintf(appmeshEksCleanupScript, region, meshName, eksClusterName))
}

const (
	appmeshEksCleanupScript = `
region=%s
meshName=%s
eksClusterName=%s
awsAccountID=$(echo $(aws sts get-caller-identity --query 'Account'))

if [ -z ${awsAccountID+x} ]; 
then echo "Unable to fetch AWS account ID, check that your AWS credentials are configured properly." && exit 0 ; 
else echo "Using AWS Account ID $awsAccountID" ; 
fi

# Delete Appmesh mesh
# Note: pipe through cat to prevent the interactive aws prompt form blocking the script.
aws appmesh delete-mesh --mesh-name=$meshName | cat

# Delete OIDC provider for EKS cluster.
OIDCURL=$(aws eks describe-cluster --name $eksClusterName --output json | jq -r .cluster.identity.oidc.issuer | sed -e "s*https://**")
aws iam delete-open-id-connect-provider --open-id-connect-provider-arn arn:aws:iam::$awsAccountID:oidc-provider/$OIDCURL

# Delete EKS cluster
eksctl delete cluster --name=$eksClusterName --region=$region
`
)
