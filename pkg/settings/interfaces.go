package settings

import (
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/eks"
	zephyr_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/settings.zephyr.solo.io/v1alpha1/types"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type AwsSelectorsByRegion map[string][]*zephyr_settings_types.ResourceSelector

type AwsSelector interface {
	ResourceSelectorsByRegion(
		resourceSelectors []*zephyr_settings_types.ResourceSelector,
	) (AwsSelectorsByRegion, error)

	AppMeshMatchedBySelectors(
		appmeshRef *appmesh.MeshRef,
		appmeshTags []*appmesh.TagRef,
		selectors []*zephyr_settings_types.ResourceSelector,
	) (bool, error)

	EKSMatchedBySelectors(
		eksCluster *eks.Cluster,
		selectors []*zephyr_settings_types.ResourceSelector,
	) (bool, error)
}
