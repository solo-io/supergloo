package settingsutils

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
)

type settingsKey struct{}

// ContextWithSettings returns a copy of parent context in which the
// value associated with settings key is the supplied settings.
func ContextWithSettings(ctx context.Context, settings *v1alpha2.Settings) context.Context {
	return context.WithValue(ctx, settingsKey{}, settings)
}

// FromContext returns the settings stored in context.
// Returns nil if no settings is set in context, or if the stored value is
// not of correct type.
func SettingsFromContext(ctx context.Context) *v1alpha2.Settings {
	if ctx != nil {
		if settings, ok := ctx.Value(settingsKey{}).(*v1alpha2.Settings); ok {
			return settings
		}
	}
	return nil
}
