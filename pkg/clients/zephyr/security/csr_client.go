package zephyr_security

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	NilArgsError = eris.New("pointer args cannot be nil")
)

func NewCertificateSigningRequestClient(dynamicClient client.Client) CertificateSigningRequestClient {
	return &certificateSigningRequestClient{dynamicClient: dynamicClient}
}

type certificateSigningRequestClient struct {
	cluster       string
	dynamicClient client.Client
}

func (c *certificateSigningRequestClient) Update(
	ctx context.Context,
	csr *v1alpha1.MeshGroupCertificateSigningRequest,
	opts ...client.UpdateOption,
) error {
	return c.dynamicClient.Update(ctx, csr, opts...)
}

func (c *certificateSigningRequestClient) UpdateStatus(
	ctx context.Context,
	csrStatus *security_types.MeshGroupCertificateSigningRequestStatus,
	objMeta *metav1.ObjectMeta,
	opts ...client.UpdateOption,
) error {
	if csrStatus == nil || objMeta == nil {
		return NilArgsError
	}
	csrToUpdate := &v1alpha1.MeshGroupCertificateSigningRequest{
		ObjectMeta: *objMeta,
		Status:     *csrStatus,
	}
	return c.dynamicClient.Status().Update(ctx, csrToUpdate, opts...)
}

func (c *certificateSigningRequestClient) Delete(ctx context.Context,
	csr *v1alpha1.MeshGroupCertificateSigningRequest,
	opts ...client.DeleteOption,
) error {
	return c.dynamicClient.Delete(ctx, csr, opts...)
}

func (c *certificateSigningRequestClient) Get(
	ctx context.Context,
	name, namespace string,
) (*v1alpha1.MeshGroupCertificateSigningRequest, error) {

	csr := v1alpha1.MeshGroupCertificateSigningRequest{}
	err := c.dynamicClient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, &csr)
	if err != nil {
		return nil, err
	}
	return &csr, nil
}

func (c *certificateSigningRequestClient) List(
	ctx context.Context,
	opts metav1.ListOptions,
) (*v1alpha1.MeshGroupCertificateSigningRequestList, error) {

	list := v1alpha1.MeshGroupCertificateSigningRequestList{}
	err := c.dynamicClient.List(ctx, &list, &client.ListOptions{Raw: &opts})
	if err != nil {
		return nil, err
	}
	return &list, nil
}
