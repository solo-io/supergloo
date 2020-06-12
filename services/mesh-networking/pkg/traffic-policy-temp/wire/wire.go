// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	reconcile "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MeshServiceReaderProvider(client client.Client) smh_discovery.MeshServiceReader {
	return smh_discovery.MeshServiceClientProvider(client)
}
func MeshWorkloadReaderProvider(client client.Client) smh_discovery.MeshWorkloadReader {
	return smh_discovery.MeshWorkloadClientProvider(client)
}

func InitializeReconciler(ctx context.Context) (*reconcile.Reconciler, error) {
	wire.Build(
		smh_discovery.MeshClientProvider,
		MeshServiceReaderProvider,
		smh_discovery.MeshServiceClientProvider,
		smh_discovery.MeshWorkloadClientProvider,
		MeshWorkloadReaderProvider,
		smh_networking.TrafficPolicyClientProvider,

		reconcile.NewReconciler,
	)

	return new(reconcile.Reconciler), nil
}
