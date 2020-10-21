package helm

import (
	"github.com/gogo/protobuf/jsonpb"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	settingsv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/settings.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
)

// The schema for our Helm chart values. Struct members must be public for visibility to skv2 Helm generator.
type ChartValues struct {
	SmhOperatorArgs SmhOperatorArgs `json:"smhOperatorArgs"`
	Settings        SettingsValues  `json:"settings"`
}

type SmhOperatorArgs struct {
	SettingsRef SettingsRef `json:"settingsRef"`
}

type SettingsRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// we must use a custom Settings type here in order to ensure protos are marshalled to json properly
type SettingsValues settingsv1alpha2.SettingsSpec

func (v SettingsValues) MarshalJSON() ([]byte, error) {
	settings := settingsv1alpha2.SettingsSpec(v)
	js, err := (&jsonpb.Marshaler{}).MarshalToString(&settings)
	return []byte(js), err
}

// The default chart values
func defaultValues() ChartValues {
	return ChartValues{
		SmhOperatorArgs: SmhOperatorArgs{
			SettingsRef: SettingsRef{
				Name:      defaults.DefaultSettingsName,
				Namespace: defaults.DefaultPodNamespace,
			},
		},
		Settings: SettingsValues{
			Mtls: &v1alpha2.TrafficPolicySpec_MTLS{
				Istio: &v1alpha2.TrafficPolicySpec_MTLS_Istio{
					TlsMode: v1alpha2.TrafficPolicySpec_MTLS_Istio_ISTIO_MUTUAL,
				},
			},
		},
	}
}
