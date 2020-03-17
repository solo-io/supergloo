package zephyr_security

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VirtualMeshCSRClientFactory func(client client.Client) VirtualMeshCSRClient

func VirtualMeshCSRClientFactoryProvider() VirtualMeshCSRClientFactory {
	return NewVirtualMeshCSRClient
}

func NewVirtualMeshCSRClient(dynamicClient client.Client) VirtualMeshCSRClient {
	return &virtualMeshCSRClient{dynamicClient: dynamicClient}
}

type virtualMeshCSRClient struct {
	cluster       string
	dynamicClient client.Client
}

func (m *virtualMeshCSRClient) Create(
	ctx context.Context,
	csr *v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.CreateOption,
) error {
	return m.dynamicClient.Create(ctx, csr, opts...)
}

func (m *virtualMeshCSRClient) Update(
	ctx context.Context,
	csr *v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.UpdateOption,
) error {
	return m.dynamicClient.Update(ctx, csr, opts...)
}

func (m *virtualMeshCSRClient) UpdateStatus(
	ctx context.Context,
	csr *v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.UpdateOption,
) error {
	return m.dynamicClient.Status().Update(ctx, csr, opts...)
}

func (m *virtualMeshCSRClient) Delete(ctx context.Context,
	csr *v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.DeleteOption,
) error {
	return m.dynamicClient.Delete(ctx, csr, opts...)
}

func (m *virtualMeshCSRClient) Get(
	ctx context.Context,
	name, namespace string,
) (*v1alpha1.VirtualMeshCertificateSigningRequest, error) {

	csr := v1alpha1.VirtualMeshCertificateSigningRequest{}
	err := m.dynamicClient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, &csr)
	if err != nil {
		return nil, err
	}
	return &csr, nil
}

func (m *virtualMeshCSRClient) List(
	ctx context.Context,
	opts metav1.ListOptions,
) (*v1alpha1.VirtualMeshCertificateSigningRequestList, error) {

	list := v1alpha1.VirtualMeshCertificateSigningRequestList{}
	err := m.dynamicClient.List(ctx, &list, &client.ListOptions{Raw: &opts})
	if err != nil {
		return nil, err
	}
	return &list, nil
}
