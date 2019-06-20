package multimesh_test

import (
	"sync"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/test/e2e/testfuncs"
)

var _ = Describe("all e2e", func() {
	wg := sync.WaitGroup{}
	wg.Add(3)
	It("runs", func() {
		lock := &sync.Mutex{}
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			testfuncs.RunIstioE2eTests(testfuncs.IstioE2eTestParams{
				Kube:                helpers.MustKubeClient(),
				PromNamespace:       promNamespace,
				SmiIstioAdapterFile: smiIstioAdapterFile,
				GlooNamespace:       glooNamespace,
				SuperglooNamespace:  superglooNamespace,
				MeshName:            "istio-istio-system",
				BasicNamespace:      basicNamespace,
				NamespaceWithInject: namespaceWithIstioInject,
				IstioNamespace:      istioNamespace,
				RootCtx:             rootCtx,
				SharedLock:          lock,
			})
		}()

		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			testfuncs.RunLinkerdE2eTests(testfuncs.LinkerdE2eTestParams{
				Kube:                helpers.MustKubeClient(),
				PromNamespace:       promNamespace,
				GlooNamespace:       glooNamespace,
				SuperglooNamespace:  superglooNamespace,
				MeshName:            "linkerd-linkerd",
				BasicNamespace:      basicNamespace,
				NamespaceWithInject: namespaceWithLinkerdInject,
				LinkerdNamespace:    linkerdNamespace,
				RootCtx:             rootCtx,
				SharedLock:          lock,
			})
		}()

		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			testfuncs.RunAppMeshE2eTests(testfuncs.AppMeshE2eTestParams{
				Kube:                helpers.MustKubeClient(),
				MeshName:            "appmesh",
				SuperglooNamespace:  superglooNamespace,
				BasicNamespace:      basicNamespace,
				NamespaceWithInject: namespaceWithAppmeshInject,
			})
		}()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-failed:
		}
	})
})
