package fieldutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/fieldutils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

var _ = Describe("FieldOwnership", func() {
	It("registers an owner for a field, and errors when another owner with lower or equal priority attempts to register", func() {
		fieldRegistry := NewOwnershipRegistry()

		vs := &networkingv1alpha3.VirtualService{}

		// test with the real object on we use track ownership
		istioRoute := &networkingv1alpha3spec.HTTPRoute{
			CorsPolicy: &networkingv1alpha3spec.CorsPolicy{},
		}

		owner1 := []ezkube.ResourceId{&v1.ObjectRef{Name: "1"}}
		owner2 := []ezkube.ResourceId{&v1.ObjectRef{Name: "2"}}

		corsPolicyField := &istioRoute.CorsPolicy
		err := fieldRegistry.RegisterFieldOwnership(vs, corsPolicyField, owner1, &v1alpha2.TrafficPolicy{}, 1)
		Expect(err).NotTo(HaveOccurred())

		err = fieldRegistry.RegisterFieldOwnership(vs, corsPolicyField, owner2, &v1alpha2.TrafficPolicy{}, 0)
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(FieldConflictError{
			Field:    corsPolicyField,
			Owners:   owner1,
			Priority: 1,
		}))

	})
})
