package acp_translator

import (
	"context"

	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type AcpTranslatorLoop interface {
	Start(ctx context.Context) error
}

type AcpMeshTranslator interface {
	// The name which will be used to identify the translator in the logs
	Name() string
	// Translate the given AccessControlPolicy applying to targetServices to mesh-specific configuration
	Translate(
		ctx context.Context,
		targetServices []TargetService,
		acp *smh_networking.AccessControlPolicy,
	) *networking_types.AccessControlPolicyStatus_TranslatorError
}
