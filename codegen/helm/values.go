package helm

import "github.com/solo-io/service-mesh-hub/pkg/common/defaults"

// The schema for our Helm chart values. Struct members must be public for visibility to skv2 Helm generator.
type ChartValues struct {
	SmhOperatorArgs SmhOperatorArgs `json:"smhOperatorArgs"`
	Settings        Settings        `json:"settings"`
}

// default Settings values
type Settings struct {
	DefaultMtls bool `json:"defaultMtls"`
}

type SmhOperatorArgs struct {
	SettingsRef SettingsRef `json:"settingsRef"`
}

type SettingsRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
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
		Settings: Settings{
			DefaultMtls: true,
		},
	}
}
