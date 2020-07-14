package selection

import (
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/eks"
	smh_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// Nil value denotes selection of all resources in that region.
type AwsSelectorsByRegion map[string][]*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector

type AwsSelector interface {
	ResourceSelectorsByRegion(
		resourceSelectors []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector,
	) (AwsSelectorsByRegion, error)

	AwsSelectorsForAllRegions() AwsSelectorsByRegion

	// Return true if discoverySettings is nil or empty
	IsDiscoverAll(discoverySettings *smh_settings_types.SettingsSpec_AwsAccount_DiscoverySelector) bool

	AppMeshMatchedBySelectors(
		appmeshRef *appmesh.MeshRef,
		appmeshTags []*appmesh.TagRef,
		selectors []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector,
	) (bool, error)

	EKSMatchedBySelectors(
		eksCluster *eks.Cluster,
		selectors []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector,
	) (bool, error)
}
