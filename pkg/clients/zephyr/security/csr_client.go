package zephyr_security

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MeshGroupCertificateSigningRequestClientFactory func(client client.Client) MeshGroupCertificateSigningRequestClient

func MeshGroupCertificateSigningRequestClientFactoryProvider() MeshGroupCertificateSigningRequestClientFactory {
	return NewMeshGroupCertificateSigningRequestClient
}

func NewMeshGroupCertificateSigningRequestClient(dynamicClient client.Client) MeshGroupCertificateSigningRequestClient {
	return &meshGroupCertificateSigningRequestClient{dynamicClient: dynamicClient}
}

type meshGroupCertificateSigningRequestClient struct {
	cluster       string
	dynamicClient client.Client
}

func (m *meshGroupCertificateSigningRequestClient) Create(
	ctx context.Context,
	csr *v1alpha1.MeshGroupCertificateSigningRequest,
	opts ...client.CreateOption,
) error {
	return m.dynamicClient.Create(ctx, csr, opts...)
}

func (m *meshGroupCertificateSigningRequestClient) Update(
	ctx context.Context,
	csr *v1alpha1.MeshGroupCertificateSigningRequest,
	opts ...client.UpdateOption,
) error {
	return m.dynamicClient.Update(ctx, csr, opts...)
}

func (m *meshGroupCertificateSigningRequestClient) UpdateStatus(
	ctx context.Context,
	csr *v1alpha1.MeshGroupCertificateSigningRequest,
	opts ...client.UpdateOption,
) error {
	return m.dynamicClient.Status().Update(ctx, csr, opts...)
}

func (m *meshGroupCertificateSigningRequestClient) Delete(ctx context.Context,
	csr *v1alpha1.MeshGroupCertificateSigningRequest,
	opts ...client.DeleteOption,
) error {
	return m.dynamicClient.Delete(ctx, csr, opts...)
}

func (m *meshGroupCertificateSigningRequestClient) Get(
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

func (m *meshGroupCertificateSigningRequestClient) List(
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
