package aws_utils

import (
	"context"

	k8s_core_providers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/providers"
	"github.com/solo-io/skv2/pkg/multicluster"

	"github.com/keikoproj/aws-auth/pkg/mapper"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var (
	AwsAuthConfigMapKey = client.ObjectKey{Name: "aws-auth", Namespace: "kube-system"}
)

type awsAccountIdFetcher struct {
	arnParser              ArnParser
	configMapClientFactory k8s_core_providers.ConfigMapClientFactory
	mcClient               multicluster.Client
}

func NewAwsAccountIdFetcher(
	arnParser ArnParser,
	configMapClientFactory k8s_core_providers.ConfigMapClientFactory,
	mcClient multicluster.Client,
) AwsAccountIdFetcher {
	return &awsAccountIdFetcher{
		arnParser:              arnParser,
		configMapClientFactory: configMapClientFactory,
		mcClient:               mcClient,
	}
}

func (a *awsAccountIdFetcher) GetEksAccountId(
	ctx context.Context,
	clusterName string,
) (AwsAccountId, error) {
	clusterClient, err := a.mcClient.Cluster(clusterName)
	if err != nil {
		return "", err
	}
	configMap, err := a.configMapClientFactory(clusterClient).GetConfigMap(ctx, AwsAuthConfigMapKey)
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
	for _, mapRole := range authData.MapRoles {
		if mapRole.Username == "system:node:{{EC2PrivateDNSName}}" {
			accountID, err := a.arnParser.ParseAccountID(mapRole.RoleARN)
			if err != nil {
				return "", err
			}
			return AwsAccountId(accountID), nil
		}
	}
	return "", nil
}
