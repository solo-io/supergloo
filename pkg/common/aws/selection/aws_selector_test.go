package selection_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	mock_aws "github.com/solo-io/service-mesh-hub/pkg/common/aws/parser/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/selection"
)

var _ = Describe("AWS Selector", func() {
	var (
		ctrl          *gomock.Controller
		mockArnParser *mock_aws.MockArnParser
		awsSelector   selection.AwsSelector
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockArnParser = mock_aws.NewMockArnParser(ctrl)
		awsSelector = selection.NewAwsSelector(mockArnParser)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should group AWS ResourceSelectors by region", func() {
		region1 := "us-east-1"
		region2 := "us-west-1"
		region3 := "eu-west-1"
		arnString1 := "arnString1"
		arnString2 := "arnString2"
		arnString3 := "arnString3"
		mockArnParser.EXPECT().ParseRegion(arnString1).Return(region1, nil)
		mockArnParser.EXPECT().ParseRegion(arnString2).Return(region1, nil)
		mockArnParser.EXPECT().ParseRegion(arnString3).Return(region3, nil)
		resourceSelectors := []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
			{
				MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_{
					Matcher: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher{
						Regions: []string{region1, region2},
						Tags:    map[string]string{"tag1": "value1"},
					},
				},
			},
			{
				MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_{
					Matcher: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher{
						Regions: []string{region2, region3},
						Tags:    map[string]string{"tag2": "value2"},
					},
				},
			},
			{
				MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Arn{
					Arn: arnString1,
				},
			},
			{
				MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Arn{
					Arn: arnString2,
				},
			},
			{
				MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Arn{
					Arn: arnString3,
				},
			},
		}
		expectedResourceSelectorsByRegion := selection.AwsSelectorsByRegion{
			region1: {resourceSelectors[0], resourceSelectors[2], resourceSelectors[3]},
			region2: {resourceSelectors[0], resourceSelectors[1]},
			region3: {resourceSelectors[1], resourceSelectors[4]},
		}
		selectorsByRegion, err := awsSelector.ResourceSelectorsByRegion(resourceSelectors)
		Expect(err).ToNot(HaveOccurred())
		Expect(selectorsByRegion).To(Equal(expectedResourceSelectorsByRegion))
	})

	It("should return true if DiscoverySettings exists but is empty", func() {
		isDiscoverAll := awsSelector.IsDiscoverAll(&smh_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{})
		Expect(isDiscoverAll).To(BeTrue())
	})

	It("should return true if DiscoverySettings does not exist", func() {
		settings := &smh_settings_types.SettingsSpec_AwsAccount{}
		isDiscoverAll := awsSelector.IsDiscoverAll(settings.GetEksDiscovery())
		Expect(isDiscoverAll).To(BeTrue())
	})

	It("should return AwsSelectorsByRegion representing all regions", func() {
		awsSelectorsForAllRegions := make(selection.AwsSelectorsByRegion)
		for region, _ := range endpoints.AwsPartition().Regions() {
			awsSelectorsForAllRegions[region] = nil
		}
		allRegions := awsSelector.AwsSelectorsForAllRegions()
		Expect(allRegions).To(Equal(awsSelectorsForAllRegions))
	})

	It("should return true if appmesh matches any selector by region and tags", func() {
		appmeshRef := &appmesh.MeshRef{
			Arn: aws.String(""),
		}
		region := "us-east-2"
		mockArnParser.EXPECT().ParseRegion(aws.StringValue(appmeshRef.Arn)).Return(region, nil)
		appmeshTags := []*appmesh.TagRef{
			{
				Key:   aws.String("k1"),
				Value: aws.String("v1"),
			},
			{
				Key:   aws.String("k2"),
				Value: aws.String("v2"),
			},
		}
		selector := &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
			MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_{
				Matcher: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher{
					Regions: []string{"region2", region},
					Tags: map[string]string{
						"k1": "v1",
					},
				},
			},
		}
		matchedBySelectors, err := awsSelector.AppMeshMatchedBySelectors(appmeshRef, appmeshTags, []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{selector})
		Expect(err).To(BeNil())
		Expect(matchedBySelectors).To(BeTrue())
	})

	It("should return true if appmesh does not match any selector", func() {
		appmeshRef := &appmesh.MeshRef{
			Arn: aws.String(""),
		}
		region := "us-east-2"
		mockArnParser.EXPECT().ParseRegion(aws.StringValue(appmeshRef.Arn)).Return(region, nil)
		appmeshTags := []*appmesh.TagRef{
			{
				Key:   aws.String("k1"),
				Value: aws.String("v1"),
			},
			{
				Key:   aws.String("k2"),
				Value: aws.String("v2"),
			},
		}
		selector := &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
			MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_{
				Matcher: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher{
					Regions: []string{"region2", region},
					Tags: map[string]string{
						"k1": "v1",
						"k3": "v3",
					},
				},
			},
		}
		matchedBySelectors, err := awsSelector.AppMeshMatchedBySelectors(appmeshRef, appmeshTags, []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{selector})
		Expect(err).To(BeNil())
		Expect(matchedBySelectors).To(BeFalse())
	})

	It("should return true for appmesh if selector is empty or nil", func() {
		appmeshRef := &appmesh.MeshRef{
			Arn: aws.String(""),
		}
		appmeshTags := []*appmesh.TagRef{
			{
				Key:   aws.String("k1"),
				Value: aws.String("v1"),
			},
		}
		matchedBySelectors, err := awsSelector.AppMeshMatchedBySelectors(appmeshRef, appmeshTags, []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{})
		Expect(err).To(BeNil())
		Expect(matchedBySelectors).To(BeTrue())
		matchedBySelectors, err = awsSelector.AppMeshMatchedBySelectors(appmeshRef, appmeshTags, nil)
		Expect(err).To(BeNil())
		Expect(matchedBySelectors).To(BeTrue())
	})

	It("should return true if appmesh matches any selector by ARN", func() {
		appmeshRef := &appmesh.MeshRef{
			Arn: aws.String("arn"),
		}
		selector := &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
			MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Arn{
				Arn: "arn",
			},
		}
		matchedBySelectors, err := awsSelector.AppMeshMatchedBySelectors(appmeshRef, nil, []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{selector})
		Expect(err).To(BeNil())
		Expect(matchedBySelectors).To(BeTrue())
	})

	It("should return true if EKSCluster matches any selector by region and tags", func() {
		region := "us-east-2"
		eksCluster := &eks.Cluster{
			Arn: aws.String(""),
			Tags: map[string]*string{
				"k1": aws.String("v1"),
			},
		}
		mockArnParser.EXPECT().ParseRegion(aws.StringValue(eksCluster.Arn)).Return(region, nil)
		selector := &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
			MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_{
				Matcher: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher{
					Regions: []string{"region2", region},
					Tags: map[string]string{
						"k1": "v1",
					},
				},
			},
		}
		matchedBySelectors, err := awsSelector.EKSMatchedBySelectors(eksCluster, []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{selector})
		Expect(err).To(BeNil())
		Expect(matchedBySelectors).To(BeTrue())
	})

	It("should return true for EKS if selector is empty or nil", func() {
		eksCluster := &eks.Cluster{
			Arn: aws.String(""),
			Tags: map[string]*string{
				"k1": aws.String("v1"),
			},
		}
		matchedBySelectors, err := awsSelector.EKSMatchedBySelectors(eksCluster, []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{})
		Expect(err).To(BeNil())
		Expect(matchedBySelectors).To(BeTrue())
		matchedBySelectors, err = awsSelector.EKSMatchedBySelectors(eksCluster, nil)
		Expect(err).To(BeNil())
		Expect(matchedBySelectors).To(BeTrue())
	})

	It("should return true if EKSCluster matches any selector by ARN", func() {
		eksCluster := &eks.Cluster{
			Arn: aws.String("arn"),
			Tags: map[string]*string{
				"k1": aws.String("v1"),
			},
		}
		selector := &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
			MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Arn{
				Arn: "arn",
			},
		}
		matchedBySelectors, err := awsSelector.EKSMatchedBySelectors(eksCluster, []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{selector})
		Expect(err).To(BeNil())
		Expect(matchedBySelectors).To(BeTrue())
	})
})
