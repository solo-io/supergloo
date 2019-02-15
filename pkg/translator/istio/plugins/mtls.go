package plugins

import (
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type MtlsPlugin struct{}

func NewMltsPlugin() *MtlsPlugin {
	return &MtlsPlugin{}
}

func (p *MtlsPlugin) Init(params InitParams) error {
	return nil
}

func (p *MtlsPlugin) ProcessDestinationRule(params Params, in v1.EncryptionRuleSpec, out *v1alpha3.DestinationRule) error {
	if out.TrafficPolicy == nil {
		out.TrafficPolicy = &v1alpha3.TrafficPolicy{}
	}
	if in.MtlsEnabled {
		if out.TrafficPolicy.Tls != nil {
			contextutils.LoggerFrom(params.Ctx).Warnf("destination rule %v has already had its tls traffic policy set "+
				"by another plugin! overwriting", out.Metadata.Name)
		}
		// override root ca
		out.TrafficPolicy.Tls = &v1alpha3.TLSSettings{
			Mode: v1alpha3.TLSSettings_ISTIO_MUTUAL,
		}
	}
	return nil
}
