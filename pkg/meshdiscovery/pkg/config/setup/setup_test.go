package setup

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"github.com/solo-io/supergloo/pkg/registration"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("Setup", func() {
	var (
		ctx    context.Context
		cancel func()
	)
	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.TODO())
	})
	AfterEach(func() {
		cancel()
	})
	It("runs the mesh discovery registration event loop", func() {

		cs, err := clientset.ClientsetFromContext(ctx)
		Expect(err).NotTo(HaveOccurred())
		errHandler := func(err error) {
			defer GinkgoRecover()
			Expect(err).NotTo(HaveOccurred())
		}

		loop := NewDiscoveryConfigLoopStarters(cs)

		err = loop.Run(ctx, registration.EnabledConfigLoops{})
		Expect(err).NotTo(HaveOccurred())
	})
})
