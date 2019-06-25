package linkerd_test

import (
	"sync"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/test/e2e/testfuncs"
)

const superglooNamespace = "supergloo-system"

var _ = Describe("linkerd e2e", func() {
	It("manages linkerd", func() {
		testfuncs.RunLinkerdE2eTests(testfuncs.LinkerdE2eTestParams{
			Kube:                helpers.MustKubeClient(),
			PromNamespace:       promNamespace,
			GlooNamespace:       glooNamespace,
			SuperglooNamespace:  superglooNamespace,
			MeshName:            "linkerd-linkerd",
			BasicNamespace:      basicNamespace,
			NamespaceWithInject: namespaceWithInject,
			LinkerdNamespace:    linkerdNamespace,
			RootCtx:             rootCtx,
			SharedLock:          &sync.Mutex{},
		})
	})
})
