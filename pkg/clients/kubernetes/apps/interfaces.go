package kubernetes_apps

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_clients.go

type DeploymentClient interface {
	Get(ctx context.Context, objectKey client.ObjectKey) (*appsv1.Deployment, error)
}

type ReplicaSetClient interface {
	Get(ctx context.Context, objectKey client.ObjectKey) (*appsv1.ReplicaSet, error)
}
