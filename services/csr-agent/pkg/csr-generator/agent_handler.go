package csr_generator

import (
	"context"

	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	zephyr_security_controller "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	"go.uber.org/zap"
)

func CsrControllerProviderLocal(
	mgr mc_manager.AsyncManager,
) zephyr_security_controller.VirtualMeshCertificateSigningRequestEventWatcher {
	return zephyr_security_controller.NewVirtualMeshCertificateSigningRequestEventWatcher("local-csr-controller", mgr.Manager())
}

func NewVirtualMeshCSRDataSourceFactory() VirtualMeshCSRDataSourceFactory {
	return NewVirtualMeshCSRDataSource
}

func NewVirtualMeshCSRDataSource(
	ctx context.Context,
	csrClient zephyr_security.VirtualMeshCertificateSigningRequestClient,
	processor VirtualMeshCSRProcessor,
) zephyr_security_controller.VirtualMeshCertificateSigningRequestEventHandler {
	return &zephyr_security_controller.VirtualMeshCertificateSigningRequestEventHandlerFuncs{
		OnCreate: func(obj *zephyr_security.VirtualMeshCertificateSigningRequest) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.CreateEvent, obj)
			status := processor.ProcessUpsert(ctx, obj)
			if status == nil {
				logger.Debugw("csr event was not processed")
				return nil
			}
			if status.GetComputedStatus().GetState() == zephyr_core_types.Status_INVALID ||
				status.GetComputedStatus().GetState() == zephyr_core_types.Status_PROCESSING_ERROR {
				logger.Debugw("error handling csr event", zap.Error(eris.New(status.GetComputedStatus().GetMessage())))
			}
			obj.Status = *status
			err := csrClient.UpdateVirtualMeshCertificateSigningRequestStatus(ctx, obj)
			if err != nil {
				logger.Errorw("error updating csr status", zap.Error(err))
			}
			return nil
		},
		OnUpdate: func(_, new *zephyr_security.VirtualMeshCertificateSigningRequest) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.UpdateEvent, new)
			status := processor.ProcessUpsert(ctx, new)
			if status == nil {
				logger.Debugw("csr event was not processed")
				return nil
			}
			if status.GetComputedStatus().GetState() == zephyr_core_types.Status_INVALID ||
				status.GetComputedStatus().GetState() == zephyr_core_types.Status_PROCESSING_ERROR {
				logger.Debugw("error handling csr event", zap.Error(eris.New(status.GetComputedStatus().GetMessage())))
			}
			new.Status = *status
			err := csrClient.UpdateVirtualMeshCertificateSigningRequestStatus(ctx, new)
			if err != nil {
				logger.Errorw("error updating csr status", zap.Error(err))
			}
			return nil
		},
		OnDelete: func(obj *zephyr_security.VirtualMeshCertificateSigningRequest) error {
			container_runtime.BuildEventLogger(ctx, container_runtime.DeleteEvent, obj).Debugf(UnexpectedEventMsg)
			return nil
		},
		OnGeneric: func(obj *zephyr_security.VirtualMeshCertificateSigningRequest) error {
			container_runtime.BuildEventLogger(ctx, container_runtime.GenericEvent, obj).Debugf(UnexpectedEventMsg)
			return nil
		},
	}
}
