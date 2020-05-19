package aws_utils

import (
	"context"

	k8s_core_types "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type AwsAccountId string

type AppMeshPod struct {
	AwsAccountID    string
	Region          string
	AppMeshName     string
	VirtualNodeName string
}

// Scans pod for Appmesh envoy sidecar.
type AppMeshScanner interface {
	ScanPodForAppMesh(
		pod *k8s_core_types.Pod,
		awsAccountId AwsAccountId,
	) (*AppMeshPod, error)
}

type ArnParser interface {
	ParseAccountID(arn string) (string, error)
	ParseRegion(arn string) (string, error)
}

// Fetch AWS account ID from the "aws-auth.kube-system" ConfigMap.
// EKS docs suggest, but do not confirm, that this ConfigMap exists on all EKS clusters.
// Reference: https://docs.aws.amazon.com/eks/latest/userguide/add-user-role.html
// NB: This logic also assumes that the AWS account owning the EKS cluster is also the account that owns the Appmesh instance.
type AwsAccountIdFetcher interface {
	GetEksAccountId(
		ctx context.Context,
		clusterScopedClient client.Client,
	) (awsAccountID AwsAccountId, err error)
}
