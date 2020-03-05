package zephyr_networking

import (
	"context"

	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_clients.go

type MeshGroupClient interface {
	Get(ctx context.Context, name, namespace string) (*networking_v1alpha1.MeshGroup, error)
	List(ctx context.Context, opts metav1.ListOptions) (*networking_v1alpha1.MeshGroupList, error)
	UpdateStatus(ctx context.Context, meshGroup *networking_v1alpha1.MeshGroup, opts ...client.UpdateOption) error
}

type TrafficPolicyClient interface {
	Get(ctx context.Context, name string, namespace string) (*networking_v1alpha1.TrafficPolicy, error)
	Create(ctx context.Context, trafficPolicy *networking_v1alpha1.TrafficPolicy, options ...client.CreateOption) error
	Update(ctx context.Context, trafficPolicy *networking_v1alpha1.TrafficPolicy, options ...client.UpdateOption) error
	UpdateStatus(ctx context.Context, trafficPolicy *networking_v1alpha1.TrafficPolicy, options ...client.UpdateOption) error
	List(ctx context.Context, opts ...client.ListOption) (*networking_v1alpha1.TrafficPolicyList, error)
}
