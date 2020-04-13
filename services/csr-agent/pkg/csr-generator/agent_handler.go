package csr_generator

import (
	"context"

	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	securityv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	"go.uber.org/zap"
)

func CsrControllerProviderLocal(
	mgr mc_manager.AsyncManager,
) controller.VirtualMeshCertificateSigningRequestEventWatcher {
	return controller.NewVirtualMeshCertificateSigningRequestEventWatcher("local-csr-controller", mgr.Manager())
}

func NewVirtualMeshCSRDataSourceFactory() VirtualMeshCSRDataSourceFactory {
	return NewVirtualMeshCSRDataSource
}

func NewVirtualMeshCSRDataSource(
	ctx context.Context,
	csrClient zephyr_security.VirtualMeshCertificateSigningRequestClient,
	processor VirtualMeshCSRProcessor,
) controller.VirtualMeshCertificateSigningRequestEventHandler {
	return &controller.VirtualMeshCertificateSigningRequestEventHandlerFuncs{
		OnCreate: func(obj *securityv1alpha1.VirtualMeshCertificateSigningRequest) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, obj)
			status := processor.ProcessUpsert(ctx, obj)
			if status == nil {
				logger.Debugw("csr event was not processed")
				return nil
			}
			if status.GetComputedStatus().GetState() == core_types.Status_INVALID ||
				status.GetComputedStatus().GetState() == core_types.Status_PROCESSING_ERROR {
				logger.Debugw("error handling csr event", zap.Error(eris.New(status.GetComputedStatus().GetMessage())))
			}
			obj.Status = *status
			err := csrClient.UpdateVirtualMeshCertificateSigningRequestStatus(ctx, obj)
			if err != nil {
				logger.Errorw("error updating csr status", zap.Error(err))
			}
			return nil
		},
		OnUpdate: func(_, new *securityv1alpha1.VirtualMeshCertificateSigningRequest) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, new)
			status := processor.ProcessUpsert(ctx, new)
			if status == nil {
				logger.Debugw("csr event was not processed")
				return nil
			}
			if status.GetComputedStatus().GetState() == core_types.Status_INVALID ||
				status.GetComputedStatus().GetState() == core_types.Status_PROCESSING_ERROR {
				logger.Debugw("error handling csr event", zap.Error(eris.New(status.GetComputedStatus().GetMessage())))
			}
			new.Status = *status
			err := csrClient.UpdateVirtualMeshCertificateSigningRequestStatus(ctx, new)
			if err != nil {
				logger.Errorw("error updating csr status", zap.Error(err))
			}
			return nil
		},
		OnDelete: func(obj *securityv1alpha1.VirtualMeshCertificateSigningRequest) error {
			logging.BuildEventLogger(ctx, logging.DeleteEvent, obj).Debugf(UnexpectedEventMsg)
			return nil
		},
		OnGeneric: func(obj *securityv1alpha1.VirtualMeshCertificateSigningRequest) error {
			logging.BuildEventLogger(ctx, logging.GenericEvent, obj).Debugf(UnexpectedEventMsg)
			return nil
		},
	}
}
