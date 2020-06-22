package failover

import (
	"context"

	"github.com/hashicorp/go-multierror"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/reconcile"
)

type failoverServiceReconciler struct {
	ctx                   context.Context
	failoverServiceClient smh_networking.FailoverServiceClient
	virtualMeshClient     smh_networking.VirtualMeshClient
}

func (f *failoverServiceReconciler) ReconcileFailoverService(obj *smh_networking.FailoverService) (reconcile.Result, error) {
	panic("implement me")
}

func (f *failoverServiceReconciler) process(ctx context.Context) error {
	failoverServiceList, err := f.failoverServiceClient.ListFailoverService(ctx)
	if err != nil {
		return err
	}
	var multiErr *multierror.Error
	for _, failoverService := range failoverServiceList.Items {
		failoverService := failoverService
		err := f.processFailoverService(&failoverService)
		failoverService.Status = f.computeStatus(err)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	// Update status on all FailoverServices
	return multiErr.ErrorOrNil()
}

func (f *failoverServiceReconciler) processFailoverService(
	failoverService *smh_networking.FailoverService,
) error {

}

//func (f *failoverServiceReconciler) computeStatus(err error) types.FailoverServiceStatus {
//	var status types.FailoverServiceStatus
//	if err != nil {
//		status = types.FailoverServiceStatus{
//			ProcessingStatus: &smh_core_types.Status{
//				State:   smh_core_types.Status_PROCESSING_ERROR,
//				Message: err.Error(),
//			},
//		}
//	} else {
//		status = types.FailoverServiceStatus{
//			ProcessingStatus: &smh_core_types.Status{
//				State: smh_core_types.Status_ACCEPTED,
//			},
//		}
//	}
//	return status
//}
