package rest

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/solo-io/skv2/pkg/multicluster/discovery"
	"github.com/solo-io/skv2/pkg/multicluster/discovery/cloud"
)

const (
	MaxResults = 100
)

type eksFinder struct {
	eksClient        cloud.EksClient
	eksConfigBuilder discovery.EksConfigBuilder
}

func (e *eksFinder) Reconcile(ctx context.Context) error {
	input := &eks.ListClustersInput{
		MaxResults: aws.Int64(MaxResults),
	}
	for {
		listClustersOutput, err := e.eksClient.ListClusters(ctx, input)
		if err != nil {
			return err
		}
		for _, clusterName := range listClustersOutput.Clusters {
			err := e.registerCluster(ctx, aws.StringValue(clusterName))
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

func (e *eksFinder) registerCluster(ctx context.Context, clusterName string) error {
	cluster, err := e.eksClient.DescribeCluster(ctx, clusterName)
	if err != nil {
		return err
	}
	_, err = e.eksConfigBuilder.ConfigForCluster(ctx, cluster)
	if err != nil {
		return err
	}
	return nil
}
