package failover

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

/*
	Given a FailoverService and a list of MeshServices corresponding to FailoverService.Spec.services,
	translate to mesh-specific configuration.
*/
type FailoverServiceTranslator interface {
	Translate(
		ctx context.Context,
		failoverService *smh_networking.FailoverService,
		prioritizedMeshServices []*v1alpha1.MeshService,
	) *smh_networking_types.FailoverServiceStatus_TranslatorError
}
