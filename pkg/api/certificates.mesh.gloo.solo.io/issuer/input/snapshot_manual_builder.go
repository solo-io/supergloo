// Code generated by skv2. DO NOT EDIT.



/*
	Utility for manually building input snapshots. Used primarily in tests.
*/
package input

import (
	
	
	certificates_mesh_gloo_solo_io_v1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1alpha2"
	certificates_mesh_gloo_solo_io_v1alpha2_sets "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1alpha2/sets"

)

type InputSnapshotManualBuilder struct {
name string
	
		issuedCertificates certificates_mesh_gloo_solo_io_v1alpha2_sets.IssuedCertificateSet
		certificateRequests certificates_mesh_gloo_solo_io_v1alpha2_sets.CertificateRequestSet

}

func NewInputSnapshotManualBuilder(name string) *InputSnapshotManualBuilder {
return &InputSnapshotManualBuilder{
name:               name,
	
		issuedCertificates: certificates_mesh_gloo_solo_io_v1alpha2_sets.NewIssuedCertificateSet(),
		certificateRequests: certificates_mesh_gloo_solo_io_v1alpha2_sets.NewCertificateRequestSet(),
}
}

func (i *InputSnapshotManualBuilder) Build() Snapshot {
return NewSnapshot(
i.name,
	
		i.issuedCertificates,
		i.certificateRequests,
)
}
		func (i *InputSnapshotManualBuilder) AddIssuedCertificates(issuedCertificates []*certificates_mesh_gloo_solo_io_v1alpha2.IssuedCertificate) *InputSnapshotManualBuilder {
		i.issuedCertificates.Insert(issuedCertificates...)
		return i
		}
		func (i *InputSnapshotManualBuilder) AddCertificateRequests(certificateRequests []*certificates_mesh_gloo_solo_io_v1alpha2.CertificateRequest) *InputSnapshotManualBuilder {
		i.certificateRequests.Insert(certificateRequests...)
		return i
		}
