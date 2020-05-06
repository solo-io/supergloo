package eks

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	compute_target_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws"
	eks_client "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/clients/eks"
	"github.com/solo-io/skv2/pkg/multicluster/discovery/cloud"
)

const (
	MaxResults = 100
)

type eksReconciler struct {
	eksClientFactory          eks_client.EksClientFactory
	eksConfigBuilderFactory   eks_client.EksConfigBuilderFactory
	clusterRegistrationClient clients.ClusterRegistrationClient
}

func NewEksDiscoveryReconciler(
	eksClientFactory eks_client.EksClientFactory,
	eksConfigBuilderFactory eks_client.EksConfigBuilderFactory,
	clusterRegistrationClient clients.ClusterRegistrationClient,
) compute_target_aws.EksDiscoveryReconciler {
	return &eksReconciler{
		eksClientFactory:          eksClientFactory,
		eksConfigBuilderFactory:   eksConfigBuilderFactory,
		clusterRegistrationClient: clusterRegistrationClient,
	}
}

func (e *eksReconciler) Reconcile(ctx context.Context, creds *credentials.Credentials, region string) error {
	eksClient, err := e.eksClientFactory(creds, region)
	if err != nil {
		return err
	}
	input := &eks.ListClustersInput{
		MaxResults: aws.Int64(MaxResults),
	}
	for {
		listClustersOutput, err := eksClient.ListClusters(ctx, input)
		if err != nil {
			return err
		}
		for _, clusterName := range listClustersOutput.Clusters {
			err := e.registerCluster(ctx, eksClient, aws.StringValue(clusterName))
			if err != nil {
				return err
			}
		}
		if listClustersOutput.NextToken == nil {
			break
		}
		input.NextToken = listClustersOutput.NextToken
	}
	return nil
}

func (e *eksReconciler) registerCluster(
	ctx context.Context,
	eksClient cloud.EksClient,
	clusterName string,
) error {
	cluster, err := eksClient.DescribeCluster(ctx, clusterName)
	if err != nil {
		return err
	}
	config, err := e.eksConfigBuilderFactory(eksClient).ConfigForCluster(ctx, cluster)
	if err != nil {
		return err
	}
	return e.clusterRegistrationClient.Register(
		ctx,
		config,
		clusterName,
		env.GetWriteNamespace(), // TODO make this configurable
		false,
		false,
		"",
		"",
	)
}
