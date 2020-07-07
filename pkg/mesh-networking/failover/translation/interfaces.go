package translation

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go -package mock_failover_service_translation

/*
	Given a FailoverService and a list of MeshServices corresponding to FailoverService.Spec.services,
	translate to mesh-specific configuration. Each mesh-specific translator will populate
	the relevant resources in the OutputSnapshot.
*/
type FailoverServiceTranslator interface {
	Translate(
		ctx context.Context,
		failoverService *smh_networking.FailoverService,
		prioritizedMeshServices []*v1alpha1.MeshService,
		allMeshes v1alpha1sets.MeshSet,
	) (failover.MeshOutputs, *smh_networking_types.FailoverServiceStatus_TranslatorError)
}
