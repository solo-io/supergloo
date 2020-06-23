package failover

import (
	"context"

	"github.com/hashicorp/go-multierror"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/reconcile"
)

type failoverServiceReconciler struct {
	ctx                      context.Context
	failoverServiceProcessor FailoverServiceProcessor
	failoverServiceClient    smh_networking.FailoverServiceClient
	meshServiceClient        smh_discovery.MeshServiceClient
}

type InputSnapshot struct {
	FailoverServices []*smh_networking.FailoverService
	MeshServices     []*smh_discovery.MeshService
}

type OutputSnapshot struct {
	FailoverServices []*smh_networking.FailoverService
}

func (f *failoverServiceReconciler) ReconcileFailoverService(obj *smh_networking.FailoverService) (reconcile.Result, error) {
	inputSnapshot, err := f.buildInputSnapshot()
	if err != nil {
		return reconcile.Result{}, err
	}
	outputSnapshot := f.failoverServiceProcessor.Process(f.ctx, inputSnapshot)
	// Update status on all FailoverServices
	// TODO should we retry here if status update fails?
	return reconcile.Result{}, f.updateFailoverServiceStatuses(outputSnapshot)
}

func (f *failoverServiceReconciler) buildInputSnapshot() (InputSnapshot, error) {
	inputSnapshot := InputSnapshot{}
	failoverServiceList, err := f.failoverServiceClient.ListFailoverService(f.ctx)
	if err != nil {
		return InputSnapshot{}, err
	}
	for _, failoverService := range failoverServiceList.Items {
		failoverService := failoverService
		inputSnapshot.FailoverServices = append(inputSnapshot.FailoverServices, &failoverService)
	}
	meshServiceList, err := f.meshServiceClient.ListMeshService(f.ctx)
	if err != nil {
		return InputSnapshot{}, err
	}
	for _, meshService := range meshServiceList.Items {
		meshService := meshService
		inputSnapshot.MeshServices = append(inputSnapshot.MeshServices, &meshService)
	}
	return inputSnapshot, nil
}

func (f *failoverServiceReconciler) updateFailoverServiceStatuses(
	outputSnapshot OutputSnapshot,
) error {
	var multierr *multierror.Error
	for _, failoverService := range outputSnapshot.FailoverServices {
		err := f.failoverServiceClient.UpdateFailoverServiceStatus(f.ctx, failoverService)
		if err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}
	return multierr.ErrorOrNil()
}
