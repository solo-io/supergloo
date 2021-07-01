package headermanipulation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/headermanipulation"
	"github.com/solo-io/go-utils/testutils"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("HeaderManipulationDecorator", func() {
	var (
		headerManipulationDecorator decorators.TrafficPolicyVirtualServiceDecorator
		output                      *v1alpha3.HTTPRoute
	)

	BeforeEach(func() {
		headerManipulationDecorator = headermanipulation.NewHeaderManipulationDecorator()
		output = &v1alpha3.HTTPRoute{}
	})

	It("should set headers", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					HeaderManipulation: &v1.HeaderManipulation{
						AppendRequestHeaders:  map[string]string{"a": "b"},
						RemoveRequestHeaders:  []string{"3", "4"},
						AppendResponseHeaders: map[string]string{"foo": "bar"},
						RemoveResponseHeaders: []string{"1", "2"},
					},
				},
			},
		}
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
		err := headerManipulationDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, nil, nil, output, registerField)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.Headers).To(Equal(expectedHeaderManipulation))
	})

	It("should not set headers if error during field registration", func() {
		testErr := eris.New("registration error")
		registerField := func(fieldPtr, val interface{}) error {
			return testErr
		}
		appliedPolicy := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					HeaderManipulation: &v1.HeaderManipulation{},
				},
			},
		}
		err := headerManipulationDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, nil, nil, output, registerField)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(output.Fault).To(BeNil())
	})
})
