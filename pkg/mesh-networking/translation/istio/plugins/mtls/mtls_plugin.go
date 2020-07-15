package mtls

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/destinationrule"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
)

const (
	pluginName = "mtls"
)

// handles setting mTLS on a DestinationRule
type mtlsPlugin struct{
	complicatedLogicUnit interface{}
}

var _ destinationrule.Plugin = &mtlsPlugin{}

func NewMtlsPlugin() *mtlsPlugin {
	return &mtlsPlugin{}
}

func (m *mtlsPlugin) PluginName() string {
	return pluginName
}

func (m *mtlsPlugin) Process(meshService *v1alpha1.MeshService, output *v1alpha3.DestinationRule) {
	// TODO(ilackarms): currently we set all DRs to mTLS
	// in the future we'll want to make this configurable
	// https://github.com/solo-io/service-mesh-hub/issues/790
	if output.Spec.TrafficPolicy == nil {
		output.Spec.TrafficPolicy = &istiov1alpha3spec.TrafficPolicy{}
	}
	output.Spec.TrafficPolicy.Tls = &istiov1alpha3spec.ClientTLSSettings{
		Mode: istiov1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
	}
}
