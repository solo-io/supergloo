package csr_agent_controller

import (
	"context"

	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	securityv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	"github.com/solo-io/mesh-projects/pkg/logging"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	security_processors "github.com/solo-io/mesh-projects/services/common/processors/security"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type csrAgentHandler struct {
	ctx       context.Context
	csrClient zephyr_security.MeshGroupCertificateSigningRequestClient
	processor security_processors.MeshGroupCertificateSigningRequestProcessor
	predicate predicate.Predicate
}

func CsrControllerProviderLocal(
	mgr mc_manager.AsyncManager,
) (controller.MeshGroupCertificateSigningRequestController, error) {
	return controller.NewMeshGroupCertificateSigningRequestController("local-csr-controller", mgr.Manager())
}

func NewCsrAgentEventHandler(
	ctx context.Context,
	csrClient zephyr_security.MeshGroupCertificateSigningRequestClient,
	processor security_processors.MeshGroupCertificateSigningRequestProcessor,
	predicate predicate.Predicate,
) controller.MeshGroupCertificateSigningRequestEventHandler {
	return &csrAgentHandler{
		ctx:       ctx,
		csrClient: csrClient,
		processor: processor,
		predicate: predicate,
	}
}

func (c *csrAgentHandler) Create(obj *securityv1alpha1.MeshGroupCertificateSigningRequest) error {
	logger := logging.BuildEventLogger(c.ctx, logging.CreateEvent, obj)
	if !c.predicate.Create(event.CreateEvent{Meta: obj, Object: obj}) {
		logger.Debugw("skipping event")
		return nil
	}
	status := c.processor.ProcessCreate(c.ctx, obj)
	if status.GetComputedStatus().GetStatus() == core_types.ComputedStatus_INVALID ||
		status.GetComputedStatus().GetStatus() == core_types.ComputedStatus_PROCESSING_ERROR {
		logger.Debugw("error handling csr event", zap.Error(eris.New(status.GetComputedStatus().GetMessage())))
	}
	obj.Status = status
	err := c.csrClient.UpdateStatus(c.ctx, obj)
	if err != nil {
		logger.Errorw("error updating csr status", zap.Error(err))
	}
	return nil
}

func (c *csrAgentHandler) Update(old, new *securityv1alpha1.MeshGroupCertificateSigningRequest) error {
	logger := logging.BuildEventLogger(c.ctx, logging.UpdateEvent, new)
	if !c.predicate.Update(event.UpdateEvent{MetaOld: old, ObjectOld: old, MetaNew: new, ObjectNew: new}) {
		logger.Debugw("skipping event")
		return nil
	}
	status := c.processor.ProcessUpdate(c.ctx, old, new)
	if status.GetComputedStatus().GetStatus() == core_types.ComputedStatus_INVALID ||
		status.GetComputedStatus().GetStatus() == core_types.ComputedStatus_PROCESSING_ERROR {
		logger.Debugw("error handling csr event", zap.Error(eris.New(status.GetComputedStatus().GetMessage())))
	}
	new.Status = status
	err := c.csrClient.UpdateStatus(c.ctx, new)
	if err != nil {
		logger.Errorw("error updating csr status", zap.Error(err))
	}
	return nil
}

func (c *csrAgentHandler) Delete(obj *securityv1alpha1.MeshGroupCertificateSigningRequest) error {
	logging.BuildEventLogger(c.ctx, logging.DeleteEvent, obj).Debugf(UnexpectedEventMsg)
	return nil
}

func (c *csrAgentHandler) Generic(obj *securityv1alpha1.MeshGroupCertificateSigningRequest) error {
	logging.BuildEventLogger(c.ctx, logging.GenericEvent, obj).Debugf(UnexpectedEventMsg)
	return nil
}
