package mc_watcher

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	compute_target "github.com/solo-io/service-mesh-hub/services/common/compute-target"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	internal_watcher "github.com/solo-io/service-mesh-hub/services/common/compute-target/secret-event-handler/internal"
	core_v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

/*
	The following function is meant to serve as the secret watcher which will serve as the "entrypoint"
	for multi cluster applications. It will read in kube secrets with the `solo.io/kubeconfig` label
	and translate them into kube configs. The input to this function is the handler interface which
	will serve as the callbacks which receive the cluster config as they are added and deleted. This
	function is meant to be paired with the `AsyncManagerController` located in the `mcutils` package.

	The `Start()` function on the `AsyncManagerController` takes as an argument optional start functions,
	which this function returns. That way ensuring that this code is called anytime multi cluster watchers
	are necessary
*/
func StartLocalManager(
	computeTargetCredentialsHandlers []compute_target.ComputeTargetCredentialsHandler,
) mc_manager.AsyncManagerStartOptionsFunc {
	return func(ctx context.Context, mgr manager.Manager) error {
		secretCtrl := controller.NewSecretEventWatcher(mc_manager.MultiClusterController, mgr)

		mcHandler := &computeTargetHandler{
			ctx:                        ctx,
			computeTargetSecretHandler: internal_watcher.NewComputeTargetMembershipHandler(computeTargetCredentialsHandlers),
		}

		if err := secretCtrl.AddEventHandler(ctx, mcHandler, &internal_watcher.MultiClusterPredicate{}); err != nil {
			return err
		}
		return nil
	}
}

type computeTargetHandler struct {
	ctx                        context.Context
	computeTargetSecretHandler internal_watcher.ComputeTargetSecretHandler
}

func NewComputeTargetHandler(ctx context.Context,
	computeTargetSecretHandler internal_watcher.ComputeTargetSecretHandler,
) controller.SecretEventHandler {
	return &computeTargetHandler{ctx: ctx, computeTargetSecretHandler: computeTargetSecretHandler}
}

func (c *computeTargetHandler) CreateSecret(s *core_v1.Secret) error {
	resync, err := c.computeTargetSecretHandler.ComputeTargetSecretAdded(c.ctx, s)
	if err != nil {
		contextutils.LoggerFrom(c.ctx).Errorf("error adding member mesh API for secret %s.%s: %s", s.GetName(), s.GetNamespace(), err.Error())
		if resync {
			return err
		}
	}
	return nil
}

func (c *computeTargetHandler) UpdateSecret(old, new *core_v1.Secret) error {
	logger := contextutils.LoggerFrom(c.ctx)
	// If mc label has been removed from secret, remove from remote clusters
	if internal_watcher.HasRequiredMetadata(old.GetObjectMeta(), old) && !internal_watcher.HasRequiredMetadata(new.GetObjectMeta(), new) {
		resync, err := c.computeTargetSecretHandler.ComputeTargetSecretRemoved(c.ctx, new)
		if err != nil {
			logger.Errorf("error deleting member cluster for "+
				"secret %s.%s", new.GetName(), new.GetNamespace())
			if resync {
				return err
			}
		}
	}
	// if mc label has been added to secret, add to remote cluster list
	if !internal_watcher.HasRequiredMetadata(old.GetObjectMeta(), old) && internal_watcher.HasRequiredMetadata(new.GetObjectMeta(), new) {
		resync, err := c.computeTargetSecretHandler.ComputeTargetSecretAdded(c.ctx, new)
		if err != nil {
			logger.Errorf("error adding member cluster for "+
				"secret %s.%s", new.GetName(), new.GetNamespace())
			if resync {
				return err
			}
		}
	}
	return nil
}

func (c *computeTargetHandler) DeleteSecret(s *core_v1.Secret) error {
	resync, err := c.computeTargetSecretHandler.ComputeTargetSecretRemoved(c.ctx, s)
	if err != nil {
		contextutils.LoggerFrom(c.ctx).Errorf("error deleting member cluster for "+
			"secret %s.%s", s.GetName(), s.GetNamespace())
		if resync {
			return err
		}
	}
	return nil
}

func (c *computeTargetHandler) GenericSecret(obj *core_v1.Secret) error {
	contextutils.LoggerFrom(c.ctx).Warn("should never be called as generic events are not configured for secrets")
	return nil
}
