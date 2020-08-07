package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/test/utils"
	skv2core "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("AccessPolicy", func() {
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

		By("restricting connectivity when VirtualMesh with enforcement enabled is created", func() {
			virtualMesh := &networkingv1alpha2.VirtualMesh{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VirtualMesh",
					APIVersion: networkingv1alpha2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "enforcement-enabled",
					Namespace: BookinfoNamespace,
				},
				Spec: networkingv1alpha2.VirtualMeshSpec{
					Meshes: []*skv2core.ObjectRef{
						{
							Name:      "istiod-istio-system-master-cluster",
							Namespace: "service-mesh-hub",
						},
					},
					GlobalAccessPolicy: networkingv1alpha2.VirtualMeshSpec_ENABLED,
				},
			}
			err := manifest.AppendResources(virtualMesh)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())
			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring("403 Forbidden"))
		})

		By("restoring connectivity to the reviews service when AccessPolicy is created", func() {
			accessPolicy := &networkingv1alpha2.AccessPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AccessPolicy",
					APIVersion: networkingv1alpha2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "allow-reviews",
					Namespace: BookinfoNamespace,
				},
				Spec: networkingv1alpha2.AccessPolicySpec{
					SourceSelector: []*networkingv1alpha2.IdentitySelector{
						{
							KubeServiceAccountRefs: &networkingv1alpha2.IdentitySelector_KubeServiceAccountRefs{
								ServiceAccounts: []*skv2core.ClusterObjectRef{
									{
										Name:        "bookinfo-productpage",
										Namespace:   BookinfoNamespace,
										ClusterName: masterClusterName,
									},
								},
							},
						},
					},
					DestinationSelector: []*networkingv1alpha2.ServiceSelector{
						{
							KubeServiceRefs: &networkingv1alpha2.ServiceSelector_KubeServiceRefs{
								Services: []*skv2core.ClusterObjectRef{
									{
										Name:        "reviews",
										Namespace:   BookinfoNamespace,
										ClusterName: masterClusterName,
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

			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring("200 OK"))
		})

		By("restoring connectivity to all services when VirtualMesh is deleted", func() {
			err := manifest.KubeDelete(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			Eventually(curlRatings, "1m", "1s").Should(ContainSubstring("200 OK"))
		})
	})
})
