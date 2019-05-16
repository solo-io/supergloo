package smi

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("createRoutingConfig", func() {
	It("creates traffic splits", func() {
		ns := "default"
		rules := inputs.AdvancedBookInfoRoutingRules(ns, nil)
		upstreams := inputs.BookInfoUpstreams(ns)
		services := inputs.BookInfoServices(ns)
		resourceErrs := make(reporter.ResourceErrors)
		rc := createRoutingConfig(rules, upstreams, services, resourceErrs)
		Expect(rc).To(Equal("hay"))
	})
})
