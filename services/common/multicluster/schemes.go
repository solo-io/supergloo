package multicluster

import (
	"context"

	linkerd_config "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	"github.com/rotisserie/eris"
	smi_config "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	manager2 "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	networking_v1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	security_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	k8s_runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// register the mesh projects v1alpha1 CRDs with manager runtime
var AddAllV1Alpha1ToScheme manager2.AsyncManagerStartOptionsFunc = func(_ context.Context, mgr manager.Manager) error {
	addToSchemes := []func(error2 *k8s_runtime.Scheme) error{
		zephyr_discovery.AddToScheme,
		zephyr_networking.AddToScheme,
		zephyr_security.AddToScheme,
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

var AddAllIstioToScheme manager2.AsyncManagerStartOptionsFunc = func(_ context.Context, mgr manager.Manager) error {
	addToSchemes := []func(scheme *k8s_runtime.Scheme) error{
		security_v1beta1.AddToScheme,
		networking_v1alpha3.AddToScheme,
	}

	var err error
	for _, addToScheme := range addToSchemes {
		err = addToScheme(mgr.GetScheme())
		if err != nil {
			return eris.Wrap(err, "failed to add istio CRDs to manager runtime scheme")
		}
	}
	return nil
}

var AddAllLinkerdToScheme manager2.AsyncManagerStartOptionsFunc = func(_ context.Context, mgr manager.Manager) error {
	addToSchemes := []func(scheme *k8s_runtime.Scheme) error{
		linkerd_config.AddToScheme,
		smi_config.AddToScheme,
	}

	var err error
	for _, addToScheme := range addToSchemes {
		err = addToScheme(mgr.GetScheme())
		if err != nil {
			return eris.Wrap(err, "failed to add istio CRDs to manager runtime scheme")
		}
	}
	return nil
}
