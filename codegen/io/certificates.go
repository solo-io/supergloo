package io

import (
	"github.com/solo-io/gloo-mesh/codegen/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	CertificateIssuerInputTypes = Snapshot{
		schema.GroupVersion{
			Group:   "certificates." + constants.GlooMeshApiGroupSuffix,
			Version: "v1",
		}: {
			"IssuedCertificate",
			"CertificateRequest",
		},
	}

	CertificateAgentInputTypes = Snapshot{
		schema.GroupVersion{
			Group:   "certificates." + constants.GlooMeshApiGroupSuffix,
			Version: "v1",
		}: {
			"IssuedCertificate",
			"CertificateRequest",
			"PodBounceDirective",
		},
		corev1.SchemeGroupVersion: {
			"Secret",
			"Pod",
			"ConfigMap",
			"ServiceAccount",
		},
	}

	CertificateAgentOutputTypes = OutputSnapshot{
		Name: "certagent",
		Snapshot: Snapshot{
			schema.GroupVersion{
				Group:   "certificates." + constants.GlooMeshApiGroupSuffix,
				Version: "v1",
			}: {
				"CertificateRequest",
			},
			corev1.SchemeGroupVersion: {
				"Secret",
			},
		},
	}
)
