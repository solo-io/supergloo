package zephyr_networking

import (
	"context"

	networkingv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewVirtualMeshClient(dynamicClient client.Client) VirtualMeshClient {
	return &virtualMeshClient{dynamicClient: dynamicClient}
}

type virtualMeshClient struct {
	dynamicClient client.Client
}

func (v *virtualMeshClient) Get(
	ctx context.Context,
	name, namespace string,
) (*networkingv1alpha1.VirtualMesh, error) {

	csr := networkingv1alpha1.VirtualMesh{}
	err := v.dynamicClient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, &csr)
	if err != nil {
		return nil, err
	}
	return &csr, nil
}

func (v *virtualMeshClient) List(
	ctx context.Context,
	opts ...client.ListOption,
) (*networkingv1alpha1.VirtualMeshList, error) {

	list := networkingv1alpha1.VirtualMeshList{}
	err := v.dynamicClient.List(ctx, &list, opts...)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (v *virtualMeshClient) UpdateStatus(
	ctx context.Context,
	vm *networkingv1alpha1.VirtualMesh,
	opts ...client.UpdateOption,
) error {
	return v.dynamicClient.Status().Update(ctx, vm, opts...)
}

func (v *virtualMeshClient) Create(ctx context.Context, virtualMesh *networkingv1alpha1.VirtualMesh) error {
	return v.dynamicClient.Create(ctx, virtualMesh)
}
