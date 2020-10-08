package helm

// The schema for our Helm chart values. Struct members must be public for visibility to skv2 Helm generator.
type ChartValues struct {
	Settings Settings `json:"settings"`
}

// default Settings values
type Settings struct {
	DefaultMtls bool `json:"defaultMtls"`
}

// The default chart values
func defaultValues() ChartValues {
	return ChartValues{
		Settings: Settings{
			DefaultMtls: true,
		},
	}
}
