package e2e_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/test/e2e"
	"github.com/solo-io/service-mesh-hub/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/service-mesh-hub/test/data"

	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Federation", func() {
	var (
		err                   error
		virtualMeshManifest   utils.Manifest
		trafficPolicyManifest utils.Manifest
		ctx                   = context.Background()
	)

	AfterEach(func() {
		virtualMeshManifest.Cleanup(BookinfoNamespace)
		trafficPolicyManifest.Cleanup(BookinfoNamespace)
		env := e2e.GetEnv()
		env.Management.EnableContainer(ctx, BookinfoNamespace, "reviews-v1")
		env.Management.EnableContainer(ctx, BookinfoNamespace, "reviews-v2")
		env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v1")
		env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v2")
	})

	It("enables communication across clusters using global dns names", func() {
		env := e2e.GetEnv()
		virtualMeshManifest, err = utils.NewManifest("federation-virtualmesh.yaml")
		Expect(err).NotTo(HaveOccurred())
		trafficPolicyManifest, err = utils.NewManifest("federation-trafficpolicies.yaml")
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

			err = virtualMeshManifest.AppendResources(virtualMesh)
			Expect(err).NotTo(HaveOccurred())
			err = virtualMeshManifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			assertVirtualMeshStatuses()

			// check we can hit the remote service
			// give 5 minutes because the workflow depends on restarting pods
			// which can take several minutes
			Eventually(curlRemoteReviews, "10m", "2s").Should(ContainSubstring("200 OK"))
		})

		By("with federation enabled, TrafficShifts can be used for subsets across meshes ", func() {
			// create cross cluster traffic shift
			trafficShiftReviewsV3 := data.RemoteTrafficShiftPolicy("bookinfo-policy", BookinfoNamespace, &v1.ClusterObjectRef{
				Name:        "reviews",
				Namespace:   BookinfoNamespace,
				ClusterName: masterClusterName,
			}, remoteClusterName, map[string]string{"version": "v3"}, 9080)

			err = trafficPolicyManifest.AppendResources(trafficShiftReviewsV3)
			Expect(err).NotTo(HaveOccurred())
			err = trafficPolicyManifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			utils.AssertTrafficPolicyStatuses(dynamicClient, BookinfoNamespace)

			// check we can eventually hit the v3 subset
			Eventually(curlReviews, "30s", "1s").Should(ContainSubstring(`"color": "red"`))
		})

		By("with federation enabled, FailoverService should fail over to remote cluster", func() {
			failoverServiceHostname := "reviews-failover.bookinfo.global"
			curlFailoverService := func() string {
				return curlFromProductpage(fmt.Sprintf("http://%s:9080/reviews/1", failoverServiceHostname))
			}

			trafficPolicy := &networkingv1alpha2.TrafficPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TrafficPolicy",
					APIVersion: networkingv1alpha2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reviews-outlier-detection",
					Namespace: BookinfoNamespace,
				},
				Spec: networkingv1alpha2.TrafficPolicySpec{
					DestinationSelector: []*networkingv1alpha2.ServiceSelector{
						{
							KubeServiceRefs: &networkingv1alpha2.ServiceSelector_KubeServiceRefs{
								Services: []*v1.ClusterObjectRef{
									{
										Name:        "reviews",
										Namespace:   BookinfoNamespace,
										ClusterName: masterClusterName,
									},
									{
										Name:        "reviews",
										Namespace:   BookinfoNamespace,
										ClusterName: remoteClusterName,
									},
								},
							},
						},
					},
					OutlierDetection: &networkingv1alpha2.TrafficPolicySpec_OutlierDetection{
						ConsecutiveErrors: 1,
					},
				},
			}
			failoverService := &networkingv1alpha2.FailoverService{
				TypeMeta: metav1.TypeMeta{
					Kind:       "FailoverService",
					APIVersion: networkingv1alpha2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reviews-failover",
					Namespace: BookinfoNamespace,
				},
				Spec: networkingv1alpha2.FailoverServiceSpec{
					Hostname: failoverServiceHostname,
					Port: &networkingv1alpha2.FailoverServiceSpec_Port{
						Number:   9080,
						Protocol: "http",
					},
					Meshes: []*v1.ObjectRef{
						masterMesh,
					},
					BackingServices: []*networkingv1alpha2.FailoverServiceSpec_BackingService{
						{
							BackingServiceType: &networkingv1alpha2.FailoverServiceSpec_BackingService_KubeService{
								KubeService: &v1.ClusterObjectRef{
									Name:        "reviews",
									Namespace:   BookinfoNamespace,
									ClusterName: masterClusterName,
								},
							},
						},
						{
							BackingServiceType: &networkingv1alpha2.FailoverServiceSpec_BackingService_KubeService{
								KubeService: &v1.ClusterObjectRef{
									Name:        "reviews",
									Namespace:   BookinfoNamespace,
									ClusterName: remoteClusterName,
								},
							},
						},
					},
				},
			}
			// Delete previous TrafficPolicy
			err := trafficPolicyManifest.KubeDelete(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())
			err = trafficPolicyManifest.CreateOrTruncate()
			Expect(err).NotTo(HaveOccurred())
			err = trafficPolicyManifest.AppendResources(trafficPolicy)
			Expect(err).NotTo(HaveOccurred())
			err = trafficPolicyManifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())
			// Wait for TrafficPolicy with outlier detection to be processed before creating FailoverService.
			utils.AssertTrafficPolicyStatuses(dynamicClient, BookinfoNamespace)

			err = trafficPolicyManifest.AppendResources(failoverService)
			Expect(err).NotTo(HaveOccurred())
			err = trafficPolicyManifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// Make it failover to remote cluster with reviews-v3, disable all master cluster reviews pods to prove
			// that request is being served by remote cluster
			env.Management.DisableContainer(ctx, BookinfoNamespace, "reviews-v1", "reviews")
			env.Management.DisableContainer(ctx, BookinfoNamespace, "reviews-v2", "reviews")
			env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v1")
			env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v2")

			// first check that we have a response to reduce flakiness
			Eventually(curlFailoverService, "1m", "1s").Should(ContainSubstring(`"color": "red"`))
			// now check that it is consistent 10 times in a row
			Eventually(func() bool {
				for i := 0; i < 5; i++ {
					if !strings.Contains(curlFailoverService(), `"color": "red"`) {
						return false
					}
					time.Sleep(2 * time.Second)
				}
				return true
			}, "5m", "2s").Should(BeTrue())
		})

		By("re-enable management-plane reviews deployments", func() {
			err := trafficPolicyManifest.KubeDelete(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			env.Management.EnableContainer(ctx, BookinfoNamespace, "reviews-v1")
			env.Management.EnableContainer(ctx, BookinfoNamespace, "reviews-v2")
			env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v1")
			env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v2")
			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring("200 OK"))
		})

		By("delete VirtualMesh should remove the federated service", func() {
			err = virtualMeshManifest.KubeDelete(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			Eventually(curlRemoteReviews, "5m", "1s").Should(ContainSubstring("Could not resolve host"))
		})
	})
})

func assertVirtualMeshStatuses() {
	ctx := context.Background()
	virtualMesh := networkingv1alpha2.NewVirtualMeshClient(dynamicClient)

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
