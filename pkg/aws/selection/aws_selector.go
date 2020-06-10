package selection

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/rotisserie/eris"
	smh_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	aws_utils "github.com/solo-io/service-mesh-hub/pkg/aws/parser"
	"github.com/solo-io/skv2/pkg/utils"
)

var UnknownSelectorType = func(selector *smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector) error {
	return eris.Errorf("Unknown SettingsSpec_AwsAccount_ResourceSelector type: %+v", selector)
}

type awsSelector struct {
	arnParser aws_utils.ArnParser
}

func NewAwsSelector(arnParser aws_utils.ArnParser) AwsSelector {
	return &awsSelector{arnParser: arnParser}
}

func (a *awsSelector) ResourceSelectorsByRegion(
	resourceSelectors []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector,
) (AwsSelectorsByRegion, error) {
	resourceSelectorsByRegion := make(AwsSelectorsByRegion)
	for _, resourceSelector := range resourceSelectors {
		switch resourceSelector.GetMatcherType().(type) {
		case *smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_:
			for _, region := range resourceSelector.GetMatcher().GetRegions() {
				selectors, ok := resourceSelectorsByRegion[region]
				if !ok {
					resourceSelectorsByRegion[region] = []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{}
				}
				// If matches contains region but no specified tags, this indicates selection of all resources in that region.
				if resourceSelector.GetMatcher().GetTags() != nil {
					resourceSelectorsByRegion[region] = append(selectors, resourceSelector)
				}
			}
		case *smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Arn:
			region, err := a.arnParser.ParseRegion(resourceSelector.GetArn())
			if err != nil {
				return nil, err
			}
			selectors, ok := resourceSelectorsByRegion[region]
			if !ok {
				resourceSelectorsByRegion[region] = []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{}
			}
			resourceSelectorsByRegion[region] = append(selectors, resourceSelector)
		default:
			return nil, UnknownSelectorType(resourceSelector)
		}
	}
	return resourceSelectorsByRegion, nil
}

func (a *awsSelector) IsDiscoverAll(discoverySettings *smh_settings_types.SettingsSpec_AwsAccount_DiscoverySelector) bool {
	discoverAll := &smh_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{}
	return discoverySettings.Equal(discoverAll) || discoverySettings == nil
}

// Return AwsSelectorsByRegion that includes discovery for all resources in all regions.
func (a *awsSelector) AwsSelectorsForAllRegions() AwsSelectorsByRegion {
	awsSelectorsForAllRegions := make(AwsSelectorsByRegion)
	for region, _ := range endpoints.AwsPartition().Regions() {
		// Nil value denotes selection of all resources in that region.
		awsSelectorsForAllRegions[region] = nil
	}
	return awsSelectorsForAllRegions
}

// Return true if appmesh is matched by any selector, or if selectors is nil.
func (a *awsSelector) AppMeshMatchedBySelectors(
	appmeshRef *appmesh.MeshRef,
	appmeshTags []*appmesh.TagRef,
	selectors []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector,
) (bool, error) {
	if len(selectors) == 0 {
		return true, nil
	}
	for _, selector := range selectors {
		matched, err := a.appMeshMatchedBySelector(appmeshRef, appmeshTags, selector)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// Return true if EKS cluster is matched by any selector, or if selectors is nil.
func (a *awsSelector) EKSMatchedBySelectors(
	eksCluster *eks.Cluster,
	selectors []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector,
) (bool, error) {
	if len(selectors) == 0 {
		return true, nil
	}
	for _, selector := range selectors {
		matched, err := a.eksMatchedBySelector(eksCluster, selector)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func (a *awsSelector) appMeshMatchedBySelector(
	appmeshRef *appmesh.MeshRef,
	appmeshTags []*appmesh.TagRef,
	selector *smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector,
) (bool, error) {
	switch selector.GetMatcherType().(type) {
	case *smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_:
		appMeshRegion, err := a.arnParser.ParseRegion(aws.StringValue(appmeshRef.Arn))
		if err != nil {
			return false, err
		}
		matcherApplies := utils.ContainsString(selector.GetMatcher().GetRegions(), appMeshRegion) &&
			appmeshContainsTags(selector.GetMatcher().GetTags(), appmeshTags)
		return matcherApplies, nil
	case *smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Arn:
		return aws.StringValue(appmeshRef.Arn) == selector.GetArn(), nil
	default:
		return false, UnknownSelectorType(selector)
	}
}

func (a *awsSelector) eksMatchedBySelector(
	eksCluster *eks.Cluster,
	selector *smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector,
) (bool, error) {
	switch selector.GetMatcherType().(type) {
	case *smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_:
		eksRegion, err := a.arnParser.ParseRegion(aws.StringValue(eksCluster.Arn))
		if err != nil {
			return false, err
		}
		return utils.ContainsString(selector.GetMatcher().GetRegions(), eksRegion) &&
			eksContainsTags(selector.GetMatcher().GetTags(), eksCluster.Tags), nil
	case *smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Arn:
		return aws.StringValue(eksCluster.Arn) == selector.GetArn(), nil
	default:
		return false, UnknownSelectorType(selector)
	}
}

func appmeshContainsTags(selectorTags map[string]string, appMeshTags []*appmesh.TagRef) bool {
	appMeshTagsMap := make(map[string]string)
	for _, appMeshTag := range appMeshTags {
		appMeshTagsMap[aws.StringValue(appMeshTag.Key)] = aws.StringValue(appMeshTag.Value)
	}
	for key, value := range selectorTags {
		appMeshTagValue, ok := appMeshTagsMap[key]
		if !ok || appMeshTagValue != value {
			return false
		}
	}
	return true
}

func eksContainsTags(selectorTags map[string]string, eksTags map[string]*string) bool {
	for key, value := range selectorTags {
		eksTagValue, ok := eksTags[key]
		if !ok || aws.StringValue(eksTagValue) != value {
			return false
		}
	}
	return true
}
