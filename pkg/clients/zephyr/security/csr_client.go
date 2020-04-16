package zephyr_security

import (
	"context"

	security_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VirtualMeshCSRClientFactory func(client client.Client) VirtualMeshCSRClient

func VirtualMeshCSRClientFactoryProvider() VirtualMeshCSRClientFactory {
	return NewVirtualMeshCSRClient
}

func NewVirtualMeshCSRClient(dynamicClient client.Client) VirtualMeshCSRClient {
	return &virtualMeshCSRClient{dynamicClient: dynamicClient}
}

func NewVirtualMeshCSRClientForConfig(cfg *rest.Config) (VirtualMeshCSRClient, error) {
	if err := security_v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &virtualMeshCSRClient{dynamicClient: dynamicClient}, nil
}

type virtualMeshCSRClient struct {
	cluster       string
	dynamicClient client.Client
}

func (m *virtualMeshCSRClient) Create(
	ctx context.Context,
	csr *security_v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.CreateOption,
) error {
	return m.dynamicClient.Create(ctx, csr, opts...)
}

func (m *virtualMeshCSRClient) Update(
	ctx context.Context,
	csr *security_v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.UpdateOption,
) error {
	return m.dynamicClient.Update(ctx, csr, opts...)
}

func (m *virtualMeshCSRClient) UpdateStatus(
	ctx context.Context,
	csr *security_v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.UpdateOption,
) error {
	return m.dynamicClient.Status().Update(ctx, csr, opts...)
}

func (m *virtualMeshCSRClient) Delete(ctx context.Context,
	csr *security_v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.DeleteOption,
) error {
	return m.dynamicClient.Delete(ctx, csr, opts...)
}

func (m *virtualMeshCSRClient) Get(
	ctx context.Context,
	name, namespace string,
) (*security_v1alpha1.VirtualMeshCertificateSigningRequest, error) {

	csr := security_v1alpha1.VirtualMeshCertificateSigningRequest{}
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
	opts ...client.ListOption,
) (*security_v1alpha1.VirtualMeshCertificateSigningRequestList, error) {

	list := security_v1alpha1.VirtualMeshCertificateSigningRequestList{}
	err := m.dynamicClient.List(ctx, &list, opts...)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
