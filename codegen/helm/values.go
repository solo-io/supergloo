package helm

import (
	"github.com/golang/protobuf/jsonpb"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
)

// The schema for our Helm chart values. Struct members must be public for visibility to skv2 Helm generator.
type ChartValues struct {
	GlooMeshOperatorArgs       GlooMeshOperatorArgs `json:"glooMeshOperatorArgs"`
	Settings                   SettingsValues       `json:"settings"`
	DisallowIntersectingConfig bool                 `json:"disallowIntersectingConfig"`
	WatchOutputTypes           bool                 `json:"watchOutputTypes"`
}

type GlooMeshOperatorArgs struct {
	SettingsRef SettingsRef `json:"settingsRef"`
}

type SettingsRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// we must use a custom Settings type here in order to ensure protos are marshalled to json properly
type SettingsValues settingsv1.SettingsSpec

func (v SettingsValues) MarshalJSON() ([]byte, error) {
	settings := settingsv1.SettingsSpec(v)
	js, err := (&jsonpb.Marshaler{EmitDefaults: true}).MarshalToString(&settings)
	return []byte(js), err
}

// The default chart values
func DefaultValues() ChartValues {
	return ChartValues{
		GlooMeshOperatorArgs: GlooMeshOperatorArgs{
			SettingsRef: SettingsRef{
				Name:      defaults.DefaultSettingsName,
				Namespace: defaults.DefaultPodNamespace,
			},
		},
		Settings: SettingsValues{
			Mtls: &networkingv1.TrafficPolicySpec_Policy_MTLS{
				Istio: &networkingv1.TrafficPolicySpec_Policy_MTLS_Istio{
					TlsMode: networkingv1.TrafficPolicySpec_Policy_MTLS_Istio_ISTIO_MUTUAL,
				},
			},
			// needed to ensure that generated yaml uses "{}" for empty message instead of "null", which causes a schema validation error
			Discovery: &settingsv1.DiscoverySettings{
				Istio: &settingsv1.DiscoverySettings_Istio{},
			},
			Relay: &settingsv1.RelaySettings{
				Enabled: false,
				Server:  &settingsv1.GrpcServer{},
			},
		},
		DisallowIntersectingConfig: false,
		WatchOutputTypes:           true,
	}
}
