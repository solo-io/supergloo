package controller

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/core"
	"github.com/solo-io/mesh-projects/services/common"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	MeshGroupProviderSet = wire.NewSet(
		zephyr_core.NewMeshClient,
		MeshGroupValidatorProvider,
		MeshGroupEventHandlerProvider,
	)
)

func NewMeshGroupControllerStarter(
	handler controller.MeshGroupEventHandler,
	predicates ...predicate.Predicate) mc_manager.AsyncManagerStartOptionsFunc {
	return func(ctx context.Context, mgr manager.Manager) error {
		ctrl, err := controller.NewMeshGroupController(common.LocalClusterName, mgr)
		if err != nil {
			return err
		}
		if err = ctrl.AddEventHandler(ctx, handler, predicates...); err != nil {
			return err
		}
		return nil
	}
}

func MeshGroupEventHandlerProvider(ctx context.Context, validator MeshGroupValidator) controller.MeshGroupEventHandler {
	return &meshGroupEventHandler{
		ctx:       ctx,
		validator: validator,
	}
}

type meshGroupEventHandler struct {
	validator MeshGroupValidator
	ctx       context.Context
}

func (m *meshGroupEventHandler) Create(obj *v1alpha1.MeshGroup) error {
	resync, err := m.validate(obj)
	if err != nil {
		contextutils.LoggerFrom(m.ctx).Errorw("error handling create event for mesh group", zap.Error(err))
		if resync {
			return err
		}
	}
	return nil
}

func (m *meshGroupEventHandler) Update(old, new *v1alpha1.MeshGroup) error {
	resync, err := m.validate(new)
	if err != nil {
		contextutils.LoggerFrom(m.ctx).Errorw("error handling update event for mesh group", zap.Error(err))
		if resync {
			return err
		}
	}
	return nil
}

func (m *meshGroupEventHandler) Delete(obj *v1alpha1.MeshGroup) error {
	// for now do nothing
	contextutils.LoggerFrom(m.ctx).Errorf("MeshGroup %s.%s deleted, termporarily not handling this case",
		obj.GetName(), obj.GetNamespace())
	return nil
}

func (m *meshGroupEventHandler) Generic(obj *v1alpha1.MeshGroup) error {
	contextutils.LoggerFrom(m.ctx).Errorf("the generic event should never be called for mesh groups, "+
		"it was called with %s", obj.ObjectMeta.String())
	return nil
}

/*
	If the error is a network error, or something related, we should attempt to resync.
	If it is an error of our own creation we should not
*/
func (m *meshGroupEventHandler) validate(obj *v1alpha1.MeshGroup) (resync bool, err error) {
	// TODO: Write this status to the object, skipping for now as a reconciler is probably in order
	status, err := m.validator.Validate(m.ctx, obj)

	if err != nil {
		if status.Config == types.MeshGroupStatus_PROCESSING_ERROR {
			return true, err
		}
		return false, err
	}
	return false, nil

}
