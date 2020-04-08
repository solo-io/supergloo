package zephyr_security

import (
	"context"

	security_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/clientset/versioned"
	v1alpha12 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/clientset/versioned/typed/security.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func NewGeneratedVirtualMeshCSRClient(client versioned.Interface) VirtualMeshCSRClient {
	return &generatedVirtualMeshClient{client: client.SecurityV1alpha1()}
}

type generatedVirtualMeshClient struct {
	client v1alpha12.SecurityV1alpha1Interface
}

func (g *generatedVirtualMeshClient) Create(
	ctx context.Context,
	csr *security_v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.CreateOption,
) error {
	returned, err := g.client.VirtualMeshCertificateSigningRequests(csr.Namespace).Create(csr)
	if err != nil {
		return err
	}
	*csr = *returned
	return nil
}

func (g *generatedVirtualMeshClient) Update(
	ctx context.Context,
	csr *security_v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.UpdateOption,
) error {
	returned, err := g.client.VirtualMeshCertificateSigningRequests(csr.Namespace).Create(csr)
	if err != nil {
		return err
	}
	*csr = *returned
	return nil
}

func (g *generatedVirtualMeshClient) Delete(
	ctx context.Context,
	csr *security_v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.DeleteOption,
) error {
	metaOpts := &client.DeleteOptions{}
	metaOpts.ApplyOptions(opts)
	return g.client.VirtualMeshCertificateSigningRequests(csr.Namespace).Delete(csr.Name, metaOpts.AsDeleteOptions())
}

func (g *generatedVirtualMeshClient) Get(
	ctx context.Context,
	name, namespace string,
) (*security_v1alpha1.VirtualMeshCertificateSigningRequest, error) {
	return g.client.VirtualMeshCertificateSigningRequests(namespace).Get(name, v1.GetOptions{})
}

func (g *generatedVirtualMeshClient) List(
	ctx context.Context,
	opts ...client.ListOption,
) (*security_v1alpha1.VirtualMeshCertificateSigningRequestList, error) {
	listOptions := &client.ListOptions{}
	listOptions.ApplyOptions(opts)
	raw := v1.ListOptions{}
	if converted := listOptions.AsListOptions(); converted != nil {
		raw = *converted
	}
	return g.client.VirtualMeshCertificateSigningRequests(listOptions.Namespace).List(raw)

}

func (g *generatedVirtualMeshClient) UpdateStatus(
	ctx context.Context,
	vm *security_v1alpha1.VirtualMeshCertificateSigningRequest,
	opts ...client.UpdateOption,
) error {
	updated, err := g.client.VirtualMeshCertificateSigningRequests(vm.Namespace).UpdateStatus(vm)
	if err != nil {
		return err
	}
	*vm = *updated
	return nil
}
