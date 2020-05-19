package aws_utils

import (
	"context"

	"github.com/keikoproj/aws-auth/pkg/mapper"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var (
	AwsAuthConfigMapKey = client.ObjectKey{Name: "aws-auth", Namespace: "kube-system"}
)

type awsAccountIdFetcher struct {
	arnParser              ArnParser
	configMapClientFactory k8s_core.ConfigMapClientFactory
}

func NewAwsAccountIdFetcher(
	arnParser ArnParser,
	configMapClientFactory k8s_core.ConfigMapClientFactory,
) AwsAccountIdFetcher {
	return &awsAccountIdFetcher{
		arnParser:              arnParser,
		configMapClientFactory: configMapClientFactory,
	}
}

func (a *awsAccountIdFetcher) GetEksAccountId(
	ctx context.Context,
	clusterScopedClient client.Client,
) (accountId string, err error) {
	configMap, err := a.configMapClientFactory(clusterScopedClient).GetConfigMap(ctx, AwsAuthConfigMapKey)
	if err != nil && !errors.IsNotFound(err) {
		return "", err
	}
	if configMap == nil {
		return "", nil
	}
	var authData mapper.AwsAuthData
	mapRoles, ok := configMap.Data["mapRoles"]
	if !ok {
		return "", nil
	}
	err = yaml.Unmarshal([]byte(mapRoles), &authData.MapRoles)
	if err != nil {
		return "", err
	}
	// Fetch AWS account ID from the "aws-auth.kube-system" ConfigMap.
	// EKS docs suggest, but do not confirm, that this ConfigMap exists on all EKS clusters.
	// Reference: https://docs.aws.amazon.com/eks/latest/userguide/add-user-role.html
	// NB: This logic also assumes that the AWS account owning the EKS cluster is also the account that owns the Appmesh instance.
	for _, mapRole := range authData.MapRoles {
		if mapRole.Username == "system:node:{{EC2PrivateDNSName}}" {
			accountID, err := a.arnParser.ParseAccountID(mapRole.RoleARN)
			if err != nil {
				return "", err
			}
			return accountID, nil
		}
	}
	return "", nil
}
