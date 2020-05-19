package aws_utils

import (
	"context"

	k8s_core_types "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type AppMeshPod struct {
	AwsAccountID    string
	Region          string
	AppMeshName     string
	VirtualNodeName string
}

// Scans pod for Appmesh envoy sidecar.
type AppMeshScanner interface {
	ScanPodForAppMesh(
		ctx context.Context,
		pod *k8s_core_types.Pod,
		remoteClient client.Client,
	) (*AppMeshPod, error)
}

type ArnParser interface {
	ParseAccountID(arn string) (string, error)
	ParseRegion(arn string) (string, error)
}

type AwsAccountIdFetcher interface {
	GetEksAccountId(
		ctx context.Context,
		clusterScopedClient client.Client,
	) (accountId string, err error)
}
