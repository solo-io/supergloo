package zephyr_networking

import (
	"context"

	networkingv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

type MeshGroupClient interface {
	Get(ctx context.Context, name, namespace string) (*networkingv1alpha1.MeshGroup, error)
	List(ctx context.Context, opts v1.ListOptions) (*networkingv1alpha1.MeshGroupList, error)
	UpdateStatus(ctx context.Context, meshGroup *networkingv1alpha1.MeshGroup, opts ...client.UpdateOption) error
}
