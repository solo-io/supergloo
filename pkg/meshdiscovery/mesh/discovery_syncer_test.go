package mesh_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh/istio"
)

var _ = Describe("discovery syncer", func() {
	var (
		meshClient v1.MeshClient
	)
	BeforeEach(func() {
		clients.UseMemoryClients()
		meshClient = clients.MustMeshClient()
	})
	It("can run without erroring", func() {
		ctx := context.TODO()
		snap := &v1.DiscoverySnapshot{}
		mockPlugin := istio.NewMockIstioMeshDiscovery()
		syncer := mesh.NewMeshDiscoverySyncer(meshClient, mockPlugin)
		err := syncer.Sync(ctx, snap)
		Expect(err).NotTo(HaveOccurred())
	})
})
