package istio_test

import (
	"sync"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/test/e2e/testfuncs"
)

const superglooNamespace = "supergloo-system"

var _ = Describe("istio e2e", func() {
	It("manages istio", func() {
		testfuncs.RunIstioE2eTests(testfuncs.IstioE2eTestParams{
			Kube:                helpers.MustKubeClient(),
			PromNamespace:       promNamespace,
			SmiIstioAdapterFile: smiIstioAdapterFile,
			GlooNamespace:       glooNamespace,
			SuperglooNamespace:  superglooNamespace,
			MeshName:            "istio-istio-system",
			BasicNamespace:      basicNamespace,
			NamespaceWithInject: namespaceWithInject,
			IstioNamespace:      istioNamespace,
			RootCtx:             rootCtx,
			SharedLock:          &sync.Mutex{},
		})
	})
})
