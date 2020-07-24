package headermanipulation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/headermanipulation"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("HeaderManipulationDecorator", func() {
	var (
		headerManipulationDecorator trafficpolicy.VirtualServiceDecorator
	)

	BeforeEach(func() {
		headerManipulationDecorator = headermanipulation.NewHeaderManipulationDecorator()
	})

	It("should set headers", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				HeaderManipulation: &v1alpha2.TrafficPolicySpec_HeaderManipulation{
					AppendRequestHeaders:  map[string]string{"a": "b"},
					RemoveRequestHeaders:  []string{"3", "4"},
					AppendResponseHeaders: map[string]string{"foo": "bar"},
					RemoveResponseHeaders: []string{"1", "2"},
				},
			},
		}
		output := &v1alpha3.HTTPRoute{}
		expectedHeaderManipulation := &v1alpha3.Headers{
			Request: &v1alpha3.Headers_HeaderOperations{
				Add:    map[string]string{"a": "b"},
				Remove: []string{"3", "4"},
			},
			Response: &v1alpha3.Headers_HeaderOperations{
				Add:    map[string]string{"foo": "bar"},
				Remove: []string{"1", "2"},
			},
		}
		err := headerManipulationDecorator.ApplyToVirtualService(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.Headers).To(Equal(expectedHeaderManipulation))
	})

	It("should not set headers if error during field registration", func() {
		testErr := eris.New("registration error")
		registerField := func(fieldPtr, val interface{}) error {
			return testErr
		}
		appliedPolicy := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				HeaderManipulation: &v1alpha2.TrafficPolicySpec_HeaderManipulation{},
			},
		}
		output := &v1alpha3.HTTPRoute{}
		err := headerManipulationDecorator.ApplyToVirtualService(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(output.Fault).To(BeNil())
	})
})
