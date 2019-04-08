package appmesh

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	whclients "github.com/solo-io/supergloo/pkg/webhook/clients"
	"github.com/solo-io/supergloo/pkg/webhook/test"
)

var _ = Describe("config translator", func() {
	var (
		testResources *test.ResourcesForTest
	)
	BeforeEach(func() {
		clients.UseMemoryClients()
		testResources = test.GetTestResources(whclients.Codecs.UniversalDeserializer())
	})
	Context("get pods for mesh", func() {
		It("can filter valid mesh pods", func() {
			_, _, err := getPodsForMesh(testResources.AppMeshInjectEnabledNamespaceSelector, testResources.InjectedPodList)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("get upstreams for mesh", func() {

	})
})
