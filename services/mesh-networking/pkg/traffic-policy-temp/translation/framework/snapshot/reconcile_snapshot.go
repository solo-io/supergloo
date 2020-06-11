package snapshot

import (
	"context"

	istio_networking "github.com/solo-io/service-mesh-hub/pkg/api/istio/networking/v1alpha3"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot/reconcilers"
)

type snapshotReconciler struct {
	dynamicClientGetter          multicluster.DynamicClientGetter
	virtualServiceClientFactory  istio_networking.VirtualServiceClientFactory
	destinationRuleClientFactory istio_networking.DestinationRuleClientFactory

	virtualServiceReconciler  reconcilers.VirtualServiceReconciler
	destinationRuleReconciler reconcilers.DestinationRuleReconciler
}

func (r *snapshotReconciler) ReconcileAllSnapshots(ctx context.Context, clusterNameToSnapshot ClusterNameToSnapshot) error {

	for cluster, snapshot := range clusterNameToSnapshot {
		client, err := r.dynamicClientGetter.GetClientForCluster(ctx, cluster.Name)
		if err != nil {
			return
			// todo continue
		}
		if snapshot.Istio != nil {

			vsclient, err := r.virtualServiceClientFactory(client)
			if err != nil {
				return
				// todo continue
			}
			destRules, err := r.destinationRuleClientFactory(client)
			if err != nil {
				return
				// todo continue
			}

			err = virtualServiceReconciler.Reconcile(ctx, snapshot.Istio.VirtualServices)
			if err != nil {
				return
				// todo continue
			}
			err = destinationRuleReconciler.Reconcile(ctx, snapshot.Istio.DestinationRules)
			if err != nil {
				return
				// todo continue
			}

		}

	}

}
