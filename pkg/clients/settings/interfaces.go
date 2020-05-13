package settings

import (
	"context"

	zephyr_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/settings.zephyr.solo.io/v1alpha1/types"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// Convenience wrapper around fetching the global Settings object.
type SettingsHelperClient interface {
	GetAWSSettingsForAccount(ctx context.Context, accountId string) (*zephyr_settings_types.AwsAccountSettings, error)
}
