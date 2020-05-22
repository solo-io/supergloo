package reconcilers

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/rotisserie/eris"
	istio_networking_clients "github.com/solo-io/service-mesh-hub/pkg/api/istio/networking/v1alpha3"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	istio_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VirtualServiceReconciler interface {
	Reconcile(ctx context.Context, desiredGlobalState []*istio_networking.VirtualService) error
}

// a reconciler can either be whole-cluster scoped or scoped to a namespace and/or labels
// either ScopedToWholeCluster() must be set or one of the following two methods must be called with non-zero values
// that's just to force it to be obvious when we're going to be reconciling EVERYTHING across a whole cluster
type VirtualServiceReconcilerBuilder interface {
	ScopedToWholeCluster() VirtualServiceReconcilerBuilder
	ScopedToNamespace(namespace string) VirtualServiceReconcilerBuilder
	ScopedToLabels(labels map[string]string) VirtualServiceReconcilerBuilder
	WithClient(client client.Client) VirtualServiceReconcilerBuilder

	Build() (VirtualServiceReconciler, error)
}

type virtualServiceReconcilerBuilder struct {
	clusterScoped bool
	namespace     string
	labels        map[string]string
	client        istio_networking_clients.VirtualServiceClient
}

func NewVirtualServiceReconcilerBuilder() VirtualServiceReconcilerBuilder {
	return &virtualServiceReconcilerBuilder{}
}

func (v *virtualServiceReconcilerBuilder) ScopedToWholeCluster() VirtualServiceReconcilerBuilder {
	v.clusterScoped = true
	return v
}

func (v *virtualServiceReconcilerBuilder) ScopedToNamespace(namespace string) VirtualServiceReconcilerBuilder {
	v.namespace = namespace
	return v
}

func (v *virtualServiceReconcilerBuilder) ScopedToLabels(labels map[string]string) VirtualServiceReconcilerBuilder {
	v.labels = labels
	return v
}

func (v *virtualServiceReconcilerBuilder) WithClient(client client.Client) VirtualServiceReconcilerBuilder {
	v.client = istio_networking_clients.NewVirtualServiceClient(client)
	return v
}

func (v *virtualServiceReconcilerBuilder) Build() (VirtualServiceReconciler, error) {
	if v.client == nil {
		return nil, eris.New("Must provide a client")
	}

	if !v.clusterScoped && v.namespace == "" && len(v.labels) == 0 {
		return nil, eris.New("Must either configure this reconciler to be cluster-scoped or explicitly scope it to a namespace/label")
	}

	if v.clusterScoped && (v.namespace != "" || len(v.labels) > 0) {
		return nil, eris.New("Cannot be cluster-scoped and scoped to a namespace/label")
	}

	return &virtualServiceReconciler{
		namespace:            v.namespace,
		labels:               v.labels,
		virtualServiceClient: v.client,
	}, nil
}

type virtualServiceReconciler struct {
	virtualServiceClient istio_networking_clients.VirtualServiceClient

	namespace string
	labels    map[string]string
}

func (v *virtualServiceReconciler) Reconcile(ctx context.Context, desiredGlobalState []*istio_networking.VirtualService) error {
	virtualServiceList, err := v.virtualServiceClient.ListVirtualService(
		ctx,

		// if this reconciler has been scoped to the whole cluster, these two values will be their respective zero-values and this will list all the objects on the cluster
		client.InNamespace(v.namespace),
		client.MatchingLabels(v.labels),
	)
	if err != nil {
		return err
	}

	nameNamespaceToExistingState := map[string]*istio_networking.VirtualService{}
	nameNamespaceToDesiredState := map[string]*istio_networking.VirtualService{}

	for _, existingVsIter := range virtualServiceList.Items {
		existingVs := existingVsIter
		nameNamespaceToExistingState[clients.ToUniqueSingleClusterString(existingVs.ObjectMeta)] = &existingVs
	}

	for _, desiredVsIter := range desiredGlobalState {
		desiredVs := desiredVsIter
		nameNamespaceToDesiredState[clients.ToUniqueSingleClusterString(desiredVs.ObjectMeta)] = desiredVs
	}

	// update and delete existing VS's
	for _, existingVirtualService := range virtualServiceList.Items {
		desiredState, shouldBeAlive := nameNamespaceToDesiredState[clients.ToUniqueSingleClusterString(existingVirtualService.ObjectMeta)]

		if !shouldBeAlive {
			err = v.virtualServiceClient.DeleteVirtualService(ctx, clients.ObjectMetaToObjectKey(existingVirtualService.ObjectMeta))
			if err != nil {
				return err
			}
		} else if !proto.Equal(&existingVirtualService.Spec, &desiredState.Spec) {
			err = v.virtualServiceClient.UpdateVirtualService(ctx, desiredState)
			if err != nil {
				return err
			}
		}
	}

	// create new VS's
	for _, desiredVirtualService := range desiredGlobalState {
		_, isAlreadyExisting := nameNamespaceToExistingState[clients.ToUniqueSingleClusterString(desiredVirtualService.ObjectMeta)]

		if !isAlreadyExisting {
			err := v.virtualServiceClient.CreateVirtualService(ctx, desiredVirtualService)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
