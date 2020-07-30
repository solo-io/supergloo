package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = FDescribe("SMH e2e", func() {
	var (
	//accessPolicyClient v1alpha2.AccessPolicyClient
	)

	BeforeEach(func() {
		var err error
		//accessPolicyClient = v1alpha2.NewAccessPolicyClient(dynamicClient)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		//err := testutils.Kubectl("delete", "-f", policyManifest)
		//Expect(err).NotTo(HaveOccurred())
	})

	It("should enforce access policy when enabled", func() {
		By("creating a VirtualMesh with access policy enforcement enabled", func() {
			virtualMesh := &v1alpha2.VirtualMesh{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VirtualMesh",
					APIVersion: v1alpha2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "enforcement-enabled",
					Namespace: appNamespace,
				},
				Spec: v1alpha2.VirtualMeshSpec{
					Meshes: []*v1.ObjectRef{
						{
							Name:      "istiod-istio-system-master-cluster",
							Namespace: "service-mesh-hub",
						},
					},
					EnforceAccessControl: v1alpha2.VirtualMeshSpec_ENABLED,
				},
			}
			err := utils.WriteTestManifest(policyManifest, []metav1.Object{
				virtualMesh,
			})
			Expect(err).NotTo(HaveOccurred())
			err = testutils.Kubectl("apply", "-n="+appNamespace, "-f="+policyManifest)
			Expect(err).NotTo(HaveOccurred())
			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring("403 Forbidden"))
		})

		By("restore connectivity by deleting VirtualMesh", func() {
			err := testutils.Kubectl("delete", "-f", policyManifest)
			Expect(err).NotTo(HaveOccurred())

			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring("200 OK"))
		})
	})
})
