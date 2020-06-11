package prompts_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_interactive "github.com/solo-io/service-mesh-hub/cli/pkg/common/interactive/mocks"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/prompts"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	"k8s.io/apimachinery/pkg/labels"
)

var _ = Describe("Interactive", func() {
	var (
		ctrl                  *gomock.Controller
		mockInteractivePrompt *mock_interactive.MockInteractivePrompt
		message               string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockInteractivePrompt = mock_interactive.NewMockInteractivePrompt(ctrl)
		message = "message"
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should select service selectors", func() {
		meshServiceNames := []string{"service1", "service2", "service3"}
		meshServiceNamesToRef := map[string]*smh_core_types.ResourceRef{
			meshServiceNames[0]: {
				Name: meshServiceNames[0],
			},
			meshServiceNames[1]: {
				Name: meshServiceNames[1],
			},
			meshServiceNames[2]: {
				Name: meshServiceNames[2],
			},
		}
		selectedNames := []string{meshServiceNames[0], meshServiceNames[2]}
		mockInteractivePrompt.
			EXPECT().
			SelectMultipleValues(message, meshServiceNames).
			Return(selectedNames, nil)
		expectedServiceSelector := &smh_core_types.ServiceSelector{
			ServiceSelectorType: &smh_core_types.ServiceSelector_ServiceRefs_{
				ServiceRefs: &smh_core_types.ServiceSelector_ServiceRefs{
					Services: []*smh_core_types.ResourceRef{
						meshServiceNamesToRef[meshServiceNames[0]],
						meshServiceNamesToRef[meshServiceNames[2]],
					},
				},
			},
		}
		selector, err := prompts.SelectServiceSelector(message, meshServiceNames, meshServiceNamesToRef, mockInteractivePrompt)
		Expect(err).ToNot(HaveOccurred())
		Expect(selector).To(Equal(expectedServiceSelector))
	})

	It("should prompt for labels", func() {
		labelSet := labels.Set(map[string]string{"k1": "v1", "k2": "v2"})
		mockInteractivePrompt.
			EXPECT().
			PromptValueWithValidator(message, "", gomock.Any()).
			Return("k1=v1, k2=v2", nil)
		selector, err := prompts.PromptLabels(message, mockInteractivePrompt)
		Expect(err).ToNot(HaveOccurred())
		Expect(selector).To(Equal(labelSet))
	})

	It("should prompt for comma delimited values", func() {
		values := []string{"v1", "v2", "v3"}
		mockInteractivePrompt.
			EXPECT().
			PromptValueWithValidator(message, "", gomock.Any()).
			Return("v1, v2 ,   v3", nil)
		inputValues, err := prompts.PromptCommaDelimitedValues(message, mockInteractivePrompt)
		Expect(err).ToNot(HaveOccurred())
		Expect(inputValues).To(Equal(values))
	})
})
