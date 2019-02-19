package plugins_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/translator/istio/plugins"
)

var _ = Describe("Mtls", func() {
	Context("when mtls is enabled", func() {
		It("sets ISTIO_MUTUAL on destination rules", func() {
			p := NewMltsPlugin()
			params := Params{
				Ctx: context.TODO(),
			}
			in := v1.EncryptionRuleSpec{
				MtlsEnabled: true,
			}
			out := &v1alpha3.DestinationRule{}
			err := p.ProcessDestinationRule(params, in, out)
			Expect(err).NotTo(HaveOccurred())
			Expect(out.TrafficPolicy).NotTo(BeNil())
			Expect(out.TrafficPolicy.Tls).NotTo(BeNil())
			Expect(out.TrafficPolicy.Tls.Mode).To(Equal(v1alpha3.TLSSettings_ISTIO_MUTUAL))
		})
	})
})
