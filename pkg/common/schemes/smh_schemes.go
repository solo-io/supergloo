package schemes

import (
	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	linkerdconfig "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	smiaccess1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	smispecsv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	smisplitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	certificates "github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/v1alpha2"
	smhdiscovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	smhnetworking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	skv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// contains all the Schemes for registering the CRDs
// with which Service Mesh Hub interacts.
// we share one SchemeBuilder for all our Managers
// as there's no harm in registering all I/O types internally.
var SchemeBuilder = runtime.SchemeBuilder{
	// internal types
	smhdiscovery.AddToScheme,
	smhnetworking.AddToScheme,
	certificates.AddToScheme,

	// external types
	istiosecurityv1beta1.AddToScheme,
	istionetworkingv1alpha3.AddToScheme,
	linkerdconfig.AddToScheme,
	smisplitv1alpha2.AddToScheme,
	smispecsv1alpha3.AddToScheme,
	smiaccess1alpha2.AddToScheme,
	appmeshv1beta2.AddToScheme,

	// sk types
	skv1alpha1.AddToScheme,
}

func AddToScheme(s *runtime.Scheme) error {
	return SchemeBuilder.AddToScheme(s)
}
