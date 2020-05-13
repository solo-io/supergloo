package settings

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/rotisserie/eris"
	zephyr_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/settings.zephyr.solo.io/v1alpha1/types"
	aws_utils "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser"
	"github.com/solo-io/skv2/pkg/utils"
)

var UnknownSelectorType = func(selector *zephyr_settings_types.ResourceSelector) error {
	return eris.Errorf("Unknown ResourceSelector type: %+v", selector)
}

type awsSelector struct {
	arnParser aws_utils.ArnParser
}

func NewAwsSelector(arnParser aws_utils.ArnParser) AwsSelector {
	return &awsSelector{arnParser: arnParser}
}

func (s *awsSelector) ResourceSelectorsByRegion(
	resourceSelectors []*zephyr_settings_types.ResourceSelector,
) (AwsSelectorsByRegion, error) {
	resourceSelectorsByRegion := make(AwsSelectorsByRegion)
	for _, resourceSelector := range resourceSelectors {
		switch resourceSelector.GetMatcherType().(type) {
		case *zephyr_settings_types.ResourceSelector_Matcher_:
			for _, region := range resourceSelector.GetMatcher().GetRegions() {
				selectors, ok := resourceSelectorsByRegion[region]
				if !ok {
					resourceSelectorsByRegion[region] = []*zephyr_settings_types.ResourceSelector{}
				}
				resourceSelectorsByRegion[region] = append(selectors, resourceSelector)
			}
		case *zephyr_settings_types.ResourceSelector_Arn:
			region, err := s.arnParser.ParseRegion(resourceSelector.GetArn())
			if err != nil {
				return nil, err
			}
			selectors, ok := resourceSelectorsByRegion[region]
			if !ok {
				resourceSelectorsByRegion[region] = []*zephyr_settings_types.ResourceSelector{}
			}
			resourceSelectorsByRegion[region] = append(selectors, resourceSelector)
		default:
			return nil, UnknownSelectorType(resourceSelector)
		}
	}
	return resourceSelectorsByRegion, nil
}

func (s *awsSelector) AppMeshMatchedBySelectors(
	appmeshRef *appmesh.MeshRef,
	appmeshTags []*appmesh.TagRef,
	selectors []*zephyr_settings_types.ResourceSelector,
) (bool, error) {
	for _, selector := range selectors {
		matched, err := s.appMeshMatchedBySelector(appmeshRef, appmeshTags, selector)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func (s *awsSelector) EKSMatchedBySelectors(
	eksCluster *eks.Cluster,
	selectors []*zephyr_settings_types.ResourceSelector,
) (bool, error) {
	for _, selector := range selectors {
		matched, err := s.eksMatchedBySelector(eksCluster, selector)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func (s *awsSelector) appMeshMatchedBySelector(
	appmeshRef *appmesh.MeshRef,
	appmeshTags []*appmesh.TagRef,
	selector *zephyr_settings_types.ResourceSelector,
) (bool, error) {
	switch selector.GetMatcherType().(type) {
	case *zephyr_settings_types.ResourceSelector_Matcher_:
		appMeshRegion, err := s.arnParser.ParseRegion(aws.StringValue(appmeshRef.Arn))
		if err != nil {
			return false, err
		}
		matcherApplies := utils.ContainsString(selector.GetMatcher().GetRegions(), appMeshRegion) &&
			appmeshContainsTags(selector.GetMatcher().GetTags(), appmeshTags)
		return matcherApplies, nil
	case *zephyr_settings_types.ResourceSelector_Arn:
		return aws.StringValue(appmeshRef.Arn) == selector.GetArn(), nil
	default:
		return false, UnknownSelectorType(selector)
	}
	return false, nil
}

func (s *awsSelector) eksMatchedBySelector(
	eksCluster *eks.Cluster,
	selector *zephyr_settings_types.ResourceSelector,
) (bool, error) {
	switch selector.GetMatcherType().(type) {
	case *zephyr_settings_types.ResourceSelector_Matcher_:
		eksRegion, err := s.arnParser.ParseRegion(aws.StringValue(eksCluster.Arn))
		if err != nil {
			return false, err
		}
		return utils.ContainsString(selector.GetMatcher().GetRegions(), eksRegion) &&
			eksContainsTags(selector.GetMatcher().GetTags(), eksCluster.Tags), nil
	case *zephyr_settings_types.ResourceSelector_Arn:
		return aws.StringValue(eksCluster.Arn) == selector.GetArn(), nil
	default:
		return false, UnknownSelectorType(selector)
	}
	return false, nil
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
