package zephyr_networking

import (
	"context"

	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)


func NewVirtualMeshClientForConfig(cfg *rest.Config) (VirtualMeshClient, error) {
	if err := networking_v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &virtualMeshClient{dynamicClient: dynamicClient}, nil
}

func NewVirtualMeshClient(dynamicClient client.Client) VirtualMeshClient {
	return &virtualMeshClient{dynamicClient: dynamicClient}
}

type virtualMeshClient struct {
	dynamicClient client.Client
}

func (v *virtualMeshClient) Get(
	ctx context.Context,
	name, namespace string,
) (*networking_v1alpha1.VirtualMesh, error) {

	csr := networking_v1alpha1.VirtualMesh{}
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
) (*networking_v1alpha1.VirtualMeshList, error) {

	list := networking_v1alpha1.VirtualMeshList{}
	err := v.dynamicClient.List(ctx, &list, opts...)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (v *virtualMeshClient) UpdateStatus(
	ctx context.Context,
	vm *networking_v1alpha1.VirtualMesh,
	opts ...client.UpdateOption,
) error {
	return v.dynamicClient.Status().Update(ctx, vm, opts...)
}

func (v *virtualMeshClient) Create(ctx context.Context, virtualMesh *networking_v1alpha1.VirtualMesh) error {
	return v.dynamicClient.Create(ctx, virtualMesh)
}
