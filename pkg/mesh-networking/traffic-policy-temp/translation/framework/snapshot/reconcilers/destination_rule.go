package reconcilers

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	istio_networking_clients "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	istio_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DestinationRuleReconciler interface {
	Reconcile(ctx context.Context, desiredGlobalState []*istio_networking.DestinationRule) error
}

// a reconciler can either be whole-cluster scoped or scoped to a namespace. In addition labels can be used.
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

	if v.clusterScoped && v.namespace != "" {
		return nil, eris.New("Cannot be cluster-scoped and scoped to a namespace")
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
	logger := contextutils.LoggerFrom(ctx)
	logger.Debug("reconciling destination rules")
	existingObjList, err := v.client.ListDestinationRule(
		ctx,

		// if this reconciler has been scoped to the whole cluster, these two values will be their respective zero-values and this will list all the objects on the cluster
		client.InNamespace(v.namespace),
		client.MatchingLabels(v.labels),
	)
	if err != nil {
		return err
	}

	nameNamespaceToDesiredState := map[string]*istio_networking.DestinationRule{}

	for _, desiredObjIter := range desiredGlobalState {
		desiredObj := desiredObjIter
		nameNamespaceToDesiredState[selection.ToUniqueSingleClusterString(desiredObj.ObjectMeta)] = desiredObj
	}
	var multierr error

	// update and delete existing objects
	for _, existingObj := range existingObjList.Items {
		key := selection.ToUniqueSingleClusterString(existingObj.ObjectMeta)
		desiredState, shouldBeAlive := nameNamespaceToDesiredState[key]
		delete(nameNamespaceToDesiredState, key)
		if !shouldBeAlive {
			logger.Debugw("deleting destination rule", "ref", existingObj.ObjectMeta)
			err = v.client.DeleteDestinationRule(ctx, selection.ObjectMetaToObjectKey(existingObj.ObjectMeta))
			if err != nil {
				logger.Warnw("error deleting destination rule", "error", err, "ref", existingObj.ObjectMeta)
				multierr = multierror.Append(multierr, err)
			}
		} else if !proto.Equal(&existingObj.Spec, &desiredState.Spec) {
			// make sure we use the same resource version for updates
			desiredState.ObjectMeta.ResourceVersion = existingObj.ObjectMeta.ResourceVersion
			v.addLabels(desiredState)

			logger.Debugw("updating destination rule", "ref", existingObj.ObjectMeta)
			err = v.client.UpdateDestinationRule(ctx, desiredState)
			if err != nil {
				logger.Warnw("error updating destination rule", "error", err, "ref", existingObj.ObjectMeta)
				multierr = multierror.Append(multierr, err)
			}
		}
	}

	// create new objects of what's left in the map
	for _, desiredObj := range nameNamespaceToDesiredState {
		logger.Debugw("creating destination rule", "ref", desiredObj.ObjectMeta)

		// add our labels:
		v.addLabels(desiredObj)

		err := v.client.CreateDestinationRule(ctx, desiredObj)
		if err != nil {
			logger.Warnw("error creating destination rule", "error", err, "ref", desiredObj.ObjectMeta)
			multierr = multierror.Append(multierr, err)
		}
	}

	return multierr
}

func (v *destinationRuleReconciler) addLabels(desiredDestRule *istio_networking.DestinationRule) {
	if desiredDestRule.Labels == nil && len(v.labels) != 0 {
		desiredDestRule.Labels = make(map[string]string)
	}
	for k, v := range v.labels {
		desiredDestRule.Labels[k] = v
	}
}
