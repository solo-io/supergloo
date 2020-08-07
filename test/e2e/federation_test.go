package e2e_test

import (
	"context"
	"time"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

	"github.com/solo-io/service-mesh-hub/test/data"

	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Federation", func() {
	var (
		err      error
		manifest utils.Manifest
	)

	AfterEach(func() {
		manifest.Cleanup(BookinfoNamespace)
	})

	It("enables communication across clusters using global dns names", func() {
		manifest, err = utils.NewManifest("bookinfo-policies.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("initially curling remote reviews should fail to resolve", func() {
			Expect(curlRemoteReviews()).To(ContainSubstring("Could not resolve host"))
		})

		By("creating a VirtualMesh with federation enabled, cross-mesh communication should be enabled", func() {
			virtualMesh := data.SelfSignedVirtualMesh(
				"bookinfo-federation",
				BookinfoNamespace,
				[]*v1.ObjectRef{
					masterMesh,
					remoteMesh,
				})

			err = manifest.AppendResources(virtualMesh)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			assertVirtualMeshStatuses()

			// check we can hit the remote service
			// give 5 minutes because the workflow depends on restarting pods
			// which can take several minutes
			Eventually(curlRemoteReviews, "5m", "1s").Should(ContainSubstring(`"color": "black"`))
		})

		By("delete VirtualMesh should remove the federated service", func() {
			err = manifest.KubeDelete(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			Eventually(curlRemoteReviews, "5m", "1s").Should(ContainSubstring("Could not resolve host"))
		})
	})
})

func assertVirtualMeshStatuses() {
	ctx := context.Background()
	virtualMesh := v1alpha2.NewVirtualMeshClient(dynamicClient)

	EventuallyWithOffset(1, func() bool {
		list, err := virtualMesh.ListVirtualMesh(ctx, client.InNamespace(BookinfoNamespace))
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		ExpectWithOffset(1, list.Items).To(HaveLen(1))
		for _, policy := range list.Items {
			if policy.Status.ObservedGeneration == 0 {
				return false
			}
		}
		return true
	}, time.Second*60).Should(BeTrue())
}
