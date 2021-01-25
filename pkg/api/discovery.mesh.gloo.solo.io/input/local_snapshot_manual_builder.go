// Code generated by skv2. DO NOT EDIT.

/*
	Utility for manually building input snapshots. Used primarily in tests.
*/
package input

import (
	settings_mesh_gloo_solo_io_v1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
	settings_mesh_gloo_solo_io_v1alpha2_sets "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2/sets"
)

type InputSettingsSnapshotManualBuilder struct {
	name string

	settingsMeshGlooSoloIov1Alpha2Settings settings_mesh_gloo_solo_io_v1alpha2_sets.SettingsSet
}

func NewInputSettingsSnapshotManualBuilder(name string) *InputSettingsSnapshotManualBuilder {
	return &InputSettingsSnapshotManualBuilder{
		name: name,

		settingsMeshGlooSoloIov1Alpha2Settings: settings_mesh_gloo_solo_io_v1alpha2_sets.NewSettingsSet(),
	}
}

func (i *InputSettingsSnapshotManualBuilder) Build() SettingsSnapshot {
	return NewSettingsSnapshot(
		i.name,

		i.settingsMeshGlooSoloIov1Alpha2Settings,
	)
}
func (i *InputSettingsSnapshotManualBuilder) AddSettingsMeshGlooSoloIov1Alpha2Settings(settingsMeshGlooSoloIov1Alpha2Settings []*settings_mesh_gloo_solo_io_v1alpha2.Settings) *InputSettingsSnapshotManualBuilder {
	i.settingsMeshGlooSoloIov1Alpha2Settings.Insert(settingsMeshGlooSoloIov1Alpha2Settings...)
	return i
}
