package utils

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var syncerNameKey = mustKey("syncer_name")

const syncersActiveDesc = "The active number of config syncers"

var (
	mConfigSyncersActive = stats.Int64("supergloo.solo.io/config_syncers/syncers_active", syncersActiveDesc, "1")

	configSyncersView = &view.View{
		Name:        "supergloo.solo.io/config_syncers/syncers_active",
		Measure:     mConfigSyncersActive,
		Description: syncersActiveDesc,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{syncerNameKey},
	}
)

func init() {
	view.Register(configSyncersView)
}

func mustKey(name string) tag.Key {
	key, err := tag.NewKey(name)
	if err != nil {
		panic(err)
	}
	return key
}

func RecordActive(ctx context.Context, syncerName string, active bool) error {
	var value int64
	if active {
		value = 1
	}
	return stats.RecordWithTags(ctx, []tag.Mutator{tag.Upsert(syncerNameKey, syncerName)}, mConfigSyncersActive.M(value))
}
