package csr_agent_controller

import (
	"bytes"
	"context"
	"strings"

	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	security_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	"github.com/solo-io/mesh-projects/pkg/security/certgen"
	cert_secrets "github.com/solo-io/mesh-projects/pkg/security/secrets"
	security_processors "github.com/solo-io/mesh-projects/services/common/processors/security"
	pki_util "istio.io/istio/security/pkg/pki/util"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	IstioCaSecretName  = "cacerts"
	UnexpectedEventMsg = "unexpected event for CSR"
)

var (
	FailedToRetrievePrivateKeyError = func(err error) error {
		return eris.Wrapf(err, "failed to retrieve private key")
	}
	FailedToGenerateCSRError = func(err error) error {
		return eris.Wrapf(err, "failed to generate CSR")
	}
	FailesToAddCsrToResource = func(err error) error {
		return eris.Wrapf(err, "failed to update resource with csr bytes")
	}
	FailedToUpdateCaError = func(err error) error {
		return eris.Wrapf(err, "failed to update ca")
	}
)

type istioCsrProcessor struct {
	csrClient    zephyr_security.MeshGroupCertificateSigningRequestClient
	secretClient kubernetes_core.SecretsClient
	certClient   CertClient
	signer       certgen.Signer
}

func NewCsrAgentIstioProcessor(
	csrClient zephyr_security.MeshGroupCertificateSigningRequestClient,
	secretClient kubernetes_core.SecretsClient,
	certClient CertClient,
	signer certgen.Signer,
) security_processors.MeshGroupCertificateSigningRequestProcessor {
	return &istioCsrProcessor{
		csrClient:    csrClient,
		secretClient: secretClient,
		certClient:   certClient,
		signer:       signer,
	}
}

func (i *istioCsrProcessor) ProcessCreate(
	ctx context.Context,
	obj *security_v1alpha1.MeshGroupCertificateSigningRequest,
) security_types.MeshGroupCertificateSigningRequestStatus {
	return i.process(ctx, obj)
}

func (i *istioCsrProcessor) ProcessUpdate(
	ctx context.Context,
	old, new *security_v1alpha1.MeshGroupCertificateSigningRequest,
) security_types.MeshGroupCertificateSigningRequestStatus {
	return i.process(ctx, new)
}

func (i *istioCsrProcessor) process(
	ctx context.Context,
	obj *security_v1alpha1.MeshGroupCertificateSigningRequest,
) security_types.MeshGroupCertificateSigningRequestStatus {
	rootCaData, err := i.certClient.EnsureSecretKey(ctx, obj)
	if err != nil {
		wrapped := FailedToRetrievePrivateKeyError(err)
		obj.Status.ComputedStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_INVALID,
			Message: wrapped.Error(),
		}
		return obj.Status
	}

	// csr data has not yet been saturated, meaning this is a new request
	if len(obj.Spec.GetCsrData()) == 0 {
		return i.generateCsr(ctx, obj, rootCaData)
	} else {
		// csr data has been saturated, this csr is ready to be reprocessed
		if err = i.updateCa(ctx, obj, rootCaData); err != nil {
			wrapped := FailedToUpdateCaError(err)
			obj.Status.ComputedStatus = &core_types.ComputedStatus{
				Status:  core_types.ComputedStatus_INVALID,
				Message: wrapped.Error(),
			}
			return obj.Status
		}
		obj.Status.ComputedStatus = &core_types.ComputedStatus{Status: core_types.ComputedStatus_ACCEPTED}
	}

	return obj.Status
}

func (i *istioCsrProcessor) generateCsr(
	ctx context.Context,
	obj *security_v1alpha1.MeshGroupCertificateSigningRequest,
	rootCaData *cert_secrets.RootCaData,
) security_types.MeshGroupCertificateSigningRequestStatus {
	csr, err := i.signer.GenCSRWithKey(pki_util.CertOptions{
		Host:          strings.Join(obj.Spec.GetHosts(), ","),
		Org:           obj.Spec.GetOrg(),
		SignerPrivPem: rootCaData.CaPrivateKey,
	})
	if err != nil {
		wrapped := FailedToGenerateCSRError(err)
		obj.Status.ComputedStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_INVALID,
			Message: wrapped.Error(),
		}
		return obj.Status
	}

	obj.Spec.CsrData = csr
	if err = i.csrClient.Update(ctx, obj); err != nil {
		wrapped := FailesToAddCsrToResource(err)
		obj.Status.ComputedStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_INVALID,
			Message: wrapped.Error(),
		}
		return obj.Status
	}
	obj.Status.ComputedStatus = &core_types.ComputedStatus{Status: core_types.ComputedStatus_ACCEPTED}
	return obj.Status
}

func (i *istioCsrProcessor) updateCa(
	ctx context.Context,
	obj *security_v1alpha1.MeshGroupCertificateSigningRequest,
	rootCaData *cert_secrets.RootCaData,
) error {
	rootCaData.CaCert = obj.Status.GetResponse().GetCaCertificate()
	rootCaData.RootCert = obj.Status.GetResponse().GetRootCertificate()
	rootCaData.CertChain = certgen.AppendRootCerts(rootCaData.CaCert, rootCaData.RootCert)
	secretName, secretNamespace := IstioCaSecretName, "istio-system"

	certSecret := rootCaData.BuildSecret(secretName, secretNamespace)
	existing, err := i.secretClient.Get(ctx, secretName, secretNamespace)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		return i.secretClient.Create(ctx, certSecret)
	}

	if i.certsAreEqual(existing.Data, certSecret.Data) {
		return nil
	}

	existing.Data = certSecret.Data
	return i.secretClient.Update(ctx, existing)
}

func (i *istioCsrProcessor) certsAreEqual(
	old, new map[string][]byte,
) bool {
	if len(old) != len(new) {
		return false
	}
	for oldKey, oldVal := range old {
		newVal, ok := new[oldKey]
		if !ok {
			return false
		}
		// 0 represents equality from this function
		if bytes.Compare(oldVal, newVal) != 0 {
			return false
		}
	}
	return true
}

func (i *istioCsrProcessor) ProcessDelete(
	ctx context.Context,
	csr *security_v1alpha1.MeshGroupCertificateSigningRequest,
) security_types.MeshGroupCertificateSigningRequestStatus {
	// TODO: handle deletion of mesh group certificate signing requests
	// https://github.com/solo-io/mesh-projects/issues/227
	return csr.Status
}
