package settings

import (
	"context"

	zephyr_settings "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1"
	zephyr_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type awsSettingsHelperClient struct {
	settingsClient zephyr_settings.SettingsClient
}

func NewAwsSettingsHelperClient(settingsClient zephyr_settings.SettingsClient) SettingsHelperClient {
	return &awsSettingsHelperClient{settingsClient: settingsClient}
}

func (a *awsSettingsHelperClient) getSettingsSpec(ctx context.Context) (*zephyr_settings_types.SettingsSpec, error) {
	settings, err := a.settingsClient.GetSettings(
		ctx,
		client.ObjectKey{Name: metadata.GlobalSettingsName, Namespace: env.GetWriteNamespace()},
	)
	if errors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		return &settings.Spec, nil
	}
}

func (a *awsSettingsHelperClient) GetAWSSettingsForAccount(
	ctx context.Context,
	accountId string,
) (*zephyr_settings_types.SettingsSpec_AwsAccount, error) {
	settingsSpec, err := a.getSettingsSpec(ctx)
	if err != nil {
		return nil, err
	}
	if settingsSpec.GetAws().GetDisabled() {
		return nil, nil
	}
	for _, awsAccountConfig := range settingsSpec.GetAws().GetAccounts() {
		if awsAccountConfig.GetAccountId() == accountId {
			return awsAccountConfig, nil
		}
	}
	return nil, nil
}
