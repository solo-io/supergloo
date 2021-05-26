package helm

import (
	"github.com/golang/protobuf/jsonpb"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
)

// The schema for our Helm chart values. Struct members must be public for visibility to skv2 Helm generator.
type ChartValues struct {
	GlooMeshOperatorArgs       GlooMeshOperatorArgs `json:"glooMeshOperatorArgs"       desc:"Command line argument to Gloo Mesh deployments."`
	Settings                   SettingsValues       `json:"settings"                   desc:"Values for the Settings object. See the [Settings API doc](../../../../api/github.com.solo-io.gloo-mesh.api.settings.v1.settings) for details."`
	DisallowIntersectingConfig bool                 `json:"disallowIntersectingConfig" desc:"If true, Gloo Mesh will detect and report errors when outputting service mesh configuration that overlaps with existing config not managed by Gloo Mesh."`
	WatchOutputTypes           bool                 `json:"watchOutputTypes"           desc:"If true, Gloo Mesh will watch service mesh config types output by Gloo Mesh, and resync upon changes."`
	DefaultMetricsPort         uint32               `json:"defaultMetricsPort"         desc:"The port on which to serve internal Prometheus metrics for the Gloo Mesh application. Set to 0 to disable."`
	Verbose                    bool                 `json:"verbose"                    desc:"If true, enables verbose/debug logging."`
}

type GlooMeshOperatorArgs struct {
	SettingsRef SettingsRef `json:"settingsRef" desc:"Name/namespace of the Settings object."`
}

type SettingsRef struct {
	Name      string `json:"name"      desc:"Name of the Settings object."`
	Namespace string `json:"namespace" desc:"Namespace of the Settings object."`
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
			Dashboard: &settingsv1.DashboardSettings{},
		},
		DefaultMetricsPort:         defaults.MetricsPort,
		DisallowIntersectingConfig: false,
		WatchOutputTypes:           true,
		Verbose:                    false,
	}
}
