package zephyr_security

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

type MeshGroupCSRClient interface {
	Create(ctx context.Context, csr *v1alpha1.MeshGroupCertificateSigningRequest, opts ...client.CreateOption) error
	Update(ctx context.Context, csr *v1alpha1.MeshGroupCertificateSigningRequest, opts ...client.UpdateOption) error
	UpdateStatus(ctx context.Context, csr *v1alpha1.MeshGroupCertificateSigningRequest, opts ...client.UpdateOption) error
	Get(ctx context.Context, name, namespace string) (*v1alpha1.MeshGroupCertificateSigningRequest, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.MeshGroupCertificateSigningRequestList, error)
	Delete(ctx context.Context, csr *v1alpha1.MeshGroupCertificateSigningRequest, opts ...client.DeleteOption) error
}
