package multicluster

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// register the mesh projects v1alpha1 CRDs with manager runtime
var AddSchemeV1Alpha1 mc_manager.AsyncManagerStartOptionsFunc = func(_ context.Context, mgr manager.Manager) error {
	err := v1alpha1.AddToScheme(mgr.GetScheme())
	if err != nil {
		return eris.Wrap(err, "failed to register v1alpha1 CRDs with the manager runtime")
	}

	return nil
}
