package multicluster

import (
	"context"

	"github.com/rotisserie/eris"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	security_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	k8s_runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// register the mesh projects v1alpha1 CRDs with manager runtime
var AddAllV1Alpha1ToScheme mc_manager.AsyncManagerStartOptionsFunc = func(_ context.Context, mgr manager.Manager) error {
	addToSchemes := []func(error2 *k8s_runtime.Scheme) error{
		discovery_v1alpha1.AddToScheme,
		networking_v1alpha1.AddToScheme,
		security_v1alpha1.AddToScheme,
	}
	var err error
	for _, addToScheme := range addToSchemes {
		err = addToScheme(mgr.GetScheme())
		if err != nil {
			return eris.Wrap(err, "failed to add v1alpha1 CRDs to manager runtime scheme")
		}
	}
	return nil
}
