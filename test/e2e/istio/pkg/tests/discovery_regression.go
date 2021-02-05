package tests

import (
	"context"

	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/test/e2e"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// run Enterprise Discovery regression tests
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

				Eventually(func() error {
					var multiErr *multierror.Error
					mgmtTrafficTargets, err := trafficTargetMgmtClient.ListTrafficTarget(ctx)
					Expect(err).NotTo(HaveOccurred())
					for _, v := range mgmtTrafficTargets.Items {

						for _, workloadRef := range v.Spec.GetWorkloads() {
							workload, err := env.Management.WorkloadClient.GetWorkload(
								ctx,
								client.ObjectKey{
									Namespace: workloadRef.GetNamespace(),
									Name:      workloadRef.GetName(),
								},
							)
							Expect(err).NotTo(HaveOccurred())

							if len(workload.Spec.GetEndpoints()) == 0 {
								multiErr = multierror.Append(multiErr, eris.Errorf(
									"%s has no endpoints",
									sets.TypedKey(workload)),
								)
							}
						}
					}
					return multiErr
				}, "5m", "2s").ShouldNot(HaveOccurred())
			})

		})
	})
}
