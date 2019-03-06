package options_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	. "github.com/solo-io/supergloo/cli/pkg/options"
)

var _ = Describe("ResourceRefsValue", func() {
	It("errors if format is not NAMESPACE.NAME", func() {
		val := &ResourceRefsValue{}
		err := val.Set("badformat")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("refs must be specified in the format <NAMESPACE>.<NAME>"))
	})
	It("appends resource refs from format NAMESPACE.NAME", func() {
		val := &ResourceRefsValue{}
		err := val.Set("good.format")
		Expect(err).NotTo(HaveOccurred())
		err = val.Set("great.format")
		Expect(err).NotTo(HaveOccurred())
		Expect(*val).To(HaveLen(2))
		Expect((*val)[0]).To(Equal(core.ResourceRef{"format", "good"}))
		Expect((*val)[1]).To(Equal(core.ResourceRef{"format", "great"}))
	})
})

var _ = Describe("MapStringStringValue", func() {
	It("errors if format is not NAMESPACE.NAME", func() {
		val := &MapStringStringValue{}
		err := val.Set("badformat")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("map entries must be specified in the format KEY=VALUE"))
	})
	It("appends resource refs from format NAMESPACE.NAME", func() {
		val := &MapStringStringValue{}
		err := val.Set("good=format")
		Expect(err).NotTo(HaveOccurred())
		err = val.Set("great=format")
		Expect(err).NotTo(HaveOccurred())
		Expect(*val).To(HaveLen(2))
		Expect((*val)["good"]).To(Equal("format"))
		Expect((*val)["great"]).To(Equal("format"))
	})
})

var _ = Describe("RequestMatchersValue", func() {
	It("errors if format is not json request matcher string", func() {
		val := &RequestMatchersValue{}
		err := val.Set("badformat")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("request matcher must be specified as valid request matcher json"))
	})
	It("appends resource request matchers from json format", func() {
		val := &RequestMatchersValue{}
		err := val.Set(`{"path_prefix": "/", "methods": ["GET", "POST"], "header_matchers": {"foo":"bar"}}`)
		Expect(err).NotTo(HaveOccurred())
		err = val.Set(`{"path_exact": "/", "methods": ["POST"], "header_matchers": {"baz":"qux"}}`)
		Expect(err).NotTo(HaveOccurred())

		Expect(*val).To(HaveLen(2))
		Expect((*val)[0]).To(Equal(RequestMatcher{
			PathPrefix:    "/",
			Methods:       []string{"GET", "POST"},
			HeaderMatcher: map[string]string{"foo": "bar"},
		}))
		Expect((*val)[1]).To(Equal(RequestMatcher{
			PathExact:     "/",
			Methods:       []string{"POST"},
			HeaderMatcher: map[string]string{"baz": "qux"},
		}))
	})
})
