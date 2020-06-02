package settings_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1"
	zephyr_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/aws/settings"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	mock_zephyr_settings_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/settings.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("SettingsClient", func() {
	var (
		ctrl                 *gomock.Controller
		ctx                  context.Context
		mockSettingsClient   *mock_zephyr_settings_clients.MockSettingsClient
		settingsHelperClient settings.SettingsHelperClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockSettingsClient = mock_zephyr_settings_clients.NewMockSettingsClient(ctrl)
		settingsHelperClient = settings.NewAwsSettingsHelperClient(mockSettingsClient)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectGetSettingsSpec = func(settingsSpec zephyr_settings_types.SettingsSpec) {
		settings := &v1alpha1.Settings{Spec: settingsSpec}
		mockSettingsClient.
			EXPECT().
			GetSettings(ctx, client.ObjectKey{Name: metadata.GlobalSettingsName, Namespace: container_runtime.GetWriteNamespace()}).
			Return(settings, nil)
	}

	It("should get AWS Settings for account ID", func() {
		accountSettings := &zephyr_settings_types.SettingsSpec_AwsAccount{
			AccountId:     "account-id",
			MeshDiscovery: &zephyr_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{},
			EksDiscovery:  &zephyr_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{},
		}
		settingsSpec := zephyr_settings_types.SettingsSpec{
			Aws: &zephyr_settings_types.SettingsSpec_Aws{
				Accounts: []*zephyr_settings_types.SettingsSpec_AwsAccount{
					accountSettings,
				},
			},
		}
		expectGetSettingsSpec(settingsSpec)
		accountSettings, err := settingsHelperClient.GetAWSSettingsForAccount(ctx, accountSettings.GetAccountId())
		Expect(err).ToNot(HaveOccurred())
		Expect(accountSettings).To(Equal(accountSettings))
	})

	It("should return nil if accountID not found", func() {
		accountSettings := &zephyr_settings_types.SettingsSpec_AwsAccount{
			AccountId:     "account-id",
			MeshDiscovery: &zephyr_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{},
			EksDiscovery:  &zephyr_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{},
		}
		settingsSpec := zephyr_settings_types.SettingsSpec{
			Aws: &zephyr_settings_types.SettingsSpec_Aws{
				Accounts: []*zephyr_settings_types.SettingsSpec_AwsAccount{
					accountSettings,
				},
			},
		}
		expectGetSettingsSpec(settingsSpec)
		accountSettings, err := settingsHelperClient.GetAWSSettingsForAccount(ctx, "missing accountID")
		Expect(err).ToNot(HaveOccurred())
		Expect(accountSettings).To(BeNil())
	})

	It("should return nil if disabled for account", func() {
		accountSettings := &zephyr_settings_types.SettingsSpec_AwsAccount{
			AccountId:     "account-id",
			MeshDiscovery: &zephyr_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{},
			EksDiscovery:  &zephyr_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{},
		}
		settingsSpec := zephyr_settings_types.SettingsSpec{
			Aws: &zephyr_settings_types.SettingsSpec_Aws{
				Disabled: true,
				Accounts: []*zephyr_settings_types.SettingsSpec_AwsAccount{
					accountSettings,
				},
			},
		}
		expectGetSettingsSpec(settingsSpec)
		accountSettings, err := settingsHelperClient.GetAWSSettingsForAccount(ctx, "missing accountID")
		Expect(err).ToNot(HaveOccurred())
		Expect(accountSettings).To(BeNil())
	})
})
