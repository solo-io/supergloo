package schemes

import (
	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	linkerdconfigv1alpha2 "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	smiaccess1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	smispecsv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	smisplitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	enterprisenetworkingv1beta1 "github.com/solo-io/gloo-mesh/pkg/api/networking.enterprise.mesh.gloo.solo.io/v1beta1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	observabilityv1 "github.com/solo-io/gloo-mesh/pkg/api/observability.enterprise.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	xdsv1beta1 "github.com/solo-io/gloo-mesh/pkg/api/xds.agent.enterprise.mesh.gloo.solo.io/v1beta1"
	skv2multiclusterv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// contains all the Schemes for registering the CRDs
// with which Gloo Mesh interacts.
// we share one SchemeBuilder for all our Managers
// as there's no harm in registering all I/O types internally.
var SchemeBuilder = runtime.SchemeBuilder{
	// internal types
	discoveryv1.AddToScheme,
	networkingv1.AddToScheme,
	certificatesv1.AddToScheme,
	xdsv1beta1.AddToScheme,
	enterprisenetworkingv1beta1.AddToScheme,
	observabilityv1.AddToScheme,
	settingsv1.AddToScheme,

	// external types
	istiosecurityv1beta1.AddToScheme,
	istionetworkingv1alpha3.AddToScheme,
	linkerdconfigv1alpha2.AddToScheme,
	smisplitv1alpha2.AddToScheme,
	smispecsv1alpha3.AddToScheme,
	smiaccess1alpha2.AddToScheme,
	appmeshv1beta2.AddToScheme,

	// sk types
	skv2multiclusterv1alpha1.AddToScheme,
}

func AddToScheme(s *runtime.Scheme) error {
	return SchemeBuilder.AddToScheme(s)
}
