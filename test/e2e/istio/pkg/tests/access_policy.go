package tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/test/utils"
	skv2core "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// run tests for AccessPolicy CRD functionality
func AccessPolicyTest() {
	var (
		err      error
		manifest utils.Manifest
	)

	AfterEach(func() {
		manifest.Cleanup(BookinfoNamespace)
	})

	It("controls global access policy enforcement", func() {
		manifest, err = utils.NewManifest("access_policy_test_manifest.yaml")
		Expect(err).ToNot(HaveOccurred())

		By("restricting connectivity when global access policy enforcement is enabled", func() {
			VirtualMesh.Spec.GlobalAccessPolicy = networkingv1.VirtualMeshSpec_ENABLED
			VirtualMeshManifest.CreateOrTruncate()
			err := VirtualMeshManifest.AppendResources(VirtualMesh)
			Expect(err).NotTo(HaveOccurred())
			err = VirtualMeshManifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			Eventually(CurlReviews, "1m", "1s").Should(ContainSubstring("403 Forbidden"))
		})

		By("restoring connectivity to the reviews service when AccessPolicy is created", func() {
			accessPolicy := &networkingv1.AccessPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AccessPolicy",
					APIVersion: networkingv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "allow-reviews",
					Namespace: BookinfoNamespace,
				},
				Spec: networkingv1.AccessPolicySpec{
					SourceSelector: []*commonv1.IdentitySelector{
						{
							KubeServiceAccountRefs: &commonv1.IdentitySelector_KubeServiceAccountRefs{
								ServiceAccounts: []*skv2core.ClusterObjectRef{
									{
										Name:        "bookinfo-productpage",
										Namespace:   BookinfoNamespace,
										ClusterName: MgmtClusterName,
									},
								},
							},
						},
					},
					DestinationSelector: []*commonv1.DestinationSelector{
						{
							KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
								Services: []*skv2core.ClusterObjectRef{
									{
										Name:        "reviews",
										Namespace:   BookinfoNamespace,
										ClusterName: MgmtClusterName,
									},
								},
							},
						},
					},
				},
			}
			err := manifest.AppendResources(accessPolicy)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			Eventually(CurlReviews, "1m", "1s").Should(ContainSubstring("200 OK"))
		})

		By("restoring connectivity to all services when global access policy enforcement is disabled", func() {
			VirtualMesh.Spec.GlobalAccessPolicy = networkingv1.VirtualMeshSpec_DISABLED
			VirtualMeshManifest.CreateOrTruncate()
			err := VirtualMeshManifest.AppendResources(VirtualMesh)
			Expect(err).NotTo(HaveOccurred())
			err = VirtualMeshManifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			err = manifest.KubeDelete(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			Eventually(CurlRatings, "1m", "1s").Should(ContainSubstring("200 OK"))
		})
	})
}
