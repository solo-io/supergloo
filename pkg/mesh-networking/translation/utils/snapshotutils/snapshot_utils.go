package snapshotutils

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/settings.smh.solo.io/v1alpha2"
)

// TODO extend skv2 snapshots with singleton object utilities
// Safely fetch the single Settings object from snapshot. Log and error if not singleton.
func GetSingletonSettings(ctx context.Context, in input.Snapshot) (*v1alpha2.Settings, error) {
	settings := in.Settings().List()
	n := len(settings)
	if n != 1 {
		err := eris.Errorf("Snapshot does not contain single Settings object, %d found.", n)
		contextutils.LoggerFrom(ctx).Errorf("%+v", err)
		return nil, err
	}
	return settings[0], nil
}
