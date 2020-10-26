package settings

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	v1alpha22 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/settings.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/snapshotutils"
	"github.com/solo-io/skv2/pkg/ezkube"
)

var (
	MissingRequiredField = func(settingsRef ezkube.ResourceId, fieldName string) error {
		return eris.Errorf("Settings object %s.%s is missing required field %s", settingsRef.GetName(), settingsRef.GetNamespace(), fieldName)
	}
)

// Validate that the reference Settings object exists and that all required fields are specified.
func Validate(ctx context.Context, in input.Snapshot) (*v1alpha2.Settings, error) {
	settings, err := snapshotutils.GetSingletonSettings(ctx, in)
	if err != nil {
		return nil, err
	}

	settings.Status = v1alpha2.SettingsStatus{
		ObservedGeneration: settings.Generation,
		State:              v1alpha22.ApprovalState_ACCEPTED,
	}
	if errs := validateRequiredFields(settings); len(errs) > 0 {
		var errStrings []string
		for _, err := range errs {
			errStrings = append(errStrings, err.Error())
		}
		settings.Status.Errors = errStrings
		settings.Status.State = v1alpha22.ApprovalState_INVALID
		return nil, eris.Errorf("invalid Settings: %v", errStrings)
	}
	return settings, nil
}

// Validate that required fields are set.
func validateRequiredFields(settings *v1alpha2.Settings) []error {
	var errs []error

	return errs
}
