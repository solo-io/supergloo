package reconcilers

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/rotisserie/eris"
	istio_networking_clients "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	istio_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DestinationRuleReconciler interface {
	Reconcile(ctx context.Context, desiredGlobalState []*istio_networking.DestinationRule) error
}

// a reconciler can either be whole-cluster scoped or scoped to a namespace and/or labels
// either ScopedToWholeCluster() must be set or one of the following two methods must be called with non-zero values
// that's just to force it to be obvious when we're going to be reconciling EVERYTHING across a whole cluster
type DestinationRuleReconcilerBuilder interface {
	ScopedToWholeCluster() DestinationRuleReconcilerBuilder
	ScopedToNamespace(namespace string) DestinationRuleReconcilerBuilder
	ScopedToLabels(labels map[string]string) DestinationRuleReconcilerBuilder
	WithClient(client client.Client) DestinationRuleReconcilerBuilder

	Build() (DestinationRuleReconciler, error)
}

type destinationRuleReconcilerBuilder struct {
	clusterScoped bool
	namespace     string
	labels        map[string]string
	client        istio_networking_clients.DestinationRuleClient
}

func NewDestinationRuleReconcilerBuilder() DestinationRuleReconcilerBuilder {
	return &destinationRuleReconcilerBuilder{}
}

func (v *destinationRuleReconcilerBuilder) ScopedToWholeCluster() DestinationRuleReconcilerBuilder {
	v.clusterScoped = true
	return v
}

func (v *destinationRuleReconcilerBuilder) ScopedToNamespace(namespace string) DestinationRuleReconcilerBuilder {
	v.namespace = namespace
	return v
}

func (v *destinationRuleReconcilerBuilder) ScopedToLabels(labels map[string]string) DestinationRuleReconcilerBuilder {
	v.labels = labels
	return v
}

func (v *destinationRuleReconcilerBuilder) WithClient(client client.Client) DestinationRuleReconcilerBuilder {
	v.client = istio_networking_clients.NewDestinationRuleClient(client)
	return v
}

func (v *destinationRuleReconcilerBuilder) Build() (DestinationRuleReconciler, error) {
	if v.client == nil {
		return nil, eris.New("Must provide a client")
	}

	if !v.clusterScoped && v.namespace == "" && len(v.labels) == 0 {
		return nil, eris.New("Must either configure this reconciler to be cluster-scoped or explicitly scope it to a namespace/label")
	}

	if v.clusterScoped && (v.namespace != "" || len(v.labels) > 0) {
		return nil, eris.New("Cannot be cluster-scoped and scoped to a namespace/label")
	}

	return &destinationRuleReconciler{
		namespace: v.namespace,
		labels:    v.labels,
		client:    v.client,
	}, nil
}

// visible for testing
func NewDestinationRuleReconciler(namespace string, labels map[string]string, client istio_networking_clients.DestinationRuleClient) DestinationRuleReconciler {
	return &destinationRuleReconciler{
		namespace: namespace,
		labels:    labels,
		client:    client,
	}
}

type destinationRuleReconciler struct {
	client istio_networking_clients.DestinationRuleClient

	namespace string
	labels    map[string]string
}

func (v *destinationRuleReconciler) Reconcile(ctx context.Context, desiredGlobalState []*istio_networking.DestinationRule) error {
	existingObjList, err := v.client.ListDestinationRule(
		ctx,

		// if this reconciler has been scoped to the whole cluster, these two values will be their respective zero-values and this will list all the objects on the cluster
		client.InNamespace(v.namespace),
		client.MatchingLabels(v.labels),
	)
	if err != nil {
		return err
	}

	nameNamespaceToExistingState := map[string]*istio_networking.DestinationRule{}
	nameNamespaceToDesiredState := map[string]*istio_networking.DestinationRule{}

	for _, existingObjIter := range existingObjList.Items {
		existingObj := existingObjIter
		nameNamespaceToExistingState[selection.ToUniqueSingleClusterString(existingObj.ObjectMeta)] = &existingObj
	}

	for _, desiredObjIter := range desiredGlobalState {
		desiredObj := desiredObjIter
		nameNamespaceToDesiredState[selection.ToUniqueSingleClusterString(desiredObj.ObjectMeta)] = desiredObj
	}

	// update and delete existing objects
	for _, existingObj := range existingObjList.Items {
		desiredState, shouldBeAlive := nameNamespaceToDesiredState[selection.ToUniqueSingleClusterString(existingObj.ObjectMeta)]

		if !shouldBeAlive {
			err = v.client.DeleteDestinationRule(ctx, selection.ObjectMetaToObjectKey(existingObj.ObjectMeta))
			if err != nil {
				return err
			}
		} else if !proto.Equal(&existingObj.Spec, &desiredState.Spec) {
			// make sure we use the same resource version for updates
			desiredState.ObjectMeta.ResourceVersion = existingObj.ObjectMeta.ResourceVersion
			err = v.client.UpdateDestinationRule(ctx, desiredState)
			if err != nil {
				return err
			}
		}
	}

	// create new objects
	for _, desiredObj := range desiredGlobalState {
		_, isAlreadyExisting := nameNamespaceToExistingState[selection.ToUniqueSingleClusterString(desiredObj.ObjectMeta)]

		if !isAlreadyExisting {
			err := v.client.CreateDestinationRule(ctx, desiredObj)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
