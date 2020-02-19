package kubernetes_apps

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type replicaSetClient struct {
	client client.Client
}

type ReplicaSetClientFactory func(client client.Client) ReplicaSetClient

func ReplicaSetClientFactoryProvider() ReplicaSetClientFactory {
	return NewReplicaSetClient
}

func NewReplicaSetClient(client client.Client) ReplicaSetClient {
	return &replicaSetClient{client: client}
}

func (d *replicaSetClient) GetReplicaSet(ctx context.Context, objectKey client.ObjectKey) (*appsv1.ReplicaSet, error) {
	replicaSet := &appsv1.ReplicaSet{}
	err := d.client.Get(ctx, objectKey, replicaSet)
	if err != nil {
		return nil, err
	}
	return replicaSet, nil
}
