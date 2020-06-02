package cert_signer

import (
	"context"

	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	zephyr_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/csr/certgen"
	csr_generator "github.com/solo-io/service-mesh-hub/services/csr-agent/pkg/csr-generator"
)

var (
	VirtualMeshTrustBundleNotFoundMsg = func(err error, ref *zephyr_core_types.ResourceRef) error {
		return eris.Wrapf(err, "could not get root ca bundle associated with virtual mesh %s.%s", ref.GetName(),
			ref.GetNamespace())
	}
	FailedToSignCertError = func(err error) error {
		return eris.Wrapf(err, "could not sign cert from CSR")
	}
)

func NewVirtualMeshCSRSigningProcessor(signer VirtualMeshCSRSigner) csr_generator.VirtualMeshCSRProcessor {
	return &csr_generator.VirtualMeshCSRProcessorFuncs{
		OnProcessUpsert: func(
			ctx context.Context,
			csr *zephyr_security.VirtualMeshCertificateSigningRequest,
		) *zephyr_security_types.VirtualMeshCertificateSigningRequestStatus {
			return signer.Sign(ctx, csr)
		},
		OnProcessDelete: nil,
	}
}

type certSinger struct {
	mgCertClient VirtualMeshCertClient
	csrClient    zephyr_security.VirtualMeshCertificateSigningRequestClient
	signer       certgen.Signer
}

func NewVirtualMeshCSRSigner(
	mgCertClient VirtualMeshCertClient,
	csrClient zephyr_security.VirtualMeshCertificateSigningRequestClient,
	signer certgen.Signer,
) VirtualMeshCSRSigner {
	return &certSinger{mgCertClient: mgCertClient, csrClient: csrClient, signer: signer}
}

func (c *certSinger) Sign(
	ctx context.Context,
	obj *zephyr_security.VirtualMeshCertificateSigningRequest,
) *zephyr_security_types.VirtualMeshCertificateSigningRequestStatus {
	if !c.shouldProcess(obj) {
		return nil
	}
	certData, err := c.mgCertClient.GetRootCaBundle(ctx, obj.Spec.GetVirtualMeshRef())
	if err != nil {
		wrapperErr := VirtualMeshTrustBundleNotFoundMsg(err, obj.Spec.GetVirtualMeshRef())
		obj.Status.ComputedStatus = &zephyr_core_types.Status{
			State:   zephyr_core_types.Status_INVALID,
			Message: wrapperErr.Error(),
		}
		return &obj.Status
	}

	cert, err := c.signer.GenCertFromEncodedCSR(
		obj.Spec.GetCsrData(),
		certData.RootCert,
		certData.PrivateKey,
		nil,
		certgen.YearDuration(),
		true,
	)
	if err != nil {
		wrappedErr := FailedToSignCertError(err)
		obj.Status.ComputedStatus = &zephyr_core_types.Status{
			State:   zephyr_core_types.Status_INVALID,
			Message: wrappedErr.Error(),
		}
		return &obj.Status
	}

	// set the cert on the obj object to the cert, and update it
	obj.Status.Response = &zephyr_security_types.VirtualMeshCertificateSigningRequestStatus_Response{
		CaCertificate:   cert,
		RootCertificate: certData.RootCert,
	}
	obj.Status.ComputedStatus = &zephyr_core_types.Status{
		State: zephyr_core_types.Status_ACCEPTED,
	}
	return &obj.Status
}

func (c *certSinger) shouldProcess(csr *zephyr_security.VirtualMeshCertificateSigningRequest) bool {
	// TODO: make this configurable so third party workflows can be enabled
	switch {
	// Third party approval is not in the correct state
	case csr.Status.GetThirdPartyApproval().GetApprovalStatus() != zephyr_security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow_APPROVED &&
		csr.Status.GetThirdPartyApproval().GetApprovalStatus() != zephyr_security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow_PENDING:
		return false
	// CSR data has not yet been populated
	case len(csr.Spec.GetCsrData()) == 0:
		return false
	// virtual mesh Ref hasn't been set
	case csr.Spec.GetVirtualMeshRef() == nil:
		return false
	// Both the ca cert and root cert have been populated
	case len(csr.Status.GetResponse().GetCaCertificate()) != 0 &&
		len(csr.Status.GetResponse().GetRootCertificate()) != 0:
		return false
	default:
		return true
	}
}

func (c *certSinger) ProcessDelete(
	ctx context.Context,
	csr *zephyr_security.VirtualMeshCertificateSigningRequest,
) (computedStatus *zephyr_security_types.VirtualMeshCertificateSigningRequestStatus) {
	return
}
