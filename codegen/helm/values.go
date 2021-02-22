package helm

import (
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
)

// The schema for our Helm chart values. Struct members must be public for visibility to skv2 Helm generator.
type ChartValues struct {
	GlooMeshOperatorArgs GlooMeshOperatorArgs `json:"glooMeshOperatorArgs"`
	//Settings                   SettingsValues       `json:"settings"`
	DisallowIntersectingConfig bool `json:"disallowIntersectingConfig"`
	WatchOutputTypes           bool `json:"watchOutputTypes"`
}

type GlooMeshOperatorArgs struct {
	SettingsRef SettingsRef `json:"settingsRef"`
}

type SettingsRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

//// we must use a custom Settings type here in order to ensure protos are marshalled to json properly
//type SettingsValues settingsv1alpha2.SettingsSpec
//
//func (v SettingsValues) MarshalJSON() ([]byte, error) {
//	settings := settingsv1alpha2.SettingsSpec(v)
//	js, err := (&jsonpb.Marshaler{EmitDefaults: true}).MarshalToString(&settings)
//	return []byte(js), err
//}

// The default chart values
func DefaultValues() ChartValues {
	return ChartValues{
		GlooMeshOperatorArgs: GlooMeshOperatorArgs{
			SettingsRef: SettingsRef{
				Name:      defaults.DefaultSettingsName,
				Namespace: defaults.DefaultPodNamespace,
			},
		},
		//Settings: SettingsValues{
		//	Mtls: &v1alpha2.TrafficPolicy_MTLS{
		//		Istio: &v1alpha2.TrafficPolicy_MTLS_Istio{
		//			TlsMode: v1alpha2.TrafficPolicy_MTLS_Istio_ISTIO_MUTUAL,
		//		},
		//	},
		//	// needed to ensure that generated yaml uses "{}" for empty message instead of "null", which causes a schema validation error
		//	Discovery: &settingsv1alpha2.DiscoverySettings{
		//		Istio: &settingsv1alpha2.DiscoverySettings_Istio{},
		//	},
		//	Relay: &settingsv1alpha2.RelaySettings{
		//		Enabled: false,
		//		Server:  &settingsv1alpha2.GrpcServer{},
		//	},
		//},
		DisallowIntersectingConfig: false,
		WatchOutputTypes:           true,
	}
}
