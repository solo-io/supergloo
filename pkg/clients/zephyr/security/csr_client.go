package zephyr_security

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MeshGroupCSRClientFactory func(client client.Client) MeshGroupCSRClient

func MeshGroupCSRClientFactoryProvider() MeshGroupCSRClientFactory {
	return NewMeshGroupCSRClient
}

func NewMeshGroupCSRClient(dynamicClient client.Client) MeshGroupCSRClient {
	return &meshGroupCSRClient{dynamicClient: dynamicClient}
}

type meshGroupCSRClient struct {
	cluster       string
	dynamicClient client.Client
}

func (m *meshGroupCSRClient) Create(
	ctx context.Context,
	csr *v1alpha1.MeshGroupCertificateSigningRequest,
	opts ...client.CreateOption,
) error {
	return m.dynamicClient.Create(ctx, csr, opts...)
}

func (m *meshGroupCSRClient) Update(
	ctx context.Context,
	csr *v1alpha1.MeshGroupCertificateSigningRequest,
	opts ...client.UpdateOption,
) error {
	return m.dynamicClient.Update(ctx, csr, opts...)
}

func (m *meshGroupCSRClient) UpdateStatus(
	ctx context.Context,
	csr *v1alpha1.MeshGroupCertificateSigningRequest,
	opts ...client.UpdateOption,
) error {
	return m.dynamicClient.Status().Update(ctx, csr, opts...)
}

func (m *meshGroupCSRClient) Delete(ctx context.Context,
	csr *v1alpha1.MeshGroupCertificateSigningRequest,
	opts ...client.DeleteOption,
) error {
	return m.dynamicClient.Delete(ctx, csr, opts...)
}

func (m *meshGroupCSRClient) Get(
	ctx context.Context,
	name, namespace string,
) (*v1alpha1.MeshGroupCertificateSigningRequest, error) {

	csr := v1alpha1.MeshGroupCertificateSigningRequest{}
	err := m.dynamicClient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, &csr)
	if err != nil {
		return nil, err
	}
	return &csr, nil
}

func (m *meshGroupCSRClient) List(
	ctx context.Context,
	opts metav1.ListOptions,
) (*v1alpha1.MeshGroupCertificateSigningRequestList, error) {

	list := v1alpha1.MeshGroupCertificateSigningRequestList{}
	err := m.dynamicClient.List(ctx, &list, &client.ListOptions{Raw: &opts})
	if err != nil {
		return nil, err
	}
	return &list, nil
}
