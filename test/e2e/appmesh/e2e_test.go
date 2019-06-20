package appmesh_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/test/e2e/testfuncs"
)

var _ = Describe("E2e", func() {
	It("registers and tests appmesh", func() {
		testfuncs.RunAppMeshE2eTests(testfuncs.AppMeshE2eTestParams{
			Kube:                helpers.MustKubeClient(),
			MeshName:            "appmesh",
			SuperglooNamespace:  superglooNamespace,
			BasicNamespace:      basicNamespace,
			NamespaceWithInject: namespaceWithInject,
		})
	})
})
