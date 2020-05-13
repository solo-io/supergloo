package settings_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/settings.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/settings"
	mock_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser/mocks"
)

var _ = Describe("Utils", func() {
	var (
		ctrl              *gomock.Controller
		mockArnParser     *mock_aws.MockArnParser
		settingsConverter settings.AwsSelector
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockArnParser = mock_aws.NewMockArnParser(ctrl)
		settingsConverter = settings.NewAwsSelector(mockArnParser)
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
		resourceSelectors := []*zephyr_settings_types.ResourceSelector{
			{
				MatcherType: &zephyr_settings_types.ResourceSelector_Matcher_{
					Matcher: &zephyr_settings_types.ResourceSelector_Matcher{
						Regions: []string{region1, region2},
						Tags:    map[string]string{"tag1": "value1"},
					},
				},
			},
			{
				MatcherType: &zephyr_settings_types.ResourceSelector_Matcher_{
					Matcher: &zephyr_settings_types.ResourceSelector_Matcher{
						Regions: []string{region2, region3},
						Tags:    map[string]string{"tag2": "value2"},
					},
				},
			},
			{
				MatcherType: &zephyr_settings_types.ResourceSelector_Arn{
					Arn: arnString1,
				},
			},
			{
				MatcherType: &zephyr_settings_types.ResourceSelector_Arn{
					Arn: arnString2,
				},
			},
			{
				MatcherType: &zephyr_settings_types.ResourceSelector_Arn{
					Arn: arnString3,
				},
			},
		}
		expectedResourceSelectorsByRegion := settings.AwsSelectorsByRegion{
			region1: {resourceSelectors[0], resourceSelectors[2], resourceSelectors[3]},
			region2: {resourceSelectors[0], resourceSelectors[1]},
			region3: {resourceSelectors[1], resourceSelectors[4]},
		}
		selectorsByRegion, err := settingsConverter.ResourceSelectorsByRegion(resourceSelectors)
		Expect(err).ToNot(HaveOccurred())
		Expect(selectorsByRegion).To(Equal(expectedResourceSelectorsByRegion))
	})
})
