package acp_translator

import (
	"context"

	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type AcpTranslatorLoop interface {
	Start(ctx context.Context) error
}

type AcpMeshTranslator interface {
	// Translate the given AccessControlPolicy applying to targetServices to mesh-specific configuration
	Translate(
		ctx context.Context,
		targetServices []TargetService,
		acp *networking_v1alpha1.AccessControlPolicy,
	) *networking_types.AccessControlPolicyStatus_TranslatorError
}
