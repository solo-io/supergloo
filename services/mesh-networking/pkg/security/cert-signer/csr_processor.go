package cert_signer

import (
	"context"

	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	security_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	"github.com/solo-io/mesh-projects/pkg/security/certgen"
	csr_generator "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator"
)

var (
	VirtualMeshTrustBundleNotFoundMsg = func(err error, ref *core_types.ResourceRef) error {
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
			csr *security_v1alpha1.VirtualMeshCertificateSigningRequest,
		) *security_types.VirtualMeshCertificateSigningRequestStatus {
			return signer.Sign(ctx, csr)
		},
		OnProcessDelete: nil,
	}
}

type certSinger struct {
	mgCertClient VirtualMeshCertClient
	csrClient    zephyr_security.VirtualMeshCSRClient
	signer       certgen.Signer
}

func NewVirtualMeshCSRSigner(
	mgCertClient VirtualMeshCertClient,
	csrClient zephyr_security.VirtualMeshCSRClient,
	signer certgen.Signer,
) VirtualMeshCSRSigner {
	return &certSinger{mgCertClient: mgCertClient, csrClient: csrClient, signer: signer}
}

func (c *certSinger) Sign(
	ctx context.Context,
	obj *security_v1alpha1.VirtualMeshCertificateSigningRequest,
) *security_types.VirtualMeshCertificateSigningRequestStatus {
	if !c.shouldProcess(obj) {
		return nil
	}
	certData, err := c.mgCertClient.GetRootCaBundle(ctx, obj.Spec.GetVirtualMeshRef())
	if err != nil {
		wrapperErr := VirtualMeshTrustBundleNotFoundMsg(err, obj.Spec.GetVirtualMeshRef())
		obj.Status.ComputedStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_INVALID,
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
		obj.Status.ComputedStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_INVALID,
			Message: wrappedErr.Error(),
		}
		return &obj.Status
	}

	// set the cert on the obj object to the cert, and update it
	obj.Status.Response = &security_types.VirtualMeshCertificateSigningResponse{
		CaCertificate:   cert,
		RootCertificate: certData.RootCert,
	}
	obj.Status.ComputedStatus = &core_types.ComputedStatus{
		Status: core_types.ComputedStatus_ACCEPTED,
	}
	return &obj.Status
}

func (c *certSinger) shouldProcess(csr *security_v1alpha1.VirtualMeshCertificateSigningRequest) bool {
	// TODO: make this configurable so third party workflows can be enabled
	switch {
	// Third party approval is not in the correct state
	case csr.Status.GetThirdPartyApproval().GetApprovalStatus() != security_types.ThirdPartyApprovalWorkflow_APPROVED &&
		csr.Status.GetThirdPartyApproval().GetApprovalStatus() != security_types.ThirdPartyApprovalWorkflow_PENDING:
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
	csr *security_v1alpha1.VirtualMeshCertificateSigningRequest,
) (computedStatus *security_types.VirtualMeshCertificateSigningRequestStatus) {
	return
}
