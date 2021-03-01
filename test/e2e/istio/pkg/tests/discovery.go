package tests

import (
	"context"

	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/test/e2e"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

// run Enterprise Discovery regression tests
func DiscoveryTest() {
	var (
		ctx context.Context
	)

	Context("Flat Networking", func() {

		BeforeEach(func() {
			ctx = context.Background()

			VirtualMesh.Spec.Federation.FlatNetwork = true
			VirtualMeshManifest.CreateOrTruncate()
			err := VirtualMeshManifest.AppendResources(VirtualMesh)
			Expect(err).NotTo(HaveOccurred())
			err = VirtualMeshManifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			VirtualMesh.Spec.Federation.FlatNetwork = false
			VirtualMeshManifest.CreateOrTruncate()
			err := VirtualMeshManifest.AppendResources(VirtualMesh)
			Expect(err).NotTo(HaveOccurred())
			err = VirtualMeshManifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can discover all service endpoint when flat networking enabled", func() {

			By("enabling flat networking on the virtual mesh", func() {

				env := e2e.GetEnv()
				destinationMgmtClient := env.Management.DestinationClient

				Eventually(func() error {
					var multiErr *multierror.Error
					mgmtDestinations, err := destinationMgmtClient.ListDestination(ctx)
					Expect(err).NotTo(HaveOccurred())
					for _, v := range mgmtDestinations.Items {
						if len(v.Spec.GetKubeService().GetEndpointSubsets()) == 0 {
							multiErr = multierror.Append(multiErr, eris.Errorf(
								"%s has no endpoints",
								sets.TypedKey(&v)),
							)
						}
					}
					return multiErr
				}, "5m", "2s").ShouldNot(HaveOccurred())
			})

		})
	})
}
