package schemes

import (
	linkerd_config "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	smi_config "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	skv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	networking_v1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	security_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// contains all the Schemes for registering the CRDs
// with which Service Mesh Hub interacts.
// we share one SchemeBuilder for all our Managers
// as there's no harm in registering all I/O types internally.
var SchemeBuilder = runtime.SchemeBuilder{
	// internal types
	smh_discovery.AddToScheme,
	smh_networking.AddToScheme,

	// external types
	security_v1beta1.AddToScheme,
	networking_v1alpha3.AddToScheme,
	linkerd_config.AddToScheme,
	smi_config.AddToScheme,

	// sk types
	skv1alpha1.AddToScheme,
}
