package istio_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/supergloo/test/inputs"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/translator/istio"
)

var _ = FDescribe("Translator", func() {
	It("gahagafaga", func() {
		t := NewTranslator("hi", nil)
		meshConfig, resourceErrs, err := t.Translate(context.TODO(), &v1.ConfigSnapshot{
			Upstreams: map[string]gloov1.UpstreamList{"": inputs.BookInfoUpstrams()},
		})
		Expect(meshConfig).NotTo(HaveOccurred())
		Expect(resourceErrs).NotTo(HaveOccurred())
		Expect(err).NotTo(HaveOccurred())
	})
})
