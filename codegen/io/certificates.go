package io

import (
	"github.com/solo-io/service-mesh-hub/codegen/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	CertificateIssuerInputTypes = Snapshot{
		schema.GroupVersion{
			Group:   "certificates." + constants.ServiceMeshHubApiGroupSuffix,
			Version: "v1alpha2",
		}: {
			"IssuedCertificate",
			"CertificateRequest",
		},
	}

	CertificateAgentInputTypes = Snapshot{
		schema.GroupVersion{
			Group:   "certificates." + constants.ServiceMeshHubApiGroupSuffix,
			Version: "v1alpha2",
		}: {
			"IssuedCertificate",
			"CertificateRequest",
		},
		corev1.SchemeGroupVersion: {
			"Secret",
			"Pod",
		},
	}

	CertificateAgentOutputTypes = OutputSnapshot{
		Name: "certagent",
		Snapshot: Snapshot{
			schema.GroupVersion{
				Group:   "certificates." + constants.ServiceMeshHubApiGroupSuffix,
				Version: "v1alpha2",
			}: {
				"CertificateRequest",
			},
			corev1.SchemeGroupVersion: {
				"Secret",
			},
		},
	}
)
