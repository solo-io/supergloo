package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NodeClientFactory func(client client.Client) NodeClient

func NewNodeClientFactory() NodeClientFactory {
	return NewNodeClient
}

func NewNodeClient(client client.Client) NodeClient {
	return &nodeClient{
		client: client,
	}
}

type nodeClient struct {
	client client.Client
}

func (s *nodeClient) Get(ctx context.Context, name string) (*corev1.Node, error) {
	node := corev1.Node{}
	err := s.client.Get(ctx, client.ObjectKey{Name: name}, &node)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *nodeClient) List(ctx context.Context, options ...client.ListOption) (*corev1.NodeList, error) {
	nodeList := corev1.NodeList{}
	err := s.client.List(ctx, &nodeList, options...)
	if err != nil {
		return &nodeList, err
	}
	return &nodeList, nil
}
