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

// run tests for AccessPolicy CRD functionality
func DiscoveryRegressionTest() {
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
			VirtualMesh.Spec.Federation.FlatNetwork = true
			VirtualMeshManifest.CreateOrTruncate()
			err := VirtualMeshManifest.AppendResources(VirtualMesh)
			Expect(err).NotTo(HaveOccurred())
			err = VirtualMeshManifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can discover all service endpoint when flat networking enabled", func() {

			By("enabling flat networking on the virtual mesh", func() {

				env := e2e.GetEnv()
				trafficTargetMgmtClient := env.Management.TrafficTargetClient
				trafficTargetRemoteClient := env.Remote.TrafficTargetClient

				Eventually(func() error {
					var multiErr *multierror.Error
					mgmtTrafficTargets, err := trafficTargetMgmtClient.ListTrafficTarget(ctx)
					Expect(err).NotTo(HaveOccurred())
					for _, v := range mgmtTrafficTargets.Items {
						if len(v.Spec.GetKubeService().GetEndpoints()) == 0 {
							multiErr = multierror.Append(multiErr, eris.Errorf(
								"%s has no endpoints",
								sets.TypedKey(&v)),
							)
						}
					}
					remoteTrafficTargets, err := trafficTargetRemoteClient.ListTrafficTarget(ctx)
					Expect(err).NotTo(HaveOccurred())
					for _, v := range remoteTrafficTargets.Items {
						if len(v.Spec.GetKubeService().GetEndpoints()) == 0 {
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
