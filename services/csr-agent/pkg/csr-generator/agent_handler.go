package csr_generator

import (
	"context"

	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	securityv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	"github.com/solo-io/mesh-projects/pkg/logging"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	"go.uber.org/zap"
)

func CsrControllerProviderLocal(
	mgr mc_manager.AsyncManager,
) (controller.MeshGroupCertificateSigningRequestController, error) {
	return controller.NewMeshGroupCertificateSigningRequestController("local-csr-controller", mgr.Manager())
}

func NewMeshGroupCSRDataSourceFactory() MeshGroupCSRDataSourceFactory {
	return NewMeshGroupCSRDataSource
}

func NewMeshGroupCSRDataSource(
	ctx context.Context,
	csrClient zephyr_security.MeshGroupCSRClient,
	processor MeshGroupCSRProcessor,
) controller.MeshGroupCertificateSigningRequestEventHandler {
	return &controller.MeshGroupCertificateSigningRequestEventHandlerFuncs{
		OnCreate: func(obj *securityv1alpha1.MeshGroupCertificateSigningRequest) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, obj)
			status := processor.ProcessUpsert(ctx, obj)
			if status == nil {
				logger.Debugw("csr event was not processed")
				return nil
			}
			if status.GetComputedStatus().GetStatus() == core_types.ComputedStatus_INVALID ||
				status.GetComputedStatus().GetStatus() == core_types.ComputedStatus_PROCESSING_ERROR {
				logger.Debugw("error handling csr event", zap.Error(eris.New(status.GetComputedStatus().GetMessage())))
			}
			obj.Status = *status
			err := csrClient.UpdateStatus(ctx, obj)
			if err != nil {
				logger.Errorw("error updating csr status", zap.Error(err))
			}
			return nil
		},
		OnUpdate: func(_, new *securityv1alpha1.MeshGroupCertificateSigningRequest) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, new)
			status := processor.ProcessUpsert(ctx, new)
			if status == nil {
				logger.Debugw("csr event was not processed")
				return nil
			}
			if status.GetComputedStatus().GetStatus() == core_types.ComputedStatus_INVALID ||
				status.GetComputedStatus().GetStatus() == core_types.ComputedStatus_PROCESSING_ERROR {
				logger.Debugw("error handling csr event", zap.Error(eris.New(status.GetComputedStatus().GetMessage())))
			}
			new.Status = *status
			err := csrClient.UpdateStatus(ctx, new)
			if err != nil {
				logger.Errorw("error updating csr status", zap.Error(err))
			}
			return nil
		},
		OnDelete: func(obj *securityv1alpha1.MeshGroupCertificateSigningRequest) error {
			logging.BuildEventLogger(ctx, logging.DeleteEvent, obj).Debugf(UnexpectedEventMsg)
			return nil
		},
		OnGeneric: func(obj *securityv1alpha1.MeshGroupCertificateSigningRequest) error {
			logging.BuildEventLogger(ctx, logging.GenericEvent, obj).Debugf(UnexpectedEventMsg)
			return nil
		},
	}
}
